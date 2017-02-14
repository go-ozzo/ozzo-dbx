// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbx

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Builder supports building SQL statements in a DB-agnostic way.
// Builder mainly provides two sets of query building methods: those building SELECT statements
// and those manipulating DB data or schema (e.g. INSERT statements, CREATE TABLE statements).
type Builder interface {
	// NewQuery creates a new Query object with the given SQL statement.
	// The SQL statement may contain parameter placeholders which can be bound with actual parameter
	// values before the statement is executed.
	NewQuery(string) *Query
	// Select returns a new SelectQuery object that can be used to build a SELECT statement.
	// The parameters to this method should be the list column names to be selected.
	// A column name may have an optional alias name. For example, Select("id", "my_name AS name").
	Select(...string) *SelectQuery
	// ModelQuery returns a new ModelQuery object that can be used to perform model insertion, update, and deletion.
	// The parameter to this method should be a pointer to the model struct that needs to be inserted, updated, or deleted.
	Model(interface{}) *ModelQuery

	// GeneratePlaceholder generates an anonymous parameter placeholder with the given parameter ID.
	GeneratePlaceholder(int) string

	// Quote quotes a string so that it can be embedded in a SQL statement as a string value.
	Quote(string) string
	// QuoteSimpleTableName quotes a simple table name.
	// A simple table name does not contain any schema prefix.
	QuoteSimpleTableName(string) string
	// QuoteSimpleColumnName quotes a simple column name.
	// A simple column name does not contain any table prefix.
	QuoteSimpleColumnName(string) string

	// QueryBuilder returns the query builder supporting the current DB.
	QueryBuilder() QueryBuilder

	// Insert creates a Query that represents an INSERT SQL statement.
	// The keys of cols are the column names, while the values of cols are the corresponding column
	// values to be inserted.
	Insert(table string, cols Params) *Query
	// Upsert creates a Query that represents an UPSERT SQL statement.
	// Upsert inserts a row into the table if the primary key or unique index is not found.
	// Otherwise it will update the row with the new values.
	// The keys of cols are the column names, while the values of cols are the corresponding column
	// values to be inserted.
	Upsert(table string, cols Params, constraints ...string) *Query
	// Update creates a Query that represents an UPDATE SQL statement.
	// The keys of cols are the column names, while the values of cols are the corresponding new column
	// values. If the "where" expression is nil, the UPDATE SQL statement will have no WHERE clause
	// (be careful in this case as the SQL statement will update ALL rows in the table).
	Update(table string, cols Params, where Expression) *Query
	// Delete creates a Query that represents a DELETE SQL statement.
	// If the "where" expression is nil, the DELETE SQL statement will have no WHERE clause
	// (be careful in this case as the SQL statement will delete ALL rows in the table).
	Delete(table string, where Expression) *Query

	// CreateTable creates a Query that represents a CREATE TABLE SQL statement.
	// The keys of cols are the column names, while the values of cols are the corresponding column types.
	// The optional "options" parameters will be appended to the generated SQL statement.
	CreateTable(table string, cols map[string]string, options ...string) *Query
	// RenameTable creates a Query that can be used to rename a table.
	RenameTable(oldName, newName string) *Query
	// DropTable creates a Query that can be used to drop a table.
	DropTable(table string) *Query
	// TruncateTable creates a Query that can be used to truncate a table.
	TruncateTable(table string) *Query

	// AddColumn creates a Query that can be used to add a column to a table.
	AddColumn(table, col, typ string) *Query
	// DropColumn creates a Query that can be used to drop a column from a table.
	DropColumn(table, col string) *Query
	// RenameColumn creates a Query that can be used to rename a column in a table.
	RenameColumn(table, oldName, newName string) *Query
	// AlterColumn creates a Query that can be used to change the definition of a table column.
	AlterColumn(table, col, typ string) *Query

	// AddPrimaryKey creates a Query that can be used to specify primary key(s) for a table.
	// The "name" parameter specifies the name of the primary key constraint.
	AddPrimaryKey(table, name string, cols ...string) *Query
	// DropPrimaryKey creates a Query that can be used to remove the named primary key constraint from a table.
	DropPrimaryKey(table, name string) *Query

	// AddForeignKey creates a Query that can be used to add a foreign key constraint to a table.
	// The length of cols and refCols must be the same as they refer to the primary and referential columns.
	// The optional "options" parameters will be appended to the SQL statement. They can be used to
	// specify options such as "ON DELETE CASCADE".
	AddForeignKey(table, name string, cols, refCols []string, refTable string, options ...string) *Query
	// DropForeignKey creates a Query that can be used to remove the named foreign key constraint from a table.
	DropForeignKey(table, name string) *Query

	// CreateIndex creates a Query that can be used to create an index for a table.
	CreateIndex(table, name string, cols ...string) *Query
	// CreateUniqueIndex creates a Query that can be used to create a unique index for a table.
	CreateUniqueIndex(table, name string, cols ...string) *Query
	// DropIndex creates a Query that can be used to remove the named index from a table.
	DropIndex(table, name string) *Query
}

