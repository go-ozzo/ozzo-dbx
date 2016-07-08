// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMssqlBuilder_QuoteSimpleTableName(t *testing.T) {
	b := getMssqlBuilder()
	assert.Equal(t, b.QuoteSimpleTableName(`abc`), "[abc]", "t1")
	assert.Equal(t, b.QuoteSimpleTableName("[abc]"), "[abc]", "t2")
	assert.Equal(t, b.QuoteSimpleTableName(`{{abc}}`), "[{{abc}}]", "t3")
	assert.Equal(t, b.QuoteSimpleTableName(`a.bc`), "[a.bc]", "t4")
}

func TestMssqlBuilder_QuoteSimpleColumnName(t *testing.T) {
	b := getMssqlBuilder()
	assert.Equal(t, b.QuoteSimpleColumnName(`abc`), "[abc]", "t1")
	assert.Equal(t, b.QuoteSimpleColumnName("[abc]"), "[abc]", "t2")
	assert.Equal(t, b.QuoteSimpleColumnName(`{{abc}}`), "[{{abc}}]", "t3")
	assert.Equal(t, b.QuoteSimpleColumnName(`a.bc`), "[a.bc]", "t4")
	assert.Equal(t, b.QuoteSimpleColumnName(`*`), `*`, "t5")
}

func TestMssqlBuilder_RenameTable(t *testing.T) {
	b := getMssqlBuilder()
	q := b.RenameTable("users", "user")
	assert.Equal(t, q.SQL(), `sp_name 'users', 'user'`, "t1")
}

func TestMssqlBuilder_RenameColumn(t *testing.T) {
	b := getMssqlBuilder()
	q := b.RenameColumn("users", "name", "username")
	assert.Equal(t, q.SQL(), `sp_name 'users.name', 'username', 'COLUMN'`, "t1")
}

func TestMssqlBuilder_AlterColumn(t *testing.T) {
	b := getMssqlBuilder()
	q := b.AlterColumn("users", "name", "int")
	assert.Equal(t, q.SQL(), `ALTER TABLE [users] ALTER COLUMN [name] int`, "t1")
}

func TestMssqlQueryBuilder_BuildOrderByAndLimit(t *testing.T) {
	qb := getMssqlBuilder().QueryBuilder()

	sql := qb.BuildOrderByAndLimit("SELECT *", []string{"name"}, 10, 2)
	expected := "SELECT *\nORDER BY [name]\nOFFSET 2 ROWS\nFETCH NEXT 10 ROWS ONLY"
	assert.Equal(t, sql, expected, "t1")

	sql = qb.BuildOrderByAndLimit("SELECT *", nil, -1, -1)
	expected = "SELECT *"
	assert.Equal(t, sql, expected, "t2")

	sql = qb.BuildOrderByAndLimit("SELECT *", []string{"name"}, -1, -1)
	expected = "SELECT *\nORDER BY [name]"
	assert.Equal(t, sql, expected, "t3")

	sql = qb.BuildOrderByAndLimit("SELECT *", nil, 10, -1)
	expected = "SELECT *\nORDER BY (SELECT NULL)\nOFFSET 0 ROWS\nFETCH NEXT 10 ROWS ONLY"
	assert.Equal(t, sql, expected, "t4")
}

func getMssqlBuilder() Builder {
	db := getDB()
	b := NewMssqlBuilder(db, db.sqlDB)
	db.Builder = b
	return b
}
