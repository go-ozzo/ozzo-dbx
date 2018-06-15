// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package dbx provides a set of DB-agnostic and easy-to-use query building methods for relational databases.
package dbx

import (
	"bytes"
	"context"
	"database/sql"
	"regexp"
	"strings"
)

type (
	// LogFunc logs a message for each SQL statement being executed.
	// This method takes one or multiple parameters. If a single parameter
	// is provided, it will be treated as the log message. If multiple parameters
	// are provided, they will be passed to fmt.Sprintf() to generate the log message.
	LogFunc func(format string, a ...interface{})

	// PerfFunc is called when a query finishes execution.
	// The query execution time is passed to this function so that the DB performance
	// can be profiled. The "ns" parameter gives the number of nanoseconds that the
	// SQL statement takes to execute, while the "execute" parameter indicates whether
	// the SQL statement is executed or queried (usually SELECT statements).
	PerfFunc func(ns int64, sql string, execute bool)

	// BuilderFunc creates a Builder instance using the given DB instance and Executor.
	BuilderFunc func(*DB, Executor) Builder

	// DB enhances sql.DB by providing a set of DB-agnostic query building methods.
	// DB allows easier query building and population of data into Go variables.
	DB struct {
		Builder

		// FieldMapper maps struct fields to DB columns. Defaults to DefaultFieldMapFunc.
		FieldMapper FieldMapFunc
		// LogFunc logs the SQL statements being executed. Defaults to nil, meaning no logging.
		LogFunc LogFunc
		// PerfFunc logs the SQL execution time. Defaults to nil, meaning no performance profiling.
		PerfFunc PerfFunc

		sqlDB      *sql.DB
		driverName string
	}

	// Errors represents a list of errors.
	Errors []error
)

// BuilderFuncMap lists supported BuilderFunc according to DB driver names.
// You may modify this variable to add the builder support for a new DB driver.
// If a DB driver is not listed here, the StandardBuilder will be used.
var BuilderFuncMap = map[string]BuilderFunc{
	"sqlite3":  NewSqliteBuilder,
	"mysql":    NewMysqlBuilder,
	"postgres": NewPgsqlBuilder,
	"pgx":      NewPgsqlBuilder,
	"mssql":    NewMssqlBuilder,
	"oci8":     NewOciBuilder,
}

// NewFromDB encapsulates an existing database connection.
func NewFromDB(sqlDB *sql.DB, driverName string) *DB {
	db := &DB{
		driverName:  driverName,
		sqlDB:       sqlDB,
		FieldMapper: DefaultFieldMapFunc,
	}
	db.Builder = db.newBuilder(db.sqlDB)
	return db
}

// Open opens a database specified by a driver name and data source name (DSN).
// Note that Open does not check if DSN is specified correctly. It doesn't try to establish a DB connection either.
// Please refer to sql.Open() for more information.
func Open(driverName, dsn string) (*DB, error) {
	sqlDB, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}

	return NewFromDB(sqlDB, driverName), nil
}

// MustOpen opens a database and establishes a connection to it.
// Please refer to sql.Open() and sql.Ping() for more information.
func MustOpen(driverName, dsn string) (*DB, error) {
	db, err := Open(driverName, dsn)
	if err != nil {
		return nil, err
	}
	if err := db.sqlDB.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// Clone makes a shallow copy of DB.
func (db *DB) Clone() *DB {
	db2 := &DB{
		driverName:  db.driverName,
		sqlDB:       db.sqlDB,
		FieldMapper: db.FieldMapper,
		PerfFunc:    db.PerfFunc,
		LogFunc:     db.LogFunc,
	}
	db2.Builder = db2.newBuilder(db.sqlDB)
	return db2
}

// DB returns the sql.DB instance encapsulated by dbx.DB.
func (db *DB) DB() *sql.DB {
	return db.sqlDB
}

// Close closes the database, releasing any open resources.
// It is rare to Close a DB, as the DB handle is meant to be
// long-lived and shared between many goroutines.
func (db *DB) Close() error {
	return db.sqlDB.Close()
}

// Begin starts a transaction.
func (db *DB) Begin() (*Tx, error) {
	tx, err := db.sqlDB.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{db.newBuilder(tx), tx}, nil
}

// BeginTx starts a transaction with the given context and transaction options.
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.sqlDB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{db.newBuilder(tx), tx}, nil
}

