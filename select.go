// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbx

import (
	"context"
	gosql "database/sql"
	"errors"
	"fmt"
	"reflect"
)

// SelectQuery represents a DB-agnostic SELECT query.
// It can be built into a DB-specific query by calling the Build() method.
type SelectQuery struct {
	// FieldMapper maps struct field names to DB column names.
	FieldMapper FieldMapFunc
	// TableMapper maps structs to DB table names.
	TableMapper TableMapFunc

	builder Builder
	ctx     context.Context

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
	lock         string
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
		ctx:         db.ctx,
		FieldMapper: db.FieldMapper,
		TableMapper: db.TableMapper,
	}
}

// Context returns the context associated with the query.
func (s *SelectQuery) Context() context.Context {
	return s.ctx
}

// WithContext associates a context with the query.
func (s *SelectQuery) WithContext(ctx context.Context) *SelectQuery {
	s.ctx = ctx
	return s
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
func (s *SelectQuery) Join(typ, table string, on Expression) *SelectQuery {
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

// ForUpdate makes the query a locking read: FOR UPDATE locks the selected rows exclusively until
// the end of the transaction. The clause is placed after ORDER BY, LIMIT and OFFSET.
//
// Row locks exist only inside a transaction: outside one, the statement commits at once and the
// lock is released immediately. So a locking read must be built from a Tx (tx.Select()); built
// from a DB, whose executor is a *sql.DB, Build sets LastError on every database, SQLite included.
// A locking read cannot be combined with UNION, GROUP BY or DISTINCT (Build sets LastError); avoid
// aggregate functions too. SQLite has no row locks (a write transaction locks the whole database),
// so inside a transaction the clause is left out there. SQL Server and Oracle are not supported
// (Build sets LastError).
// Of ForUpdate, ForShare and ForUpdateSkipLocked, the last one called wins.
func (s *SelectQuery) ForUpdate() *SelectQuery {
	s.lock = lockForUpdate
	return s
}

// ForShare makes the query a locking read with FOR SHARE: the selected rows are locked in shared
// mode until the end of the transaction, so other transactions can read but not modify them.
// It requires MySQL 8.0+ or PostgreSQL; MySQL 5.7 and MariaDB do not support FOR SHARE.
// See ForUpdate for the rules that apply to all locking reads.
func (s *SelectQuery) ForShare() *SelectQuery {
	s.lock = lockForShare
	return s
}

// ForUpdateSkipLocked makes the query a locking read with FOR UPDATE SKIP LOCKED: like ForUpdate,
// but rows locked by other transactions are skipped instead of waited for. It is meant for job
// queues, where each worker claims rows the others have not taken.
// It requires MySQL 8.0+ or PostgreSQL 9.5+. See ForUpdate for the rules that apply to all
// locking reads.
func (s *SelectQuery) ForUpdateSkipLocked() *SelectQuery {
	s.lock = lockForUpdateSkipLocked
	return s
}

// The locking clauses set by ForUpdate, ForShare and ForUpdateSkipLocked.
const (
	lockForUpdate           = "FOR UPDATE"
	lockForShare            = "FOR SHARE"
	lockForUpdateSkipLocked = "FOR UPDATE SKIP LOCKED"
)

// lockMethods names the SelectQuery method of each locking clause, for error messages.
var lockMethods = map[string]string{
	lockForUpdate:           "ForUpdate",
	lockForShare:            "ForShare",
	lockForUpdateSkipLocked: "ForUpdateSkipLocked",
}

// buildLock appends the locking clause to sql, the SELECT statement with its ORDER BY and LIMIT.
func (s *SelectQuery) buildLock(qb QueryBuilder, sql string) (string, error) {
	// PostgreSQL rejects these combinations; one rule for every database keeps queries portable
	switch {
	case len(s.union) > 0:
		return sql, errors.New("a locking read cannot be combined with UNION")
	case len(s.groupBy) > 0:
		return sql, errors.New("a locking read cannot be combined with GROUP BY")
	case s.distinct:
		return sql, errors.New("a locking read cannot be combined with DISTINCT")
	}
	// Outside a transaction the lock is released as soon as the statement ends, so it protects
	// nothing (two job-queue workers could claim the same row). Rejected on every database,
	// SQLite included, so that the bug surfaces in SQLite-backed test suites too. Only a plain
	// *sql.DB is rejected, so custom or instrumented executors never get a false positive.
	if eb, ok := s.builder.(interface{ Executor() Executor }); ok {
		if _, ok := eb.Executor().(*gosql.DB); ok {
			return sql, fmt.Errorf("%v requires a transaction: outside one the lock is released when the statement ends", lockMethods[s.lock])
		}
	}
	if _, ok := s.builder.(*SqliteBuilder); ok {
		return sql, nil
	}
	lqb, ok := qb.(LockQueryBuilder)
	if !ok {
		return sql, fmt.Errorf("the query builder %T does not support locking reads", qb)
	}
	return lqb.BuildLock(sql, s.lock)
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
	var lockErr error
	if s.lock != "" {
		sql, lockErr = s.buildLock(qb, sql)
	}
	if union := qb.BuildUnion(s.union, params); union != "" {
		sql = fmt.Sprintf("(%v) %v", sql, union)
	}

	q := s.builder.NewQuery(sql).Bind(params)
	if lockErr != nil {
		q.LastError = lockErr
	}
	return q
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
		if tableName := s.TableMapper(a); tableName != "" {
			s.from = []string{tableName}
		}
	}
	return s.Build().WithContext(s.ctx).One(a)
}

// Model selects the row with the specified primary key and populates the model with the row data.
//
// The model variable should be a pointer to a struct. If the query does not specify a "from" clause,
// it will use the model struct to determine which table to select data from. It will also use the model
// to infer the name of the primary key column. Only simple primary key is supported. For composite primary keys,
// please use Where() to specify the filtering condition.
func (s *SelectQuery) Model(pk, model interface{}) error {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr { //nolint:govet // reflect.Ptr is valid, govet toolchain mismatch
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
		if tableName := s.TableMapper(slice); tableName != "" {
			s.from = []string{tableName}
		}
	}
	return s.Build().WithContext(s.ctx).All(slice)
}

// Rows builds and executes the SELECT query and returns a Rows object for data retrieval purpose.
// This is a shortcut to SelectQuery.Build().Rows().
func (s *SelectQuery) Rows() (*Rows, error) {
	return s.Build().WithContext(s.ctx).Rows()
}

// Row builds and executes the SELECT query and populates the first row of the result into the specified variables.
// This is a shortcut to SelectQuery.Build().Row().
func (s *SelectQuery) Row(a ...interface{}) error {
	return s.Build().WithContext(s.ctx).Row(a...)
}

// Column builds and executes the SELECT statement and populates the first column of the result into a slice.
// Note that the parameter must be a pointer to a slice.
// This is a shortcut to SelectQuery.Build().Column().
func (s *SelectQuery) Column(a interface{}) error {
	return s.Build().WithContext(s.ctx).Column(a)
}
