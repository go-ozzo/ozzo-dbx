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
	assert.Equal(t, q.SQL(), `INSERT INTO "users" ("age", "name") VALUES ({:p0}, {:p1}) ON CONFLICT ("id") DO UPDATE SET "age"={:p2}, "name"={:p3}`, "t1")
	assert.Equal(t, q.Params()["p0"], 30, "t2")
	assert.Equal(t, q.Params()["p1"], "James", "t3")
	assert.Equal(t, q.Params()["p2"], 30, "t2")
	assert.Equal(t, q.Params()["p3"], "James", "t3")
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
	db := getDB()
	b := NewPgsqlBuilder(db, db.sqlDB)
	db.Builder = b
	return b
}
