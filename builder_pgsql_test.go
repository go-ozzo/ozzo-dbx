// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dbx

import (
	"testing"
)

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
	assertEqual(t, q.SQL(), `ALTER TABLE "users" ALTER COLUMN "name" int`, "t1")
}

func getPgsqlBuilder() Builder {
	db := getDB()
	b := NewPgsqlBuilder(db, db.BaseDB)
	db.Builder = b
	return b
}
