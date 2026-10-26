// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMysqlBuilder_QuoteSimpleTableName(t *testing.T) {
	b := getMysqlBuilder()
	assert.Equal(t, b.QuoteSimpleTableName(`abc`), "`abc`", "t1")
	assert.Equal(t, b.QuoteSimpleTableName("`abc`"), "`abc`", "t2")
	assert.Equal(t, b.QuoteSimpleTableName(`{{abc}}`), "`{{abc}}`", "t3")
	assert.Equal(t, b.QuoteSimpleTableName(`a.bc`), "`a.bc`", "t4")
}

func TestMysqlBuilder_QuoteSimpleColumnName(t *testing.T) {
	b := getMysqlBuilder()
	assert.Equal(t, b.QuoteSimpleColumnName(`abc`), "`abc`", "t1")
	assert.Equal(t, b.QuoteSimpleColumnName("`abc`"), "`abc`", "t2")
	assert.Equal(t, b.QuoteSimpleColumnName(`{{abc}}`), "`{{abc}}`", "t3")
	assert.Equal(t, b.QuoteSimpleColumnName(`a.bc`), "`a.bc`", "t4")
	assert.Equal(t, b.QuoteSimpleColumnName(`*`), `*`, "t5")
}

func TestMysqlBuilder_Upsert(t *testing.T) {
	b := getMysqlBuilder()
	q := b.Upsert("users", Params{
		"name": "James",
		"age":  30,
	})
	// the update part copies the proposed row, so each value is bound once
	assert.Equal(t, "INSERT INTO `users` (`age`, `name`) VALUES ({:p0}, {:p1}) ON DUPLICATE KEY UPDATE `age`=VALUES(`age`), `name`=VALUES(`name`)", q.SQL(), "t1")
	assert.Equal(t, Params{"p0": 30, "p1": "James"}, q.Params(), "t2")
	// the executed statement must carry the ON DUPLICATE KEY UPDATE clause
	assert.Equal(t, "INSERT INTO `users` (`age`, `name`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `age`=VALUES(`age`), `name`=VALUES(`name`)", q.rawSQL, "t4")
	assert.Equal(t, []string{"p0", "p1"}, q.placeholders, "t5")
	assert.Nil(t, q.LastError, "t6")
}

func TestMysqlBuilder_UpsertConstraints(t *testing.T) {
	b := getMysqlBuilder()

	// constraint columns are left out of the update list
	q := b.Upsert("users", Params{
		"id":   1,
		"name": "James",
		"age":  30,
	}, "id")
	assert.Equal(t, "INSERT INTO `users` (`age`, `id`, `name`) VALUES ({:p0}, {:p1}, {:p2}) ON DUPLICATE KEY UPDATE `age`=VALUES(`age`), `name`=VALUES(`name`)", q.SQL(), "t1")
	assert.Equal(t, "INSERT INTO `users` (`age`, `id`, `name`) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE `age`=VALUES(`age`), `name`=VALUES(`name`)", q.rawSQL, "t2")
	assert.Equal(t, Params{"p0": 30, "p1": 1, "p2": "James"}, q.Params(), "t3")
	assert.Equal(t, []string{"p0", "p1", "p2"}, q.placeholders, "t3b")
	assert.Nil(t, q.LastError, "t4")

	// an expression is rendered once, in the insert part; the update part copies its result
	q = b.Upsert("users", Params{
		"id":      1,
		"updated": NewExp("NOW()"),
	}, "id")
	assert.Equal(t, "INSERT INTO `users` (`id`, `updated`) VALUES ({:p0}, NOW()) ON DUPLICATE KEY UPDATE `updated`=VALUES(`updated`)", q.SQL(), "t5")
	assert.Nil(t, q.LastError, "t6")

	// nothing to update besides the constraint columns
	q = b.Upsert("users", Params{"id": 1}, "id")
	assert.NotNil(t, q.LastError, "t7")
	q = b.Upsert("users", Params{})
	assert.NotNil(t, q.LastError, "t8")
}

func TestMysqlBuilder_DropPrimaryKey(t *testing.T) {
	b := getMysqlBuilder()
	q := b.DropPrimaryKey("users", "pk")
	assert.Equal(t, q.SQL(), "ALTER TABLE `users` DROP PRIMARY KEY", "t1")
}

func TestMysqlBuilder_DropForeignKey(t *testing.T) {
	b := getMysqlBuilder()
	q := b.DropForeignKey("users", "fk")
	assert.Equal(t, q.SQL(), "ALTER TABLE `users` DROP FOREIGN KEY `fk`", "t1")
}

func getMysqlBuilder() Builder {
	db := NewFromDB(nil, "mysql")
	b := NewMysqlBuilder(db, nil)
	db.Builder = b
	return b
}
