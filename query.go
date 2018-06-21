// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbx

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Params represents a list of parameter values to be bound to a SQL statement.
// The map keys are the parameter names while the map values are the corresponding parameter values.
type Params map[string]interface{}

// Executor prepares, executes, or queries a SQL statement.
type Executor interface {
	// Exec executes a SQL statement
	Exec(query string, args ...interface{}) (sql.Result, error)
	// ExecContext executes a SQL statement with the given context
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	// Query queries a SQL statement
	Query(query string, args ...interface{}) (*sql.Rows, error)
	// QueryContext queries a SQL statement with the given context
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	// Prepare creates a prepared statement
	Prepare(query string) (*sql.Stmt, error)
}

// Query represents a SQL statement to be executed.
type Query struct {
	executor Executor

	sql, rawSQL  string
	placeholders []string
	params       Params

	stmt *sql.Stmt
	ctx  context.Context

	// FieldMapper maps struct field names to DB column names.
	FieldMapper FieldMapFunc
	// LastError contains the last error (if any) of the query.
	// LastError is cleared by Execute(), Row(), Rows(), One(), and All().
	LastError error
	// LogFunc is used to log the SQL statement being executed.
	LogFunc LogFunc
	// PerfFunc is used to log the SQL execution time. It is ignored if nil.
	PerfFunc PerfFunc
}

// NewQuery creates a new Query with the given SQL statement.
func NewQuery(db *DB, executor Executor, sql string) *Query {
	rawSQL, placeholders := db.processSQL(sql)
	return &Query{
		executor:     executor,
		sql:          sql,
		rawSQL:       rawSQL,
		placeholders: placeholders,
		params:       Params{},
		FieldMapper:  db.FieldMapper,
		LogFunc:      db.LogFunc,
		PerfFunc:     db.PerfFunc,
	}
}

// SQL returns the original SQL used to create the query.
// The actual SQL (RawSQL) being executed is obtained by replacing the named
// parameter placeholders with anonymous ones.
func (q *Query) SQL() string {
	return q.sql
}

// Context returns the context associated with the query.
func (q *Query) Context() context.Context {
	return q.ctx
}

// WithContext associates a context with the query.
func (q *Query) WithContext(ctx context.Context) *Query {
	q.ctx = ctx
	return q
}

// logSQL returns the SQL statement with parameters being replaced with the actual values.
// The result is only for logging purpose and should not be used to execute.
func (q *Query) logSQL() string {
	s := q.sql
	for k, v := range q.params {
		if valuer, ok := v.(driver.Valuer); ok && valuer != nil {
			v, _ = valuer.Value()
		}
		var sv string
		if str, ok := v.(string); ok {
			sv = "'" + strings.Replace(str, "'", "''", -1) + "'"
		} else if bs, ok := v.([]byte); ok {
			sv = "'" + strings.Replace(string(bs), "'", "''", -1) + "'"
		} else {
			sv = fmt.Sprintf("%v", v)
		}
		s = strings.Replace(s, "{:"+k+"}", sv, -1)
	}
	return s
}

// log logs a message for the currently executed SQL statement.
func (q *Query) log(start time.Time, execute bool) {
	if q.LogFunc == nil && q.PerfFunc == nil {
		return
	}
	ns := time.Now().Sub(start).Nanoseconds()
	s := q.logSQL()
	if q.LogFunc != nil {
		if execute {
			q.LogFunc("[%.2fms] Execute SQL: %v", float64(ns)/1e6, s)
		} else {
			q.LogFunc("[%.2fms] Query SQL: %v", float64(ns)/1e6, s)
		}
	}
	if q.PerfFunc != nil {
		q.PerfFunc(ns, s, execute)
	}
}

// Params returns the parameters to be bound to the SQL statement represented by this query.
func (q *Query) Params() Params {
	return q.params
}

// Prepare creates a prepared statement for later queries or executions.
// Close() should be called after finishing all queries.
func (q *Query) Prepare() *Query {
	stmt, err := q.executor.Prepare(q.rawSQL)
	if err != nil {
		q.LastError = err
		return q
	}
	q.stmt = stmt
	return q
}

