// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbx

import (
	"testing"
)

func TestOciBuilder_DropIndex(t *testing.T) {
	b := getOciBuilder()
	q := b.DropIndex("users", "idx")
	assertEqual(t, q.SQL(), `DROP INDEX "idx"`, "t1")
}

func TestOciBuilder_RenameTable(t *testing.T) {
	b := getOciBuilder()
	q := b.RenameTable("users", "user")
	assertEqual(t, q.SQL(), `ALTER TABLE "users" RENAME TO "user"`, "t1")
}

func TestOciBuilder_AlterColumn(t *testing.T) {
	b := getOciBuilder()
	q := b.AlterColumn("users", "name", "int")
	assertEqual(t, q.SQL(), `ALTER TABLE "users" MODIFY "name" int`, "t1")
}

func TestOciQueryBuilder_BuildOrderByAndLimit(t *testing.T) {
	qb := getOciBuilder().QueryBuilder()

	sql := qb.BuildOrderByAndLimit("SELECT *", []string{"name"}, 10, 2)
	expected := "WITH USER_SQL AS (SELECT *\nORDER BY \"name\"),\n\tPAGINATION AS (SELECT USER_SQL.*, rownum as rowNumId FROM USER_SQL)\nSELECT * FROM PAGINATION WHERE rowNumId > 2 AND rowNum <= 10"
	assertEqual(t, sql, expected, "t1")

	sql = qb.BuildOrderByAndLimit("SELECT *", nil, -1, -1)
	expected = "SELECT *"
	assertEqual(t, sql, expected, "t2")

	sql = qb.BuildOrderByAndLimit("SELECT *", []string{"name"}, -1, -1)
	expected = "SELECT *\nORDER BY \"name\""
	assertEqual(t, sql, expected, "t3")

	sql = qb.BuildOrderByAndLimit("SELECT *", nil, 10, -1)
	expected = "WITH USER_SQL AS (SELECT *),\n\tPAGINATION AS (SELECT USER_SQL.*, rownum as rowNumId FROM USER_SQL)\nSELECT * FROM PAGINATION WHERE rowNum <= 10"
	assertEqual(t, sql, expected, "t4")
}

func getOciBuilder() Builder {
	db := getDB()
	b := NewOciBuilder(db, db.sqlDB)
	db.Builder = b
	return b
}
