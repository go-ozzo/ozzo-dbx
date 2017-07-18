// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
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

var _ Builder = &SqliteBuilder{}

// NewSqliteBuilder creates a new SqliteBuilder instance.
func NewSqliteBuilder(db *DB, executor Executor) Builder {
	return &SqliteBuilder{
		NewBaseBuilder(db, executor),
		NewBaseQueryBuilder(db),
	}
}

// QueryBuilder returns the query builder supporting the current DB.
func (b *SqliteBuilder) QueryBuilder() QueryBuilder {
	return b.qb
}

// Select returns a new SelectQuery object that can be used to build a SELECT statement.
// The parameters to this method should be the list column names to be selected.
// A column name may have an optional alias name. For example, Select("id", "my_name AS name").
func (b *SqliteBuilder) Select(cols ...string) *SelectQuery {
	return NewSelectQuery(b, b.db).Select(cols...)
}

// Model returns a new ModelQuery object that can be used to perform model-based DB operations.
// The model passed to this method should be a pointer to a model struct.
func (b *SqliteBuilder) Model(model interface{}) *ModelQuery {
	return NewModelQuery(model, b.db.FieldMapper, b.db, b)
}

// QuoteSimpleTableName quotes a simple table name.
// A simple table name does not contain any schema prefix.
func (b *SqliteBuilder) QuoteSimpleTableName(s string) string {
	if strings.ContainsAny(s, "`") {
		return s
	}
	return "`" + s + "`"
}

// QuoteSimpleColumnName quotes a simple column name.
// A simple column name does not contain any table prefix.
func (b *SqliteBuilder) QuoteSimpleColumnName(s string) string {
	if strings.Contains(s, "`") || s == "*" {
		return s
	}
	return "`" + s + "`"
}

// DropIndex creates a Query that can be used to remove the named index from a table.
func (b *SqliteBuilder) DropIndex(table, name string) *Query {
	sql := fmt.Sprintf("DROP INDEX %v", b.db.QuoteColumnName(name))
	return b.NewQuery(sql)
}

// TruncateTable creates a Query that can be used to truncate a table.
func (b *SqliteBuilder) TruncateTable(table string) *Query {
	sql := "DELETE FROM " + b.db.QuoteTableName(table)
	return b.NewQuery(sql)
}

// DropColumn creates a Query that can be used to drop a column from a table.
func (b *SqliteBuilder) DropColumn(table, col string) *Query {
	q := b.NewQuery("")
	q.LastError = errors.New("SQLite does not support dropping columns")
	return q
}

// RenameColumn creates a Query that can be used to rename a column in a table.
func (b *SqliteBuilder) RenameColumn(table, oldName, newName string) *Query {
	q := b.NewQuery("")
	q.LastError = errors.New("SQLite does not support renaming columns")
	return q
}

// AlterColumn creates a Query that can be used to change the definition of a table column.
func (b *SqliteBuilder) AlterColumn(table, col, typ string) *Query {
	q := b.NewQuery("")
	q.LastError = errors.New("SQLite does not support altering column")
	return q
}

// AddPrimaryKey creates a Query that can be used to specify primary key(s) for a table.
// The "name" parameter specifies the name of the primary key constraint.
func (b *SqliteBuilder) AddPrimaryKey(table, name string, cols ...string) *Query {
	q := b.NewQuery("")
	q.LastError = errors.New("SQLite does not support adding primary key")
	return q
}

// DropPrimaryKey creates a Query that can be used to remove the named primary key constraint from a table.
func (b *SqliteBuilder) DropPrimaryKey(table, name string) *Query {
	q := b.NewQuery("")
	q.LastError = errors.New("SQLite does not support dropping primary key")
	return q
}

// AddForeignKey creates a Query that can be used to add a foreign key constraint to a table.
// The length of cols and refCols must be the same as they refer to the primary and referential columns.
// The optional "options" parameters will be appended to the SQL statement. They can be used to
// specify options such as "ON DELETE CASCADE".
func (b *SqliteBuilder) AddForeignKey(table, name string, cols, refCols []string, refTable string, options ...string) *Query {
	q := b.NewQuery("")
	q.LastError = errors.New("SQLite does not support adding foreign keys")
	return q
}

// DropForeignKey creates a Query that can be used to remove the named foreign key constraint from a table.
func (b *SqliteBuilder) DropForeignKey(table, name string) *Query {
	q := b.NewQuery("")
	q.LastError = errors.New("SQLite does not support dropping foreign keys")
	return q
}
