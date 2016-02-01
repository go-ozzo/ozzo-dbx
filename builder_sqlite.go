// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dbx

import (
	"errors"
	"fmt"
	"strings"
)

// SqliteBuilder is the builder for SQLite databases.
type SqliteBuilder struct {
	*BaseBuilder
	qb *BaseQueryBuilder
}

// NewSqliteBuilder creates a new SqliteBuilder instance.
func NewSqliteBuilder(db *DB, executor Executor) Builder {
	return &SqliteBuilder{
		NewBaseBuilder(db, executor),
		NewBaseQueryBuilder(db),
	}
}

func (b *SqliteBuilder) QueryBuilder() QueryBuilder {
	return b.qb
}

func (b *SqliteBuilder) QuoteSimpleTableName(s string) string {
	if strings.ContainsAny(s, "`") {
		return s
	}
	return "`" + s + "`"
}

func (b *SqliteBuilder) QuoteSimpleColumnName(s string) string {
	if strings.Contains(s, "`") || s == "*" {
		return s
	}
	return "`" + s + "`"
}

func (b *SqliteBuilder) DropIndex(table, name string) *Query {
	sql := fmt.Sprintf("DROP INDEX %v", b.db.QuoteColumnName(name))
	return b.NewQuery(sql)
}

func (b *SqliteBuilder) TruncateTable(table string) *Query {
	sql := "DELETE FROM " + b.db.QuoteTableName(table)
	return b.NewQuery(sql)
}

func (b *SqliteBuilder) DropColumn(table, col string) *Query {
	q := b.NewQuery("")
	q.LastError = errors.New("SQLite does not support dropping columns")
	return q
}

func (b *SqliteBuilder) RenameColumn(table, oldName, newName string) *Query {
	q := b.NewQuery("")
	q.LastError = errors.New("SQLite does not support renaming columns")
	return q
}

func (b *SqliteBuilder) AlterColumn(table, col, typ string) *Query {
	q := b.NewQuery("")
	q.LastError = errors.New("SQLite does not support altering column")
	return q
}

func (b *SqliteBuilder) AddPrimaryKey(table, name string, cols ...string) *Query {
	q := b.NewQuery("")
	q.LastError = errors.New("SQLite does not support adding primary key")
	return q
}

func (b *SqliteBuilder) DropPrimaryKey(table, name string) *Query {
	q := b.NewQuery("")
	q.LastError = errors.New("SQLite does not support dropping primary key")
	return q
}

func (b *SqliteBuilder) AddForeignKey(table, name string, cols, refCols []string, refTable string, options ...string) *Query {
	q := b.NewQuery("")
	q.LastError = errors.New("SQLite does not support adding foreign keys")
	return q
}

func (b *SqliteBuilder) DropForeignKey(table, name string) *Query {
	q := b.NewQuery("")
	q.LastError = errors.New("SQLite does not support dropping foreign keys")
	return q
}
