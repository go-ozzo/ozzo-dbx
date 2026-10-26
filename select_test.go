// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbx

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectQuery(t *testing.T) {
	db := NewFromDB(nil, "mysql")

	// minimal select query
	q := db.Select().From("users").Build()
	expected := "SELECT * FROM `users`"
	assert.Equal(t, q.SQL(), expected, "t1")
	assert.Equal(t, len(q.Params()), 0, "t2")

	// a full select query
	q = db.Select("id", "name").
		AndSelect("age").
		Distinct(true).
		SelectOption("CALC").
		From("users").
		Where(NewExp("age>30")).
		AndWhere(NewExp("status=1")).
		OrWhere(NewExp("type=2")).
		InnerJoin("profile", NewExp("user.id=profile.id")).
		LeftJoin("team", nil).
		RightJoin("dept", nil).
		OrderBy("age DESC", "type").
		AndOrderBy("id").
		GroupBy("id").
		AndGroupBy("age").
		Having(NewExp("id>10")).
		AndHaving(NewExp("id<20")).
		OrHaving(NewExp("type=3")).
		Limit(10).
		Offset(20).
		Bind(Params{"id": 1}).
		AndBind(Params{"age": 30}).
		Build()

	expected = "SELECT DISTINCT CALC `id`, `name`, `age` FROM `users` INNER JOIN `profile` ON user.id=profile.id LEFT JOIN `team` RIGHT JOIN `dept` WHERE ((age>30) AND (status=1)) OR (type=2) GROUP BY `id`, `age` HAVING ((id>10) AND (id<20)) OR (type=3) ORDER BY `age` DESC, `type`, `id` LIMIT 10 OFFSET 20"
	assert.Equal(t, q.SQL(), expected, "t3")
	assert.Equal(t, len(q.Params()), 2, "t4")

	q3 := db.Select().AndBind(Params{"id": 1}).Build()
	assert.Equal(t, len(q3.Params()), 1)

	// union
	q1 := db.Select().From("users").Build()
	q2 := db.Select().From("posts").Build()
	q = db.Select().From("profiles").Union(q1).UnionAll(q2).Build()
	expected = "(SELECT * FROM `profiles`) UNION (SELECT * FROM `users`) UNION ALL (SELECT * FROM `posts`)"
	assert.Equal(t, q.SQL(), expected, "t5")
}