// BaseBuilder provides a basic implementation of the Builder interface.
type BaseBuilder struct {
	db       *DB
	executor Executor
}

// NewBaseBuilder creates a new BaseBuilder instance.
func NewBaseBuilder(db *DB, executor Executor) *BaseBuilder {
	return &BaseBuilder{db, executor}
}

// DB returns the DB instance that this builder is associated with.
func (b *BaseBuilder) DB() *DB {
	return b.db
}

// Executor returns the executor object (a DB instance or a transaction) for executing SQL statements.
func (b *BaseBuilder) Executor() Executor {
	return b.executor
}

// NewQuery creates a new Query object with the given SQL statement.
// The SQL statement may contain parameter placeholders which can be bound with actual parameter
// values before the statement is executed.
func (b *BaseBuilder) NewQuery(sql string) *Query {
	return NewQuery(b.db, b.executor, sql)
}

// GeneratePlaceholder generates an anonymous parameter placeholder with the given parameter ID.
func (b *BaseBuilder) GeneratePlaceholder(int) string {
	return "?"
}

// Quote quotes a string so that it can be embedded in a SQL statement as a string value.
func (b *BaseBuilder) Quote(s string) string {
	return "'" + strings.Replace(s, "'", "''", -1) + "'"
}

// QuoteSimpleTableName quotes a simple table name.
// A simple table name does not contain any schema prefix.
func (b *BaseBuilder) QuoteSimpleTableName(s string) string {
	if strings.Contains(s, `"`) {
		return s
	}
	return `"` + s + `"`
}

// QuoteSimpleColumnName quotes a simple column name.
// A simple column name does not contain any table prefix.
func (b *BaseBuilder) QuoteSimpleColumnName(s string) string {
	if strings.Contains(s, `"`) || s == "*" {
		return s
	}
	return `"` + s + `"`
}

// Insert creates a Query that represents an INSERT SQL statement.
// The keys of cols are the column names, while the values of cols are the corresponding column
// values to be inserted.
func (b *BaseBuilder) Insert(table string, cols Params) *Query {
	names := make([]string, 0, len(cols))
	for name := range cols {
		names = append(names, name)
	}
	sort.Strings(names)

	params := Params{}
	columns := make([]string, 0, len(names))
	values := make([]string, 0, len(names))
	for _, name := range names {
		columns = append(columns, b.db.QuoteColumnName(name))
		value := cols[name]
		if e, ok := value.(Expression); ok {
			values = append(values, e.Build(b.db, params))
		} else {
			values = append(values, fmt.Sprintf("{:p%v}", len(params)))
			params[fmt.Sprintf("p%v", len(params))] = value
		}
	}

	var sql string
	if len(names) == 0 {
		sql = fmt.Sprintf("INSERT INTO %v DEFAULT VALUES", b.db.QuoteTableName(table))
	} else {
		sql = fmt.Sprintf("INSERT INTO %v (%v) VALUES (%v)",
			b.db.QuoteTableName(table),
			strings.Join(columns, ", "),
			strings.Join(values, ", "),
		)
	}

	return b.NewQuery(sql).Bind(params)
}

// Upsert creates a Query that represents an UPSERT SQL statement.
// Upsert inserts a row into the table if the primary key or unique index is not found.
// Otherwise it will update the row with the new values.
// The keys of cols are the column names, while the values of cols are the corresponding column
// values to be inserted.
func (b *BaseBuilder) Upsert(table string, cols Params, constraints ...string) *Query {
	q := b.NewQuery("")
	q.LastError = errors.New("Upsert is not supported")
	return q
}

