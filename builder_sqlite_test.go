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
	assert.NotEqual(t, q.LastError, nil, "t1")
}

func TestSqliteBuilder_RenameColumn(t *testing.T) {
	b := getSqliteBuilder()
	q := b.RenameColumn("users", "name", "username")
	assert.NotEqual(t, q.LastError, nil, "t1")
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
	db := getDB()
	b := NewSqliteBuilder(db, db.sqlDB)
	db.Builder = b
	return b
}
