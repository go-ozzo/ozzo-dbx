// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbx

import (
	"fmt"
	"strings"
)

// MssqlBuilder is the builder for SQL Server databases.
type MssqlBuilder struct {
	*BaseBuilder
	qb *MssqlQueryBuilder
}

var _ Builder = &MssqlBuilder{}

// MssqlQueryBuilder is the query builder for SQL Server databases.
type MssqlQueryBuilder struct {
	*BaseQueryBuilder
}

// NewMssqlBuilder creates a new MssqlBuilder instance.
func NewMssqlBuilder(db *DB, executor Executor) Builder {
	return &MssqlBuilder{
		NewBaseBuilder(db, executor),
		&MssqlQueryBuilder{NewBaseQueryBuilder(db)},
	}
}

// QueryBuilder returns the query builder supporting the current DB.
func (b *MssqlBuilder) QueryBuilder() QueryBuilder {
	return b.qb
}

// Select returns a new SelectQuery object that can be used to build a SELECT statement.
// The parameters to this method should be the list column names to be selected.
// A column name may have an optional alias name. For example, Select("id", "my_name AS name").
func (b *MssqlBuilder) Select(cols ...string) *SelectQuery {
	return NewSelectQuery(b, b.db).Select(cols...)
}

// Model returns a new ModelQuery object that can be used to perform model-based DB operations.
// The model passed to this method should be a pointer to a model struct.
func (b *MssqlBuilder) Model(model interface{}) *ModelQuery {
	return NewModelQuery(model, b.db.FieldMapper, b.db, b)
}

// QuoteSimpleTableName quotes a simple table name.
// A simple table name does not contain any schema prefix.
func (b *MssqlBuilder) QuoteSimpleTableName(s string) string {
	if strings.Contains(s, `[`) {
		return s
	}
	return `[` + s + `]`
}

// QuoteSimpleColumnName quotes a simple column name.
// A simple column name does not contain any table prefix.
func (b *MssqlBuilder) QuoteSimpleColumnName(s string) string {
	if strings.Contains(s, `[`) || s == "*" {
		return s
	}
	return `[` + s + `]`
}

// RenameTable creates a Query that can be used to rename a table.
func (b *MssqlBuilder) RenameTable(oldName, newName string) *Query {
	sql := fmt.Sprintf("sp_name '%v', '%v'", oldName, newName)
	return b.NewQuery(sql)
}

// RenameColumn creates a Query that can be used to rename a column in a table.
func (b *MssqlBuilder) RenameColumn(table, oldName, newName string) *Query {
	sql := fmt.Sprintf("sp_name '%v.%v', '%v', 'COLUMN'", table, oldName, newName)
	return b.NewQuery(sql)
}

// AlterColumn creates a Query that can be used to change the definition of a table column.
func (b *MssqlBuilder) AlterColumn(table, col, typ string) *Query {
	col = b.db.QuoteColumnName(col)
	sql := fmt.Sprintf("ALTER TABLE %v ALTER COLUMN %v %v", b.db.QuoteTableName(table), col, typ)
	return b.NewQuery(sql)
}

// BuildOrderByAndLimit generates the ORDER BY and LIMIT clauses.
func (q *MssqlQueryBuilder) BuildOrderByAndLimit(sql string, cols []string, limit int64, offset int64) string {
	orderBy := q.BuildOrderBy(cols)
	if limit < 0 && offset < 0 {
		if orderBy == "" {
			return sql
		}
		return sql + "\n" + orderBy
	}

	// only SQL SERVER 2012 or newer are supported by this method

	if orderBy == "" {
		// ORDER BY clause is required when FETCH and OFFSET are in the SQL
		orderBy = "ORDER BY (SELECT NULL)"
	}
	sql += "\n" + orderBy

	// http://technet.microsoft.com/en-us/library/gg699618.aspx
	if offset < 0 {
		offset = 0
	}
	sql += "\n" + fmt.Sprintf("OFFSET %v ROWS", offset)
	if limit >= 0 {
		sql += "\n" + fmt.Sprintf("FETCH NEXT %v ROWS ONLY", limit)
	}
	return sql
}
