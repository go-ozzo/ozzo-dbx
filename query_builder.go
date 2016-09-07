// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbx

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// QueryBuilder builds different clauses for a SELECT SQL statement.
type QueryBuilder interface {
	// BuildSelect generates a SELECT clause from the given selected column names.
	BuildSelect(cols []string, distinct bool, option string) string
	// BuildFrom generates a FROM clause from the given tables.
	BuildFrom(tables []string) string
	// BuildGroupBy generates a GROUP BY clause from the given group-by columns.
	BuildGroupBy(cols []string) string
	// BuildJoin generates a JOIN clause from the given join information.
	BuildJoin([]JoinInfo, Params) string
	// BuildWhere generates a WHERE clause from the given expression.
	BuildWhere(Expression, Params) string
	// BuildHaving generates a HAVING clause from the given expression.
	BuildHaving(Expression, Params) string
	// BuildOrderByAndLimit generates the ORDER BY and LIMIT clauses.
	BuildOrderByAndLimit(string, []string, int64, int64) string
	// BuildUnion generates a UNION clause from the given union information.
	BuildUnion([]UnionInfo, Params) string
}

// BaseQueryBuilder provides a basic implementation of QueryBuilder.
type BaseQueryBuilder struct {
	db *DB
}

var _ QueryBuilder = &BaseQueryBuilder{}

// NewBaseQueryBuilder creates a new BaseQueryBuilder instance.
func NewBaseQueryBuilder(db *DB) *BaseQueryBuilder {
	return &BaseQueryBuilder{db}
}

// DB returns the DB instance associated with the query builder.
func (q *BaseQueryBuilder) DB() *DB {
	return q.db
}

// the regexp for columns and tables.
var selectRegex = regexp.MustCompile(`(?i:\s+as\s+|\s+)([\w\-_\.]+)$`)

// BuildSelect generates a SELECT clause from the given selected column names.
func (q *BaseQueryBuilder) BuildSelect(cols []string, distinct bool, option string) string {
	var s bytes.Buffer
	s.WriteString("SELECT ")
	if distinct {
		s.WriteString("DISTINCT ")
	}
	if option != "" {
		s.WriteString(option)
		s.WriteString(" ")
	}
	if len(cols) == 0 {
		s.WriteString("*")
		return s.String()
	}

	for i, col := range cols {
		if i > 0 {
			s.WriteString(", ")
		}
		matches := selectRegex.FindStringSubmatch(col)
		if len(matches) == 0 {
			s.WriteString(q.db.QuoteColumnName(col))
		} else {
			col := col[:len(col)-len(matches[0])]
			alias := matches[1]
			s.WriteString(q.db.QuoteColumnName(col) + " AS " + q.db.QuoteSimpleColumnName(alias))
		}
	}

	return s.String()
}

// BuildFrom generates a FROM clause from the given tables.
func (q *BaseQueryBuilder) BuildFrom(tables []string) string {
	if len(tables) == 0 {
		return ""
	}
	s := ""
	for _, table := range tables {
		table = q.quoteTableNameAndAlias(table)
		if s == "" {
			s = table
		} else {
			s += ", " + table
		}
	}
	return "FROM " + s
}

// BuildJoin generates a JOIN clause from the given join information.
func (q *BaseQueryBuilder) BuildJoin(joins []JoinInfo, params Params) string {
	if len(joins) == 0 {
		return ""
	}
	parts := []string{}
	for _, join := range joins {
		sql := join.Join + " " + q.quoteTableNameAndAlias(join.Table)
		on := ""
		if join.On != nil {
			on = join.On.Build(q.db, params)
		}
		if on != "" {
			sql += " ON " + on
		}
		parts = append(parts, sql)
	}
	return strings.Join(parts, " ")
}

// BuildWhere generates a WHERE clause from the given expression.
func (q *BaseQueryBuilder) BuildWhere(e Expression, params Params) string {
	if e != nil {
		if c := e.Build(q.db, params); c != "" {
			return "WHERE " + c
		}
	}
	return ""
}

// BuildHaving generates a HAVING clause from the given expression.
func (q *BaseQueryBuilder) BuildHaving(e Expression, params Params) string {
	if e != nil {
		if c := e.Build(q.db, params); c != "" {
			return "HAVING " + c
		}
	}
	return ""
}

// BuildGroupBy generates a GROUP BY clause from the given group-by columns.
func (q *BaseQueryBuilder) BuildGroupBy(cols []string) string {
	if len(cols) == 0 {
		return ""
	}
	s := ""
	for i, col := range cols {
		if i == 0 {
			s = q.db.QuoteColumnName(col)
		} else {
			s += ", " + q.db.QuoteColumnName(col)
		}
	}
	return "GROUP BY " + s
}

// BuildOrderByAndLimit generates the ORDER BY and LIMIT clauses.
func (q *BaseQueryBuilder) BuildOrderByAndLimit(sql string, cols []string, limit int64, offset int64) string {
	if orderBy := q.BuildOrderBy(cols); orderBy != "" {
		sql += " " + orderBy
	}
	if limit := q.BuildLimit(limit, offset); limit != "" {
		return sql + " " + limit
	}
	return sql
}

// BuildUnion generates a UNION clause from the given union information.
func (q *BaseQueryBuilder) BuildUnion(unions []UnionInfo, params Params) string {
	if len(unions) == 0 {
		return ""
	}
	sql := ""
	for i, union := range unions {
		if i > 0 {
			sql += " "
		}
		for k, v := range union.Query.params {
			params[k] = v
		}
		u := "UNION"
		if union.All {
			u = "UNION ALL"
		}
		sql += fmt.Sprintf("%v (%v)", u, union.Query.sql)
	}
	return sql
}

var orderRegex = regexp.MustCompile(`\s+((?i)ASC|DESC)$`)

// BuildOrderBy generates the ORDER BY clause.
func (q *BaseQueryBuilder) BuildOrderBy(cols []string) string {
	if len(cols) == 0 {
		return ""
	}
	s := ""
	for i, col := range cols {
		if i > 0 {
			s += ", "
		}
		matches := orderRegex.FindStringSubmatch(col)
		if len(matches) == 0 {
			s += q.db.QuoteColumnName(col)
		} else {
			col := col[:len(col)-len(matches[0])]
			dir := matches[1]
			s += q.db.QuoteColumnName(col) + " " + dir
		}
	}
	return "ORDER BY " + s
}

// BuildLimit generates the LIMIT clause.
func (q *BaseQueryBuilder) BuildLimit(limit int64, offset int64) string {
	if limit < 0 && offset > 0 {
		// most DBMS requires LIMIT when OFFSET is present
		limit = 9223372036854775807 // 2^63 - 1
	}

	sql := ""
	if limit >= 0 {
		sql = fmt.Sprintf("LIMIT %v", limit)
	}
	if offset <= 0 {
		return sql
	}
	if sql != "" {
		sql += " "
	}
	return sql + fmt.Sprintf("OFFSET %v", offset)
}

func (q *BaseQueryBuilder) quoteTableNameAndAlias(table string) string {
	matches := selectRegex.FindStringSubmatch(table)
	if len(matches) == 0 {
		return q.db.QuoteTableName(table)
	}
	table = table[:len(table)-len(matches[0])]
	return q.db.QuoteTableName(table) + " " + q.db.QuoteSimpleTableName(matches[1])
}