// Update creates a Query that represents an UPDATE SQL statement.
// The keys of cols are the column names, while the values of cols are the corresponding new column
// values. If the "where" expression is nil, the UPDATE SQL statement will have no WHERE clause
// (be careful in this case as the SQL statement will update ALL rows in the table).
func (b *BaseBuilder) Update(table string, cols Params, where Expression) *Query {
	names := make([]string, 0, len(cols))
	for name := range cols {
		names = append(names, name)
	}
	sort.Strings(names)

	params := Params{}
	lines := make([]string, 0, len(names))
	for _, name := range names {
		value := cols[name]
		name = b.db.QuoteColumnName(name)
		if e, ok := value.(Expression); ok {
			lines = append(lines, name+"="+e.Build(b.db, params))
		} else {
			lines = append(lines, fmt.Sprintf("%v={:p%v}", name, len(params)))
			params[fmt.Sprintf("p%v", len(params))] = value
		}
	}

	sql := fmt.Sprintf("UPDATE %v SET %v", b.db.QuoteTableName(table), strings.Join(lines, ", "))
	if where != nil {
		w := where.Build(b.db, params)
		if w != "" {
			sql += " WHERE " + w
		}
	}

	return b.NewQuery(sql).Bind(params)
}

// Delete creates a Query that represents a DELETE SQL statement.
// If the "where" expression is nil, the DELETE SQL statement will have no WHERE clause
// (be careful in this case as the SQL statement will delete ALL rows in the table).
func (b *BaseBuilder) Delete(table string, where Expression) *Query {
	sql := "DELETE FROM " + b.db.QuoteTableName(table)
	params := Params{}
	if where != nil {
		w := where.Build(b.db, params)
		if w != "" {
			sql += " WHERE " + w
		}
	}
	return b.NewQuery(sql).Bind(params)
}

// CreateTable creates a Query that represents a CREATE TABLE SQL statement.
// The keys of cols are the column names, while the values of cols are the corresponding column types.
// The optional "options" parameters will be appended to the generated SQL statement.
func (b *BaseBuilder) CreateTable(table string, cols map[string]string, options ...string) *Query {
	names := []string{}
	for name := range cols {
		names = append(names, name)
	}
	sort.Strings(names)

	columns := []string{}
	for _, name := range names {
		columns = append(columns, b.db.QuoteColumnName(name)+" "+cols[name])
	}

	sql := fmt.Sprintf("CREATE TABLE %v (%v)", b.db.QuoteTableName(table), strings.Join(columns, ", "))
	for _, opt := range options {
		sql += " " + opt
	}

	return b.NewQuery(sql)
}

// RenameTable creates a Query that can be used to rename a table.
func (b *BaseBuilder) RenameTable(oldName, newName string) *Query {
	sql := fmt.Sprintf("RENAME TABLE %v TO %v", b.db.QuoteTableName(oldName), b.db.QuoteTableName(newName))
	return b.NewQuery(sql)
}

// DropTable creates a Query that can be used to drop a table.
func (b *BaseBuilder) DropTable(table string) *Query {
	sql := "DROP TABLE " + b.db.QuoteTableName(table)
	return b.NewQuery(sql)
}

// TruncateTable creates a Query that can be used to truncate a table.
func (b *BaseBuilder) TruncateTable(table string) *Query {
	sql := "TRUNCATE TABLE " + b.db.QuoteTableName(table)
	return b.NewQuery(sql)
}

// AddColumn creates a Query that can be used to add a column to a table.
func (b *BaseBuilder) AddColumn(table, col, typ string) *Query {
	sql := fmt.Sprintf("ALTER TABLE %v ADD %v %v", b.db.QuoteTableName(table), b.db.QuoteColumnName(col), typ)
	return b.NewQuery(sql)
}

// DropColumn creates a Query that can be used to drop a column from a table.
func (b *BaseBuilder) DropColumn(table, col string) *Query {
	sql := fmt.Sprintf("ALTER TABLE %v DROP COLUMN %v", b.db.QuoteTableName(table), b.db.QuoteColumnName(col))
	return b.NewQuery(sql)
}

