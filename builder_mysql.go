// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dbx

import (
	"fmt"
	"regexp"
	"strings"
)

// MysqlBuilder is the builder for MySQL databases.
type MysqlBuilder struct {
	*BaseBuilder
	qb *BaseQueryBuilder
}

// NewMysqlBuilder creates a new MysqlBuilder instance.
func NewMysqlBuilder(db *DB, executor Executor) Builder {
	return &MysqlBuilder{
		NewBaseBuilder(db, executor),
		NewBaseQueryBuilder(db),
	}
}

func (b *MysqlBuilder) QueryBuilder() QueryBuilder {
	return b.qb
}

func (b *MysqlBuilder) QuoteSimpleTableName(s string) string {
	if strings.ContainsAny(s, "`") {
		return s
	}
	return "`" + s + "`"
}

func (b *MysqlBuilder) QuoteSimpleColumnName(s string) string {
	if strings.Contains(s, "`") || s == "*" {
		return s
	}
	return "`" + s + "`"
}

func (b *MysqlBuilder) RenameColumn(table, oldName, newName string) *Query {
	qt := b.db.QuoteTableName(table)
	sql := fmt.Sprintf("ALTER TABLE %v CHANGE %v %v", qt, b.db.QuoteColumnName(oldName), b.db.QuoteColumnName(newName))

	var info struct {
		SQL string `db:"Create Table"`
	}
	if err := b.db.NewQuery("SHOW CREATE TABLE " + qt).One(&info); err != nil {
		return b.db.NewQuery(sql)
	}

	regex := regexp.MustCompile("(?m)^\\s*`(.*?)`\\s+(.*?),?$")
	if matches := regex.FindAllStringSubmatch(info.SQL, -1); matches != nil {
		for _, match := range matches {
			if match[1] == oldName {
				sql += " " + match[2]
				break
			}
		}
	}

	return b.db.NewQuery(sql)
}

func (b *MysqlBuilder) DropPrimaryKey(table, name string) *Query {
	sql := fmt.Sprintf("ALTER TABLE %v DROP PRIMARY KEY", b.db.QuoteTableName(table))
	return b.db.NewQuery(sql)
}

func (b *MysqlBuilder) DropForeignKey(table, name string) *Query {
	sql := fmt.Sprintf("ALTER TABLE %v DROP FOREIGN KEY %v", b.db.QuoteTableName(table), b.db.QuoteColumnName(name))
	return b.db.NewQuery(sql)
}
