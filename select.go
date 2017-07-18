// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbx

import (
	"fmt"
	"reflect"
)

// SelectQuery represents a DB-agnostic SELECT query.
// It can be built into a DB-specific query by calling the Build() method.
type SelectQuery struct {
	// FieldMapper maps struct field names to DB column names.
	FieldMapper FieldMapFunc

	builder Builder

	selects      []string
	distinct     bool
	selectOption string
	from         []string
	where        Expression
	join         []JoinInfo
	orderBy      []string
	groupBy      []string
	having       Expression
	union        []UnionInfo
	limit        int64
	offset       int64
	params       Params
}

// JoinInfo contains the specification for a JOIN clause.
type JoinInfo struct {
	Join  string
	Table string
	On    Expression
}

// UnionInfo contains the specification for a UNION clause.
type UnionInfo struct {
	All   bool
	Query *Query
}

// NewSelectQuery creates a new SelectQuery instance.
func NewSelectQuery(builder Builder, db *DB) *SelectQuery {
	return &SelectQuery{
		builder:     builder,
		selects:     []string{},
		from:        []string{},
		join:        []JoinInfo{},
		orderBy:     []string{},
		groupBy:     []string{},
		union:       []UnionInfo{},
		limit:       -1,
		params:      Params{},
		FieldMapper: db.FieldMapper,
	}
}

// Select specifies the columns to be selected.
// Column names will be automatically quoted.
func (s *SelectQuery) Select(cols ...string) *SelectQuery {
	s.selects = cols
	return s
}

// AndSelect adds additional columns to be selected.
// Column names will be automatically quoted.
func (s *SelectQuery) AndSelect(cols ...string) *SelectQuery {
	s.selects = append(s.selects, cols...)
	return s
}

// Distinct specifies whether to select columns distinctively.
// By default, distinct is false.
func (s *SelectQuery) Distinct(v bool) *SelectQuery {
	s.distinct = v
	return s
}

// SelectOption specifies additional option that should be append to "SELECT".
func (s *SelectQuery) SelectOption(option string) *SelectQuery {
	s.selectOption = option
	return s
}

// From specifies which tables to select from.
// Table names will be automatically quoted.
func (s *SelectQuery) From(tables ...string) *SelectQuery {
	s.from = tables
	return s
}

// Where specifies the WHERE condition.
func (s *SelectQuery) Where(e Expression) *SelectQuery {
	s.where = e
	return s
}

// AndWhere concatenates a new WHERE condition with the existing one (if any) using "AND".
func (s *SelectQuery) AndWhere(e Expression) *SelectQuery {
	s.where = And(s.where, e)
	return s
}

// OrWhere concatenates a new WHERE condition with the existing one (if any) using "OR".
func (s *SelectQuery) OrWhere(e Expression) *SelectQuery {
	s.where = Or(s.where, e)
	return s
}

// Join specifies a JOIN clause.
// The "typ" parameter specifies the JOIN type (e.g. "INNER JOIN", "LEFT JOIN").
func (s *SelectQuery) Join(typ string, table string, on Expression) *SelectQuery {
	s.join = append(s.join, JoinInfo{typ, table, on})
	return s
}

// InnerJoin specifies an INNER JOIN clause.
// This is a shortcut method for Join.
func (s *SelectQuery) InnerJoin(table string, on Expression) *SelectQuery {
	return s.Join("INNER JOIN", table, on)
}

// LeftJoin specifies a LEFT JOIN clause.
// This is a shortcut method for Join.
func (s *SelectQuery) LeftJoin(table string, on Expression) *SelectQuery {
	return s.Join("LEFT JOIN", table, on)
}

// RightJoin specifies a RIGHT JOIN clause.
// This is a shortcut method for Join.
func (s *SelectQuery) RightJoin(table string, on Expression) *SelectQuery {
	return s.Join("RIGHT JOIN", table, on)
}

// OrderBy specifies the ORDER BY clause.
// Column names will be properly quoted. A column name can contain "ASC" or "DESC" to indicate its ordering direction.
func (s *SelectQuery) OrderBy(cols ...string) *SelectQuery {
	s.orderBy = cols
	return s
}

// AndOrderBy appends additional columns to the existing ORDER BY clause.
// Column names will be properly quoted. A column name can contain "ASC" or "DESC" to indicate its ordering direction.
func (s *SelectQuery) AndOrderBy(cols ...string) *SelectQuery {
	s.orderBy = append(s.orderBy, cols...)
	return s
}

// GroupBy specifies the GROUP BY clause.
// Column names will be properly quoted.
func (s *SelectQuery) GroupBy(cols ...string) *SelectQuery {
	s.groupBy = cols
	return s
}

// AndGroupBy appends additional columns to the existing GROUP BY clause.
// Column names will be properly quoted.
func (s *SelectQuery) AndGroupBy(cols ...string) *SelectQuery {
	s.groupBy = append(s.groupBy, cols...)
	return s
}

// Having specifies the HAVING clause.
func (s *SelectQuery) Having(e Expression) *SelectQuery {
	s.having = e
	return s
}

// AndHaving concatenates a new HAVING condition with the existing one (if any) using "AND".
func (s *SelectQuery) AndHaving(e Expression) *SelectQuery {
	s.having = And(s.having, e)
	return s
}

