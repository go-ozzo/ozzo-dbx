// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dbx

import (
	"testing"
)

func TestStandardBuilder_Quote(t *testing.T) {
	b := getStandardBuilder()
	assertEqual(t, b.Quote(`abc`), `'abc'`, "t1")
	assertEqual(t, b.Quote(`I'm`), `'I''m'`, "t2")
	assertEqual(t, b.Quote(``), `''`, "t3")
}

func TestStandardBuilder_QuoteSimpleTableName(t *testing.T) {
	b := getStandardBuilder()
	assertEqual(t, b.QuoteSimpleTableName(`abc`), `"abc"`, "t1")
	assertEqual(t, b.QuoteSimpleTableName(`"abc"`), `"abc"`, "t2")
	assertEqual(t, b.QuoteSimpleTableName(`{{abc}}`), `"{{abc}}"`, "t3")
	assertEqual(t, b.QuoteSimpleTableName(`a.bc`), `"a.bc"`, "t4")
}

func TestStandardBuilder_QuoteSimpleColumnName(t *testing.T) {
	b := getStandardBuilder()
	assertEqual(t, b.QuoteSimpleColumnName(`abc`), `"abc"`, "t1")
	assertEqual(t, b.QuoteSimpleColumnName(`"abc"`), `"abc"`, "t2")
	assertEqual(t, b.QuoteSimpleColumnName(`{{abc}}`), `"{{abc}}"`, "t3")
	assertEqual(t, b.QuoteSimpleColumnName(`a.bc`), `"a.bc"`, "t4")
	assertEqual(t, b.QuoteSimpleColumnName(`*`), `*`, "t5")
}

func TestStandardBuilder_Insert(t *testing.T) {
	b := getStandardBuilder()
	q := b.Insert("users", Params{
		"name": "James",
		"age":  30,
	})
	assertEqual(t, q.SQL(), `INSERT INTO "users" ("age", "name") VALUES ({:p0}, {:p1})`, "t1")
	assertEqual(t, q.Params()["p0"], 30, "t2")
	assertEqual(t, q.Params()["p1"], "James", "t3")

	q = b.Insert("users", Params{})
	assertEqual(t, q.SQL(), `INSERT INTO "users" DEFAULT VALUES`, "t2")
}

func TestStandardBuilder_Upsert(t *testing.T) {
	b := getStandardBuilder()
	q := b.Upsert("users", Params{
		"name": "James",
		"age":  30,
	})
	assertNotEqual(t, q.LastError, nil, "t1")
}

func TestStandardBuilder_Update(t *testing.T) {
	b := getStandardBuilder()
	q := b.Update("users", Params{
		"name": "James",
		"age":  30,
	}, NewExp("id=10"))
	assertEqual(t, q.SQL(), `UPDATE "users" SET "age"={:p0}, "name"={:p1} WHERE id=10`, "t1")
	assertEqual(t, q.Params()["p0"], 30, "t2")
	assertEqual(t, q.Params()["p1"], "James", "t3")

	q = b.Update("users", Params{
		"name": "James",
		"age":  30,
	}, nil)
	assertEqual(t, q.SQL(), `UPDATE "users" SET "age"={:p0}, "name"={:p1}`, "t2")
}

func TestStandardBuilder_Delete(t *testing.T) {
	b := getStandardBuilder()
	q := b.Delete("users", NewExp("id=10"))
	assertEqual(t, q.SQL(), `DELETE FROM "users" WHERE id=10`, "t1")
	q = b.Delete("users", nil)
	assertEqual(t, q.SQL(), `DELETE FROM "users"`, "t2")
}

func TestStandardBuilder_CreateTable(t *testing.T) {
	b := getStandardBuilder()
	q := b.CreateTable("users", map[string]string{
		"id":   "int primary key",
		"name": "varchar(255)",
	}, "ON DELETE CASCADE")
	assertEqual(t, q.SQL(), "CREATE TABLE \"users\" (\"id\" int primary key, \"name\" varchar(255)) ON DELETE CASCADE", "t1")
}

