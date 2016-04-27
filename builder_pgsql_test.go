// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dbx

import (
	"testing"
)

func TestPgsqlBuilder_Upsert(t *testing.T) {
	b := getPgsqlBuilder()
	q := b.Upsert("users", Params{
		"name": "James",
		"age":  30,
	}, "id")
	assertEqual(t, q.SQL(), `INSERT INTO "users" ("age", "name") VALUES ({:p0}, {:p1}) ON CONFLICT ("id") DO UPDATE SET "age"={:p2}, "name"={:p3}`, "t1")
	assertEqual(t, q.Params()["p0"], 30, "t2")
	assertEqual(t, q.Params()["p1"], "James", "t3")
	assertEqual(t, q.Params()["p2"], 30, "t2")
	assertEqual(t, q.Params()["p3"], "James", "t3")
}
func TestPgsqlBuilder_DropIndex(t *testing.T) {
	b := getPgsqlBuilder()
	q := b.DropIndex("users", "idx")
	assertEqual(t, q.SQL(), `DROP INDEX "idx"`, "t1")
}

func TestPgsqlBuilder_RenameTable(t *testing.T) {
	b := getPgsqlBuilder()
	q := b.RenameTable("users", "user")
	assertEqual(t, q.SQL(), `ALTER TABLE "users" RENAME TO "user"`, "t1")
}

func TestPgsqlBuilder_AlterColumn(t *testing.T) {
	b := getPgsqlBuilder()
	q := b.AlterColumn("users", "name", "int")
	assertEqual(t, q.SQL(), `ALTER TABLE "users" ALTER COLUMN "name" TYPE int`, "t1")
}

func getPgsqlBuilder() Builder {
	db := getDB()
	b := NewPgsqlBuilder(db, db.sqlDB)
	db.Builder = b
	return b
}
