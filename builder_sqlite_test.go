// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSqliteBuilder_QuoteSimpleTableName(t *testing.T) {
	b := getSqliteBuilder()
	assert.Equal(t, b.QuoteSimpleTableName(`abc`), "`abc`", "t1")
	assert.Equal(t, b.QuoteSimpleTableName("`abc`"), "`abc`", "t2")
	assert.Equal(t, b.QuoteSimpleTableName(`{{abc}}`), "`{{abc}}`", "t3")
	assert.Equal(t, b.QuoteSimpleTableName(`a.bc`), "`a.bc`", "t4")
}

func TestSqliteBuilder_QuoteSimpleColumnName(t *testing.T) {
	b := getSqliteBuilder()
	assert.Equal(t, b.QuoteSimpleColumnName(`abc`), "`abc`", "t1")
	assert.Equal(t, b.QuoteSimpleColumnName("`abc`"), "`abc`", "t2")
	assert.Equal(t, b.QuoteSimpleColumnName(`{{abc}}`), "`{{abc}}`", "t3")
	assert.Equal(t, b.QuoteSimpleColumnName(`a.bc`), "`a.bc`", "t4")
	assert.Equal(t, b.QuoteSimpleColumnName(`*`), `*`, "t5")
}

func TestSqliteBuilder_Upsert(t *testing.T) {
	b := getSqliteBuilder()
	q := b.Upsert("users", Params{
		"id":   1,
		"name": "James",
		"age":  30,
	}, "id")
	assert.Equal(t, "INSERT INTO `users` (`age`, `id`, `name`) VALUES ({:p0}, {:p1}, {:p2}) ON CONFLICT (`id`) DO UPDATE SET `age`=EXCLUDED.`age`, `name`=EXCLUDED.`name`", q.SQL(), "t1")
	assert.Equal(t, "INSERT INTO `users` (`age`, `id`, `name`) VALUES (?, ?, ?) ON CONFLICT (`id`) DO UPDATE SET `age`=EXCLUDED.`age`, `name`=EXCLUDED.`name`", q.rawSQL, "t2")
	assert.Equal(t, Params{"p0": 30, "p1": 1, "p2": "James"}, q.Params(), "t3")
	assert.Equal(t, []string{"p0", "p1", "p2"}, q.placeholders, "t3b")
	assert.Nil(t, q.LastError, "t4")

	// SQLite requires a conflict target for DO UPDATE
	q = b.Upsert("users", Params{"name": "James"})
	assert.NotNil(t, q.LastError, "t5")

	// nothing to update besides the conflict columns
	q = b.Upsert("users", Params{"id": 1}, "id")
	assert.NotNil(t, q.LastError, "t6")
}

func TestSqliteBuilder_DropIndex(t *testing.T) {
	b := getSqliteBuilder()
	q := b.DropIndex("users", "idx")
	assert.Equal(t, q.SQL(), "DROP INDEX `idx`", "t1")
}

func TestSqliteBuilder_TruncateTable(t *testing.T) {
	b := getSqliteBuilder()
	q := b.TruncateTable("users")
	assert.Equal(t, q.SQL(), "DELETE FROM `users`", "t1")
}

func TestSqliteBuilder_DropColumn(t *testing.T) {
	b := getSqliteBuilder()
	q := b.DropColumn("users", "age")
	assert.Equal(t, q.SQL(), "ALTER TABLE `users` DROP COLUMN `age`", "t1")
}

func TestSqliteBuilder_RenameColumn(t *testing.T) {
	b := getSqliteBuilder()
	q := b.RenameColumn("users", "name", "username")
	assert.Equal(t, q.SQL(), "ALTER TABLE `users` RENAME COLUMN `name` TO `username`", "t1")
}

func TestSqliteBuilder_RenameTable(t *testing.T) {
	b := getSqliteBuilder()
	q := b.RenameTable("users", "people")
	assert.Equal(t, q.SQL(), "ALTER TABLE `users` RENAME TO `people`", "t1")
}

func TestSqliteBuilder_AlterColumn(t *testing.T) {
	b := getSqliteBuilder()
	q := b.AlterColumn("users", "name", "int")
	assert.NotEqual(t, q.LastError, nil, "t1")
}

func TestSqliteBuilder_AddPrimaryKey(t *testing.T) {
	b := getSqliteBuilder()
	q := b.AddPrimaryKey("users", "pk", "id1", "id2")
	assert.NotEqual(t, q.LastError, nil, "t1")
}

func TestSqliteBuilder_DropPrimaryKey(t *testing.T) {
	b := getSqliteBuilder()
	q := b.DropPrimaryKey("users", "pk")
	assert.NotEqual(t, q.LastError, nil, "t1")
}

func TestSqliteBuilder_AddForeignKey(t *testing.T) {
	b := getSqliteBuilder()
	q := b.AddForeignKey("users", "fk", []string{"p1", "p2"}, []string{"f1", "f2"}, "profile", "opt")
	assert.NotEqual(t, q.LastError, nil, "t1")
}

func TestSqliteBuilder_DropForeignKey(t *testing.T) {
	b := getSqliteBuilder()
	q := b.DropForeignKey("users", "fk")
	assert.NotEqual(t, q.LastError, nil, "t1")
}

func getSqliteBuilder() Builder {
	db := NewFromDB(nil, "sqlite3")
	b := NewSqliteBuilder(db, nil)
	db.Builder = b
	return b
}