func TestStandardBuilder_RenameTable(t *testing.T) {
	b := getStandardBuilder()
	q := b.RenameTable("users", "user")
	assertEqual(t, q.SQL(), `RENAME TABLE "users" TO "user"`, "t1")
}

func TestStandardBuilder_DropTable(t *testing.T) {
	b := getStandardBuilder()
	q := b.DropTable("users")
	assertEqual(t, q.SQL(), `DROP TABLE "users"`, "t1")
}

func TestStandardBuilder_TruncateTable(t *testing.T) {
	b := getStandardBuilder()
	q := b.TruncateTable("users")
	assertEqual(t, q.SQL(), `TRUNCATE TABLE "users"`, "t1")
}

func TestStandardBuilder_AddColumn(t *testing.T) {
	b := getStandardBuilder()
	q := b.AddColumn("users", "age", "int")
	assertEqual(t, q.SQL(), `ALTER TABLE "users" ADD "age" int`, "t1")
}

func TestStandardBuilder_DropColumn(t *testing.T) {
	b := getStandardBuilder()
	q := b.DropColumn("users", "age")
	assertEqual(t, q.SQL(), `ALTER TABLE "users" DROP COLUMN "age"`, "t1")
}

func TestStandardBuilder_RenameColumn(t *testing.T) {
	b := getStandardBuilder()
	q := b.RenameColumn("users", "name", "username")
	assertEqual(t, q.SQL(), `ALTER TABLE "users" RENAME COLUMN "name" TO "username"`, "t1")
}

func TestStandardBuilder_AlterColumn(t *testing.T) {
	b := getStandardBuilder()
	q := b.AlterColumn("users", "name", "int")
	assertEqual(t, q.SQL(), `ALTER TABLE "users" CHANGE "name" "name" int`, "t1")
}

func TestStandardBuilder_AddPrimaryKey(t *testing.T) {
	b := getStandardBuilder()
	q := b.AddPrimaryKey("users", "pk", "id1", "id2")
	assertEqual(t, q.SQL(), `ALTER TABLE "users" ADD CONSTRAINT "pk" PRIMARY KEY ("id1", "id2")`, "t1")
}

func TestStandardBuilder_DropPrimaryKey(t *testing.T) {
	b := getStandardBuilder()
	q := b.DropPrimaryKey("users", "pk")
	assertEqual(t, q.SQL(), `ALTER TABLE "users" DROP CONSTRAINT "pk"`, "t1")
}

func TestStandardBuilder_AddForeignKey(t *testing.T) {
	b := getStandardBuilder()
	q := b.AddForeignKey("users", "fk", []string{"p1", "p2"}, []string{"f1", "f2"}, "profile", "opt")
	assertEqual(t, q.SQL(), `ALTER TABLE "users" ADD CONSTRAINT "fk" FOREIGN KEY ("p1", "p2") REFERENCES "profile" ("f1", "f2") opt`, "t1")
}

func TestStandardBuilder_DropForeignKey(t *testing.T) {
	b := getStandardBuilder()
	q := b.DropForeignKey("users", "fk")
	assertEqual(t, q.SQL(), `ALTER TABLE "users" DROP CONSTRAINT "fk"`, "t1")
}

func TestStandardBuilder_CreateIndex(t *testing.T) {
	b := getStandardBuilder()
	q := b.CreateIndex("users", "idx", "id1", "id2")
	assertEqual(t, q.SQL(), `CREATE INDEX "idx" ON "users" ("id1", "id2")`, "t1")
}

func TestStandardBuilder_CreateUniqueIndex(t *testing.T) {
	b := getStandardBuilder()
	q := b.CreateUniqueIndex("users", "idx", "id1", "id2")
	assertEqual(t, q.SQL(), `CREATE UNIQUE INDEX "idx" ON "users" ("id1", "id2")`, "t1")
}

func TestStandardBuilder_DropIndex(t *testing.T) {
	b := getStandardBuilder()
	q := b.DropIndex("users", "idx")
	assertEqual(t, q.SQL(), `DROP INDEX "idx" ON "users"`, "t1")
}

func getStandardBuilder() Builder {
	db := getDB()
	b := NewStandardBuilder(db, db.sqlDB)
	db.Builder = b
	return b
}