// RenameColumn creates a Query that can be used to rename a column in a table.
func (b *BaseBuilder) RenameColumn(table, oldName, newName string) *Query {
	sql := fmt.Sprintf("ALTER TABLE %v RENAME COLUMN %v TO %v", b.db.QuoteTableName(table), b.db.QuoteColumnName(oldName), b.db.QuoteColumnName(newName))
	return b.NewQuery(sql)
}

// AlterColumn creates a Query that can be used to change the definition of a table column.
func (b *BaseBuilder) AlterColumn(table, col, typ string) *Query {
	col = b.db.QuoteColumnName(col)
	sql := fmt.Sprintf("ALTER TABLE %v CHANGE %v %v %v", b.db.QuoteTableName(table), col, col, typ)
	return b.NewQuery(sql)
}

// AddPrimaryKey creates a Query that can be used to specify primary key(s) for a table.
// The "name" parameter specifies the name of the primary key constraint.
func (b *BaseBuilder) AddPrimaryKey(table, name string, cols ...string) *Query {
	sql := fmt.Sprintf("ALTER TABLE %v ADD CONSTRAINT %v PRIMARY KEY (%v)",
		b.db.QuoteTableName(table),
		b.db.QuoteColumnName(name),
		b.quoteColumns(cols))
	return b.NewQuery(sql)
}

// DropPrimaryKey creates a Query that can be used to remove the named primary key constraint from a table.
func (b *BaseBuilder) DropPrimaryKey(table, name string) *Query {
	sql := fmt.Sprintf("ALTER TABLE %v DROP CONSTRAINT %v", b.db.QuoteTableName(table), b.db.QuoteColumnName(name))
	return b.NewQuery(sql)
}

// AddForeignKey creates a Query that can be used to add a foreign key constraint to a table.
// The length of cols and refCols must be the same as they refer to the primary and referential columns.
// The optional "options" parameters will be appended to the SQL statement. They can be used to
// specify options such as "ON DELETE CASCADE".
func (b *BaseBuilder) AddForeignKey(table, name string, cols, refCols []string, refTable string, options ...string) *Query {
	sql := fmt.Sprintf("ALTER TABLE %v ADD CONSTRAINT %v FOREIGN KEY (%v) REFERENCES %v (%v)",
		b.db.QuoteTableName(table),
		b.db.QuoteColumnName(name),
		b.quoteColumns(cols),
		b.db.QuoteTableName(refTable),
		b.quoteColumns(refCols))
	for _, opt := range options {
		sql += " " + opt
	}
	return b.NewQuery(sql)
}

// DropForeignKey creates a Query that can be used to remove the named foreign key constraint from a table.
func (b *BaseBuilder) DropForeignKey(table, name string) *Query {
	sql := fmt.Sprintf("ALTER TABLE %v DROP CONSTRAINT %v", b.db.QuoteTableName(table), b.db.QuoteColumnName(name))
	return b.NewQuery(sql)
}

// CreateIndex creates a Query that can be used to create an index for a table.
func (b *BaseBuilder) CreateIndex(table, name string, cols ...string) *Query {
	sql := fmt.Sprintf("CREATE INDEX %v ON %v (%v)",
		b.db.QuoteColumnName(name),
		b.db.QuoteTableName(table),
		b.quoteColumns(cols))
	return b.NewQuery(sql)
}

// CreateUniqueIndex creates a Query that can be used to create a unique index for a table.
func (b *BaseBuilder) CreateUniqueIndex(table, name string, cols ...string) *Query {
	sql := fmt.Sprintf("CREATE UNIQUE INDEX %v ON %v (%v)",
		b.db.QuoteColumnName(name),
		b.db.QuoteTableName(table),
		b.quoteColumns(cols))
	return b.NewQuery(sql)
}

// DropIndex creates a Query that can be used to remove the named index from a table.
func (b *BaseBuilder) DropIndex(table, name string) *Query {
	sql := fmt.Sprintf("DROP INDEX %v ON %v", b.db.QuoteColumnName(name), b.db.QuoteTableName(table))
	return b.NewQuery(sql)
}

// quoteColumns quotes a list of columns and concatenates them with commas.
func (b *BaseBuilder) quoteColumns(cols []string) string {
	s := ""
	for i, col := range cols {
		if i == 0 {
			s = b.db.QuoteColumnName(col)
		} else {
			s += ", " + b.db.QuoteColumnName(col)
		}
	}
	return s
}
