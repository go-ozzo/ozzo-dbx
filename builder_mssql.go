// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
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

func (b *MssqlBuilder) QueryBuilder() QueryBuilder {
	return b.qb
}

func (b *MssqlBuilder) QuoteSimpleTableName(s string) string {
	if strings.Contains(s, `[`) {
		return s
	}
	return `[` + s + `]`
}

func (b *MssqlBuilder) QuoteSimpleColumnName(s string) string {
	if strings.Contains(s, `[`) || s == "*" {
		return s
	}
	return `[` + s + `]`
}

func (b *MssqlBuilder) RenameTable(oldName, newName string) *Query {
	sql := fmt.Sprintf("sp_name '%v', '%v'", oldName, newName)
	return b.NewQuery(sql)
}

func (b *MssqlBuilder) RenameColumn(table, oldName, newName string) *Query {
	sql := fmt.Sprintf("sp_name '%v.%v', '%v', 'COLUMN'", table, oldName, newName)
	return b.NewQuery(sql)
}

func (b *MssqlBuilder) AlterColumn(table, col, typ string) *Query {
	col = b.db.QuoteColumnName(col)
	sql := fmt.Sprintf("ALTER TABLE %v ALTER COLUMN %v %v", b.db.QuoteTableName(table), col, typ)
	return b.NewQuery(sql)
}


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