// Wrap encapsulates an existing transaction.
func (db *DB) Wrap(sqlTx *sql.Tx) *Tx {
	return &Tx{db.newBuilder(sqlTx), sqlTx}
}

// Transactional starts a transaction and executes the given function.
// If the function returns an error, the transaction will be rolled back.
// Otherwise, the transaction will be committed.
func (db *DB) Transactional(f func(*Tx) error) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			if err2 := tx.Rollback(); err2 != nil {
				if err2 == sql.ErrTxDone {
					return
				}
				err = Errors{err, err2}
			}
		} else {
			if err = tx.Commit(); err == sql.ErrTxDone {
				err = nil
			}
		}
	}()

	err = f(tx)

	return err
}

// TransactionalContext starts a transaction and executes the given function with the given context and transaction options.
// If the function returns an error, the transaction will be rolled back.
// Otherwise, the transaction will be committed.
func (db *DB) TransactionalContext(ctx context.Context, opts *sql.TxOptions, f func(*Tx) error) (err error) {
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			if err2 := tx.Rollback(); err2 != nil {
				if err2 == sql.ErrTxDone {
					return
				}
				err = Errors{err, err2}
			}
		} else {
			if err = tx.Commit(); err == sql.ErrTxDone {
				err = nil
			}
		}
	}()

	err = f(tx)

	return err
}

// DriverName returns the name of the DB driver.
func (db *DB) DriverName() string {
	return db.driverName
}

// QuoteTableName quotes the given table name appropriately.
// If the table name contains DB schema prefix, it will be handled accordingly.
// This method will do nothing if the table name is already quoted or if it contains parenthesis.
func (db *DB) QuoteTableName(s string) string {
	if strings.Contains(s, "(") || strings.Contains(s, "{{") {
		return s
	}
	if !strings.Contains(s, ".") {
		return db.QuoteSimpleTableName(s)
	}
	parts := strings.Split(s, ".")
	for i, part := range parts {
		parts[i] = db.QuoteSimpleTableName(part)
	}
	return strings.Join(parts, ".")
}

// QuoteColumnName quotes the given column name appropriately.
// If the table name contains table name prefix, it will be handled accordingly.
// This method will do nothing if the column name is already quoted or if it contains parenthesis.
func (db *DB) QuoteColumnName(s string) string {
	if strings.Contains(s, "(") || strings.Contains(s, "{{") || strings.Contains(s, "[[") {
		return s
	}
	prefix := ""
	if pos := strings.LastIndex(s, "."); pos != -1 {
		prefix = db.QuoteTableName(s[:pos]) + "."
		s = s[pos+1:]
	}
	return prefix + db.QuoteSimpleColumnName(s)
}

var (
	plRegex    = regexp.MustCompile(`\{:\w+\}`)
	quoteRegex = regexp.MustCompile(`(\{\{[\w\-\. ]+\}\}|\[\[[\w\-\. ]+\]\])`)
)

// processSQL replaces the named param placeholders in the given SQL with anonymous ones.
// It also quotes table names and column names found in the SQL if these names are enclosed
// within double square/curly brackets. The method will return the updated SQL and the list of parameter names.
func (db *DB) processSQL(s string) (string, []string) {
	var placeholders []string
	count := 0
	s = plRegex.ReplaceAllStringFunc(s, func(m string) string {
		count++
		placeholders = append(placeholders, m[2:len(m)-1])
		return db.GeneratePlaceholder(count)
	})
	s = quoteRegex.ReplaceAllStringFunc(s, func(m string) string {
		if m[0] == '{' {
			return db.QuoteTableName(m[2: len(m)-2])
		}
		return db.QuoteColumnName(m[2: len(m)-2])
	})
	return s, placeholders
}

// newBuilder creates a query builder based on the current driver name.
func (db *DB) newBuilder(executor Executor) Builder {
	builderFunc, ok := BuilderFuncMap[db.driverName]
	if !ok {
		builderFunc = NewStandardBuilder
	}
	return builderFunc(db, executor)
}

// Error returns the error string of Errors.
func (errs Errors) Error() string {
	var b bytes.Buffer
	for i, e := range errs {
		if i > 0 {
			b.WriteRune('\n')
		}
		b.WriteString(e.Error())
	}
	return b.String()
}
