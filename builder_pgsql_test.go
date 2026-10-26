// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPgsqlBuilder_Upsert(t *testing.T) {
	b := getPgsqlBuilder()
	q := b.Upsert("users", Params{
		"name": "James",
		"age":  30,
	}, "id")
	assert.Equal(t, q.sql, `INSERT INTO "users" ("age", "name") VALUES ({:p0}, {:p1}) ON CONFLICT ("id") DO UPDATE SET "age"={:p2}, "name"={:p3}`, "t1")
	assert.Equal(t, q.rawSQL, `INSERT INTO "users" ("age", "name") VALUES ($1, $2) ON CONFLICT ("id") DO UPDATE SET "age"=$3, "name"=$4`, "t2")
	assert.Equal(t, q.Params()["p0"], 30, "t3")
	assert.Equal(t, q.Params()["p1"], "James", "t4")
	assert.Equal(t, q.Params()["p2"], 30, "t5")
	assert.Equal(t, q.Params()["p3"], "James", "t6")
	assert.Nil(t, q.LastError, "t7")
}

func TestPgsqlBuilder_UpsertConstraints(t *testing.T) {
	b := getPgsqlBuilder()

	// conflict columns are left out of the update list
	q := b.Upsert("users", Params{
		"id":   1,
		"name": "James",
	}, "id")
	assert.Equal(t, `INSERT INTO "users" ("id", "name") VALUES ({:p0}, {:p1}) ON CONFLICT ("id") DO UPDATE SET "name"={:p2}`, q.SQL(), "t1")
	assert.Equal(t, `INSERT INTO "users" ("id", "name") VALUES ($1, $2) ON CONFLICT ("id") DO UPDATE SET "name"=$3`, q.rawSQL, "t2")
	assert.Nil(t, q.LastError, "t3")

	// PostgreSQL rejects ON CONFLICT DO UPDATE without a conflict target
	q = b.Upsert("users", Params{"name": "James"})
	assert.NotNil(t, q.LastError, "t4")

	// nothing to update besides the conflict columns
	q = b.Upsert("users", Params{"id": 1}, "id")
	assert.NotNil(t, q.LastError, "t5")
}
func TestPgsqlBuilder_DropIndex(t *testing.T) {
	b := getPgsqlBuilder()
	q := b.DropIndex("users", "idx")
	assert.Equal(t, q.SQL(), `DROP INDEX "idx"`, "t1")
}

func TestPgsqlBuilder_RenameTable(t *testing.T) {
	b := getPgsqlBuilder()
	q := b.RenameTable("users", "user")
	assert.Equal(t, q.SQL(), `ALTER TABLE "users" RENAME TO "user"`, "t1")
}

func TestPgsqlBuilder_AlterColumn(t *testing.T) {
	b := getPgsqlBuilder()
	q := b.AlterColumn("users", "name", "int")
	assert.Equal(t, q.SQL(), `ALTER TABLE "users" ALTER COLUMN "name" TYPE int`, "t1")
}

func getPgsqlBuilder() Builder {
	db := NewFromDB(nil, "postgres")
	b := NewPgsqlBuilder(db, nil)
	db.Builder = b
	return b
}