// OrHaving concatenates a new HAVING condition with the existing one (if any) using "OR".
func (s *SelectQuery) OrHaving(e Expression) *SelectQuery {
	s.having = Or(s.having, e)
	return s
}

// Union specifies a UNION clause.
func (s *SelectQuery) Union(q *Query) *SelectQuery {
	s.union = append(s.union, UnionInfo{false, q})
	return s
}

// UnionAll specifies a UNION ALL clause.
func (s *SelectQuery) UnionAll(q *Query) *SelectQuery {
	s.union = append(s.union, UnionInfo{true, q})
	return s
}

// Limit specifies the LIMIT clause.
// A negative limit means no limit.
func (s *SelectQuery) Limit(limit int64) *SelectQuery {
	s.limit = limit
	return s
}

// Offset specifies the OFFSET clause.
// A negative offset means no offset.
func (s *SelectQuery) Offset(offset int64) *SelectQuery {
	s.offset = offset
	return s
}

// Bind specifies the parameter values to be bound to the query.
func (s *SelectQuery) Bind(params Params) *SelectQuery {
	s.params = params
	return s
}

// AndBind appends additional parameters to be bound to the query.
func (s *SelectQuery) AndBind(params Params) *SelectQuery {
	if len(s.params) == 0 {
		s.params = params
	} else {
		for k, v := range params {
			s.params[k] = v
		}
	}
	return s
}

// Build builds the SELECT query and returns an executable Query object.
func (s *SelectQuery) Build() *Query {
	params := Params{}
	for k, v := range s.params {
		params[k] = v
	}

	qb := s.builder.QueryBuilder()

	clauses := []string{
		qb.BuildSelect(s.selects, s.distinct, s.selectOption),
		qb.BuildFrom(s.from),
		qb.BuildJoin(s.join, params),
		qb.BuildWhere(s.where, params),
		qb.BuildGroupBy(s.groupBy),
		qb.BuildHaving(s.having, params),
	}
	sql := ""
	for _, clause := range clauses {
		if clause != "" {
			if sql == "" {
				sql = clause
			} else {
				sql += " " + clause
			}
		}
	}
	sql = qb.BuildOrderByAndLimit(sql, s.orderBy, s.limit, s.offset)
	if union := qb.BuildUnion(s.union, params); union != "" {
		sql = fmt.Sprintf("(%v) %v", sql, union)
	}

	return s.builder.NewQuery(sql).Bind(params)
}

// One executes the SELECT query and populates the first row of the result into the specified variable.
//
// If the query does not specify a "from" clause, the method will try to infer the name of the table
// to be selected from by calling getTableName() which will return either the variable type name
// or the TableName() method if the variable implements the TableModel interface.
//
// Note that when the query has no rows in the result set, an sql.ErrNoRows will be returned.
func (s *SelectQuery) One(a interface{}) error {
	if len(s.from) == 0 {
		if tableName := GetTableName(a); tableName != "" {
			s.from = []string{tableName}
		}
	}
	return s.Build().One(a)
}

// Model selects the row with the specified primary key and populates the model with the row data.
//
// The model variable should be a pointer to a struct. If the query does not specify a "from" clause,
// it will use the model struct to determine which table to select data from. It will also use the model
// to infer the name of the primary key column. Only simple primary key is supported. For composite primary keys,
// please use Where() to specify the filtering condition.
func (s *SelectQuery) Model(pk, model interface{}) error {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return VarTypeError("must be a pointer to a struct")
	}
	si := getStructInfo(t, s.FieldMapper)
	if len(si.pkNames) == 1 {
		return s.AndWhere(HashExp{si.nameMap[si.pkNames[0]].dbName: pk}).One(model)
	}

	if len(si.pkNames) == 0 {
		return MissingPKError
	}
	return CompositePKError
}

// All executes the SELECT query and populates all rows of the result into a slice.
//
// Note that the slice must be passed in as a pointer.
//
// If the query does not specify a "from" clause, the method will try to infer the name of the table
// to be selected from by calling getTableName() which will return either the type name of the slice elements
// or the TableName() method if the slice element implements the TableModel interface.
func (s *SelectQuery) All(slice interface{}) error {
	if len(s.from) == 0 {
		if tableName := GetTableName(slice); tableName != "" {
			s.from = []string{tableName}
		}
	}
	return s.Build().All(slice)
}

// Rows builds and executes the SELECT query and returns a Rows object for data retrieval purpose.
// This is a shortcut to SelectQuery.Build().Rows()
func (s *SelectQuery) Rows() (*Rows, error) {
	return s.Build().Rows()
}

// Row builds and executes the SELECT query and populates the first row of the result into the specified variables.
// This is a shortcut to SelectQuery.Build().Row()
func (s *SelectQuery) Row(a ...interface{}) error {
	return s.Build().Row(a...)
}

// Column builds and executes the SELECT statement and populates the first column of the result into a slice.
// Note that the parameter must be a pointer to a slice.
// This is a shortcut to SelectQuery.Build().Column()
func (s *SelectQuery) Column(a interface{}) error {
	return s.Build().Column(a)
}
