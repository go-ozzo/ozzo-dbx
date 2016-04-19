// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dbx

import (
	"fmt"
	"regexp"
	"strings"
	"sort"
)

// MysqlBuilder is the builder for MySQL databases.
type MysqlBuilder struct {
	*BaseBuilder
	qb *BaseQueryBuilder
}

var _ Builder = &MysqlBuilder{}

// NewMysqlBuilder creates a new MysqlBuilder instance.
func NewMysqlBuilder(db *DB, executor Executor) Builder {
	return &MysqlBuilder{
		NewBaseBuilder(db, executor),
		NewBaseQueryBuilder(db),
	}
}

// QueryBuilder returns the query builder supporting the current DB.
func (b *MysqlBuilder) QueryBuilder() QueryBuilder {
	return b.qb
}

// QuoteSimpleTableName quotes a simple table name.
// A simple table name does not contain any schema prefix.
func (b *MysqlBuilder) QuoteSimpleTableName(s string) string {
	if strings.ContainsAny(s, "`") {
		return s
	}
	return "`" + s + "`"
}

// QuoteSimpleColumnName quotes a simple column name.
// A simple column name does not contain any table prefix.
func (b *MysqlBuilder) QuoteSimpleColumnName(s string) string {
	if strings.Contains(s, "`") || s == "*" {
		return s
	}
	return "`" + s + "`"
}

// Upsert creates a Query that represents an UPSERT SQL statement.
// Upsert inserts a row into the table if the primary key or unique index is not found.
// Otherwise it will update the row with the new values.
// The keys of cols are the column names, while the values of cols are the corresponding column
// values to be inserted.
func (b *MysqlBuilder) Upsert(table string, cols Params, constraints ...string) *Query {
	q := b.Insert(table, cols)

	names := []string{}
	for name := range cols {
		names = append(names, name)
	}
	sort.Strings(names)

	lines := []string{}
	for _, name := range names {
		value := cols[name]
		name = b.db.QuoteColumnName(name)
		if e, ok := value.(Expression); ok {
			lines = append(lines, name+"="+e.Build(b.db, q.params))
		} else {
			lines = append(lines, fmt.Sprintf("%v={:p%v}", name, len(q.params)))
			q.params[fmt.Sprintf("p%v", len(q.params))] = value
		}
	}

	q.sql += " ON DUPLICATE KEY UPDATE " + strings.Join(lines, ", ")

	return q
}

// RenameColumn creates a Query that can be used to rename a column in a table.
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

// DropPrimaryKey creates a Query that can be used to remove the named primary key constraint from a table.
func (b *MysqlBuilder) DropPrimaryKey(table, name string) *Query {
	sql := fmt.Sprintf("ALTER TABLE %v DROP PRIMARY KEY", b.db.QuoteTableName(table))
	return b.db.NewQuery(sql)
}

// DropForeignKey creates a Query that can be used to remove the named foreign key constraint from a table.
func (b *MysqlBuilder) DropForeignKey(table, name string) *Query {
	sql := fmt.Sprintf("ALTER TABLE %v DROP FOREIGN KEY %v", b.db.QuoteTableName(table), b.db.QuoteColumnName(name))
	return b.db.NewQuery(sql)
}