// Close closes the underlying prepared statement.
// Close does nothing if the query has not been prepared before.
func (q *Query) Close() error {
	if q.stmt == nil {
		return nil
	}

	err := q.stmt.Close()
	q.stmt = nil
	return err
}

// Bind sets the parameters that should be bound to the SQL statement.
// The parameter placeholders in the SQL statement are in the format of "{:ParamName}".
func (q *Query) Bind(params Params) *Query {
	if len(q.params) == 0 {
		q.params = params
	} else {
		for k, v := range params {
			q.params[k] = v
		}
	}
	return q
}

// Execute executes the SQL statement without retrieving data.
func (q *Query) Execute() (result sql.Result, err error) {
	err = q.LastError
	q.LastError = nil
	if err != nil {
		return
	}

	var params []interface{}
	params, err = replacePlaceholders(q.placeholders, q.params)
	if err != nil {
		return
	}

	defer q.log(time.Now(), true)

	if q.ctx == nil {
		if q.stmt == nil {
			result, err = q.executor.Exec(q.rawSQL, params...)
		} else {
			result, err = q.stmt.Exec(params...)
		}
	} else {
		if q.stmt == nil {
			result, err = q.executor.ExecContext(q.ctx, q.rawSQL, params...)
		} else {
			result, err = q.stmt.ExecContext(q.ctx, params...)
		}
	}
	return
}

// One executes the SQL statement and populates the first row of the result into a struct or NullStringMap.
// Refer to Rows.ScanStruct() and Rows.ScanMap() for more details on how to specify
// the variable to be populated.
// Note that when the query has no rows in the result set, an sql.ErrNoRows will be returned.
func (q *Query) One(a interface{}) error {
	rows, err := q.Rows()
	if err != nil {
		return err
	}
	return rows.one(a)
}

// All executes the SQL statement and populates all the resulting rows into a slice of struct or NullStringMap.
// The slice must be given as a pointer. Each slice element must be either a struct or a NullStringMap.
// Refer to Rows.ScanStruct() and Rows.ScanMap() for more details on how each slice element can be.
// If the query returns no row, the slice will be an empty slice (not nil).
func (q *Query) All(slice interface{}) error {
	rows, err := q.Rows()
	if err != nil {
		return err
	}
	return rows.all(slice)
}

// Row executes the SQL statement and populates the first row of the result into a list of variables.
// Note that the number of the variables should match to that of the columns in the query result.
// Note that when the query has no rows in the result set, an sql.ErrNoRows will be returned.
func (q *Query) Row(a ...interface{}) error {
	rows, err := q.Rows()
	if err != nil {
		return err
	}
	return rows.row(a...)
}

// Column executes the SQL statement and populates the first column of the result into a slice.
// Note that the parameter must be a pointer to a slice.
func (q *Query) Column(a interface{}) error {
	rows, err := q.Rows()
	if err != nil {
		return err
	}
	return rows.column(a)
}

// Rows executes the SQL statement and returns a Rows object to allow retrieving data row by row.
func (q *Query) Rows() (rows *Rows, err error) {
	err = q.LastError
	q.LastError = nil
	if err != nil {
		return
	}

	var params []interface{}
	params, err = replacePlaceholders(q.placeholders, q.params)
	if err != nil {
		return
	}

	defer q.log(time.Now(), false)

	var rr *sql.Rows
	if q.ctx == nil {
		if q.stmt == nil {
			rr, err = q.executor.Query(q.rawSQL, params...)
		} else {
			rr, err = q.stmt.Query(params...)
		}
	} else {
		if q.stmt == nil {
			rr, err = q.executor.QueryContext(q.ctx, q.rawSQL, params...)
		} else {
			rr, err = q.stmt.QueryContext(q.ctx, params...)
		}
	}
	rows = &Rows{rr, q.FieldMapper}
	return
}

// replacePlaceholders converts a list of named parameters into a list of anonymous parameters.
func replacePlaceholders(placeholders []string, params Params) ([]interface{}, error) {
	if len(placeholders) == 0 {
		return nil, nil
	}

	var result []interface{}
	for _, name := range placeholders {
		if value, ok := params[name]; ok {
			result = append(result, value)
		} else {
			return nil, errors.New("Named parameter not found: " + name)
		}
	}
	return result, nil
}
