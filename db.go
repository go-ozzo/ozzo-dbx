// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package dbx provides a set of DB-agnostic and easy-to-use query building methods for relational databases.
package dbx

import (
	"database/sql"
	"regexp"
	"strings"
)

// LogFunc logs a message for each SQL statement being executed.
// This method takes one or multiple parameters. If a single parameter
// is provided, it will be treated as the log message. If multiple parameters
// are provided, they will be passed to fmt.Sprintf() to generate the log message.
type LogFunc func(format string, a ...interface{})

// DB enhances sql.DB by providing a set of DB-agnostic query building methods.
// DB allows easier query building and population of data into Go variables.
type DB struct {
	Builder

	// FieldMapper maps struct fields to DB columns. Defaults to DefaultFieldMapFunc.
	FieldMapper FieldMapFunc
	// LogFunc logs the SQL statements being executed. Defaults to nil, meaning no logging.
	LogFunc LogFunc

	sqlDB      *sql.DB
	driverName string
}

// BuilderFunc creates a Builder instance using the given DB instance and Executor.
type BuilderFunc func(*DB, Executor) Builder

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

// Open opens a database specified by a driver name and data source name (DSN).
// Note that Open does not check if DSN is specified correctly. It doesn't try to establish a DB connection either.
// Please refer to sql.Open() for more information.
func Open(driverName, dsn string) (*DB, error) {
	sqlDB, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}

	db := &DB{
		driverName:  driverName,
		sqlDB:       sqlDB,
		FieldMapper: DefaultFieldMapFunc,
	}
	db.Builder = db.newBuilder(db.sqlDB)

	return db, nil
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
	//	placeholders := make([]string, 0)
	var placeholders []string
	count := 0
	s = plRegex.ReplaceAllStringFunc(s, func(m string) string {
		count++
		placeholders = append(placeholders, m[2:len(m)-1])
		return db.GeneratePlaceholder(count)
	})
	s = quoteRegex.ReplaceAllStringFunc(s, func(m string) string {
		if m[0] == '{' {
			return db.QuoteTableName(m[2 : len(m)-2])
		}
		return db.QuoteColumnName(m[2 : len(m)-2])
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
