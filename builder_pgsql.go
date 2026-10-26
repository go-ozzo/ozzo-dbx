// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbx

import (
	"errors"
	"fmt"
)

// PgsqlBuilder is the builder for PostgreSQL databases.
type PgsqlBuilder struct {
	*BaseBuilder
	qb *BaseQueryBuilder
}

var _ Builder = &PgsqlBuilder{}

// NewPgsqlBuilder creates a new PgsqlBuilder instance.
func NewPgsqlBuilder(db *DB, executor Executor) Builder {
	return &PgsqlBuilder{
		NewBaseBuilder(db, executor),
		NewBaseQueryBuilder(db),
	}
}

// Select returns a new SelectQuery object that can be used to build a SELECT statement.
// The parameters to this method should be the list column names to be selected.
// A column name may have an optional alias name. For example, Select("id", "my_name AS name").
func (b *PgsqlBuilder) Select(cols ...string) *SelectQuery {
	return NewSelectQuery(b, b.db).Select(cols...)
}

// Model returns a new ModelQuery object that can be used to perform model-based DB operations.
// The model passed to this method should be a pointer to a model struct.
func (b *PgsqlBuilder) Model(model interface{}) *ModelQuery {
	return NewModelQuery(model, b.db.FieldMapper, b.db, b)
}

// GeneratePlaceholder generates an anonymous parameter placeholder with the given parameter ID.
func (b *PgsqlBuilder) GeneratePlaceholder(i int) string {
	return fmt.Sprintf("$%v", i)
}

// QueryBuilder returns the query builder supporting the current DB.
func (b *PgsqlBuilder) QueryBuilder() QueryBuilder {
	return b.qb
}

// Upsert creates a Query that represents an UPSERT SQL statement.
// Upsert inserts a row into the table if the primary key or unique index is not found.
// Otherwise it will update the row with the new values.
// The keys of cols are the column names, while the values of cols are the corresponding column
// values to be inserted.
//
// The constraints are the conflict target columns (a primary key or unique index) and are required:
// PostgreSQL rejects ON CONFLICT DO UPDATE without a target, so Upsert sets LastError instead.
// The constraint columns are left out of the update list. Upsert also sets LastError if no column
// is left to update.
func (b *PgsqlBuilder) Upsert(table string, cols Params, constraints ...string) *Query {
	if len(constraints) == 0 {
		q := b.NewQuery("")
		q.LastError = errors.New("Upsert requires the conflict target columns (constraints) on PostgreSQL")
		return q
	}
	return b.upsert(table, cols, constraints, func(col string) string { return "EXCLUDED." + col }, func(assignments string) string {
		return " ON CONFLICT (" + b.quoteColumns(constraints) + ") DO UPDATE SET " + assignments
	})
}

// DropIndex creates a Query that can be used to remove the named index from a table.
func (b *PgsqlBuilder) DropIndex(table, name string) *Query {
	sql := fmt.Sprintf("DROP INDEX %v", b.db.QuoteColumnName(name))
	return b.NewQuery(sql)
}

// RenameTable creates a Query that can be used to rename a table.
func (b *PgsqlBuilder) RenameTable(oldName, newName string) *Query {
	sql := fmt.Sprintf("ALTER TABLE %v RENAME TO %v", b.db.QuoteTableName(oldName), b.db.QuoteTableName(newName))
	return b.NewQuery(sql)
}

// AlterColumn creates a Query that can be used to change the definition of a table column.
func (b *PgsqlBuilder) AlterColumn(table, col, typ string) *Query {
	col = b.db.QuoteColumnName(col)
	sql := fmt.Sprintf("ALTER TABLE %v ALTER COLUMN %v TYPE %v", b.db.QuoteTableName(table), col, typ)
	return b.NewQuery(sql)
}
