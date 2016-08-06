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
	getPreparedDB()
	b := getMysqlBuilder()
	q := b.Upsert("users", Params{
		"name": "James",
		"age":  30,
	})
	assert.Equal(t, q.SQL(), "INSERT INTO `users` (`age`, `name`) VALUES ({:p0}, {:p1}) ON DUPLICATE KEY UPDATE `age`={:p2}, `name`={:p3}", "t1")
	assert.Equal(t, q.Params()["p0"], 30, "t2")
	assert.Equal(t, q.Params()["p1"], "James", "t3")
	assert.Equal(t, q.Params()["p2"], 30, "t2")
	assert.Equal(t, q.Params()["p3"], "James", "t3")
}

func TestMysqlBuilder_RenameColumn(t *testing.T) {
	b := getMysqlBuilder()
	q := b.RenameColumn("users", "name", "username")
	assert.Equal(t, q.SQL(), "ALTER TABLE `users` CHANGE `name` `username`")
	q = b.RenameColumn("customer", "email", "e")
	assert.Equal(t, q.SQL(), "ALTER TABLE `customer` CHANGE `email` `e` varchar(128) NOT NULL")
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
	db := getDB()
	b := NewMysqlBuilder(db, db.sqlDB)
	db.Builder = b
	return b
}
