// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dbx

import (
	"testing"
)

func TestMysqlBuilder_QuoteSimpleTableName(t *testing.T) {
	b := getMysqlBuilder()
	assertEqual(t, b.QuoteSimpleTableName(`abc`), "`abc`", "t1")
	assertEqual(t, b.QuoteSimpleTableName("`abc`"), "`abc`", "t2")
	assertEqual(t, b.QuoteSimpleTableName(`{{abc}}`), "`{{abc}}`", "t3")
	assertEqual(t, b.QuoteSimpleTableName(`a.bc`), "`a.bc`", "t4")
}

func TestMysqlBuilder_QuoteSimpleColumnName(t *testing.T) {
	b := getMysqlBuilder()
	assertEqual(t, b.QuoteSimpleColumnName(`abc`), "`abc`", "t1")
	assertEqual(t, b.QuoteSimpleColumnName("`abc`"), "`abc`", "t2")
	assertEqual(t, b.QuoteSimpleColumnName(`{{abc}}`), "`{{abc}}`", "t3")
	assertEqual(t, b.QuoteSimpleColumnName(`a.bc`), "`a.bc`", "t4")
	assertEqual(t, b.QuoteSimpleColumnName(`*`), `*`, "t5")
}

func TestMysqlBuilder_RenameColumn(t *testing.T) {
	b := getMysqlBuilder()
	q := b.RenameColumn("users", "name", "username")
	assertEqual(t, q.SQL(), "ALTER TABLE `users` CHANGE `name` `username`", "t1")
}

func TestMysqlBuilder_DropPrimaryKey(t *testing.T) {
	b := getMysqlBuilder()
	q := b.DropPrimaryKey("users", "pk")
	assertEqual(t, q.SQL(), "ALTER TABLE `users` DROP PRIMARY KEY", "t1")
}

func TestMysqlBuilder_DropForeignKey(t *testing.T) {
	b := getMysqlBuilder()
	q := b.DropForeignKey("users", "fk")
	assertEqual(t, q.SQL(), "ALTER TABLE `users` DROP FOREIGN KEY `fk`", "t1")
}

func getMysqlBuilder() Builder {
	db := getDB()
	b := NewMysqlBuilder(db, db.sqlDB)
	db.Builder = b
	return b
}