func TestSelectQuery_Lock(t *testing.T) {
	// a transaction builder without a database: Build never executes anything
	txOn := func(driver string) *Tx { return NewFromDB(nil, driver).Wrap(&sql.Tx{}) }
	mysql := txOn("mysql")
	pgsql := txOn("postgres")

	tests := []struct {
		name string
		q    *SelectQuery
		want string
	}{
		{"mysql no lock", mysql.Select().From("users").OrderBy("id").Limit(10), "SELECT * FROM `users` ORDER BY `id` LIMIT 10"},
		{"mysql ForUpdate", mysql.Select().From("users").Where(HashExp{"id": 1}).ForUpdate(), "SELECT * FROM `users` WHERE `id`={:p0} FOR UPDATE"},
		{"mysql ForShare", mysql.Select().From("users").ForShare(), "SELECT * FROM `users` FOR SHARE"},
		{"mysql ForUpdateSkipLocked after ORDER BY/LIMIT/OFFSET", mysql.Select().From("jobs").OrderBy("id").Limit(10).Offset(5).ForUpdateSkipLocked(),
			"SELECT * FROM `jobs` ORDER BY `id` LIMIT 10 OFFSET 5 FOR UPDATE SKIP LOCKED"},
		{"pgsql ForUpdate after ORDER BY/LIMIT", pgsql.Select().From("users").OrderBy("id").Limit(1).ForUpdate(), `SELECT * FROM "users" ORDER BY "id" LIMIT 1 FOR UPDATE`},
		{"pgsql ForShare", pgsql.Select().From("users").ForShare(), `SELECT * FROM "users" FOR SHARE`},
		{"pgsql ForUpdateSkipLocked", pgsql.Select().From("jobs").Limit(5).ForUpdateSkipLocked(), `SELECT * FROM "jobs" LIMIT 5 FOR UPDATE SKIP LOCKED`},
		{"standard ForUpdate", txOn("unknown").Select().From("users").ForUpdate(), `SELECT * FROM "users" FOR UPDATE`},
		// the last lock method called wins
		{"mysql last call wins", mysql.Select().From("users").ForShare().ForUpdate(), "SELECT * FROM `users` FOR UPDATE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.q.Build()
			assert.Nil(t, q.LastError)
			assert.Equal(t, tt.want, q.SQL())
		})
	}

	// SQLite has no row locks: inside a transaction the lock is silently left out
	for _, driver := range []string{"sqlite", "sqlite3"} {
		tx := txOn(driver)
		for _, sq := range []*SelectQuery{tx.Select().From("users").ForUpdate(), tx.Select().From("users").ForShare(), tx.Select().From("users").Limit(1).ForUpdateSkipLocked()} {
			q := sq.Build()
			assert.Nil(t, q.LastError, driver)
			assert.NotContains(t, q.SQL(), "FOR ", driver)
		}
	}

	// a lock outside a transaction is released at once: rejected when the executor is a *sql.DB,
	// on every database, SQLite included
	for _, driver := range []string{"mysql", "postgres", "pgx", "unknown", "sqlite", "sqlite3"} {
		db := NewFromDB(nil, driver)
		for method, sq := range map[string]*SelectQuery{
			"ForUpdate":           db.Select().From("users").ForUpdate(),
			"ForShare":            db.Select().From("users").ForShare(),
			"ForUpdateSkipLocked": db.Select().From("users").ForUpdateSkipLocked(),
		} {
			q := sq.Build()
			if assert.NotNil(t, q.LastError, driver) {
				assert.Equal(t, method+" requires a transaction: outside one the lock is released when the statement ends", q.LastError.Error(), driver)
			}
		}
		assert.Nil(t, db.Select().From("users").Build().LastError, driver)
	}
	// any other executor is allowed
	b, _ := newRecordingBuilder("mysql", NewMysqlBuilder)
	q := b.Select().From("users").ForUpdate().Build()
	assert.Nil(t, q.LastError)
	assert.Equal(t, "SELECT * FROM `users` FOR UPDATE", q.SQL())

	// databases without support
	for _, driver := range []string{"mssql", "oci8"} {
		tx := txOn(driver)
		assert.NotNil(t, tx.Select().From("users").ForUpdate().Build().LastError, driver)
		assert.NotNil(t, tx.Select().From("users").ForShare().Build().LastError, driver)
		assert.NotNil(t, tx.Select().From("users").ForUpdateSkipLocked().Build().LastError, driver)
		assert.Nil(t, tx.Select().From("users").Build().LastError, driver)
	}

	// clauses a lock cannot be combined with, on every database
	for _, driver := range []string{"mysql", "postgres", "sqlite", "unknown"} {
		db := txOn(driver)
		other := db.Select().From("posts").Build()
		assert.NotNil(t, db.Select().From("users").Union(other).ForUpdate().Build().LastError, driver+" union")
		assert.NotNil(t, db.Select().From("users").UnionAll(other).ForShare().Build().LastError, driver+" union all")
		assert.NotNil(t, db.Select("status").From("users").GroupBy("status").ForUpdate().Build().LastError, driver+" group by")
		assert.NotNil(t, db.Select("status").From("users").Distinct(true).ForUpdateSkipLocked().Build().LastError, driver+" distinct")
	}

	// the error is returned by the executing methods
	var n int
	assert.NotNil(t, txOn("mssql").Select("id").From("users").ForUpdate().Row(&n))
}

// plainQueryBuilderBuilder is a Builder whose QueryBuilder implements only QueryBuilder,
// like a custom query builder written before LockQueryBuilder existed.
type plainQueryBuilderBuilder struct {
	Builder
}

type plainQueryBuilder struct {
	QueryBuilder
}

func (b plainQueryBuilderBuilder) QueryBuilder() QueryBuilder {
	return plainQueryBuilder{b.Builder.QueryBuilder()}
}

func TestSelectQuery_LockCustomQueryBuilder(t *testing.T) {
	db := NewFromDB(nil, "mysql")
	b := plainQueryBuilderBuilder{db.Wrap(&sql.Tx{}).Builder}

	q := NewSelectQuery(b, db).From("users").ForUpdate().Build()
	assert.NotNil(t, q.LastError)

	q = NewSelectQuery(b, db).From("users").Build()
	assert.Nil(t, q.LastError)
	assert.Equal(t, "SELECT * FROM `users`", q.SQL())
}
