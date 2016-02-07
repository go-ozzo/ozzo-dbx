// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dbx

import (
	"fmt"
)

// OciBuilder is the builder for Oracle databases.
type OciBuilder struct {
	*BaseBuilder
	qb *OciQueryBuilder
}

var _ Builder = &OciBuilder{}

// OciQueryBuilder is the query builder for Oracle databases.
type OciQueryBuilder struct {
	*BaseQueryBuilder
}

// NewOciBuilder creates a new OciBuilder instance.
func NewOciBuilder(db *DB, executor Executor) Builder {
	return &OciBuilder{
		NewBaseBuilder(db, executor),
		&OciQueryBuilder{NewBaseQueryBuilder(db)},
	}
}

// GeneratePlaceholder generates an anonymous parameter placeholder with the given parameter ID.
func (b *OciBuilder) GeneratePlaceholder(i int) string {
	return fmt.Sprintf(":p%v", i)
}

// QueryBuilder returns the query builder supporting the current DB.
func (b *OciBuilder) QueryBuilder() QueryBuilder {
	return b.qb
}

// DropIndex creates a Query that can be used to remove the named index from a table.
func (b *OciBuilder) DropIndex(table, name string) *Query {
	sql := fmt.Sprintf("DROP INDEX %v", b.db.QuoteColumnName(name))
	return b.NewQuery(sql)
}

// RenameTable creates a Query that can be used to rename a table.
func (b *OciBuilder) RenameTable(oldName, newName string) *Query {
	sql := fmt.Sprintf("ALTER TABLE %v RENAME TO %v", b.db.QuoteTableName(oldName), b.db.QuoteTableName(newName))
	return b.NewQuery(sql)
}

// AlterColumn creates a Query that can be used to change the definition of a table column.
func (b *OciBuilder) AlterColumn(table, col, typ string) *Query {
	col = b.db.QuoteColumnName(col)
	sql := fmt.Sprintf("ALTER TABLE %v MODIFY %v %v", b.db.QuoteTableName(table), col, typ)
	return b.NewQuery(sql)
}

// BuildOrderByAndLimit generates the ORDER BY and LIMIT clauses.
func (q *OciQueryBuilder) BuildOrderByAndLimit(sql string, cols []string, limit int64, offset int64) string {
	if orderBy := q.BuildOrderBy(cols); orderBy != "" {
		sql += "\n" + orderBy
	}

	c := ""
	if offset > 0 {
		c = fmt.Sprintf("rowNumId > %v", offset)
	}
	if limit >= 0 {
		if c != "" {
			c += " AND "
		}
		c += fmt.Sprintf("rowNum <= %v", limit)
	}

	if c == "" {
		return sql
	}

	return `WITH USER_SQL AS (` + sql + `),
	PAGINATION AS (SELECT USER_SQL.*, rownum as rowNumId FROM USER_SQL)
SELECT * FROM PAGINATION WHERE ` + c
}
