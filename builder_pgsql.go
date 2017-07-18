// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbx

import (
	"fmt"
	"sort"
	"strings"
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
func (b *PgsqlBuilder) Upsert(table string, cols Params, constraints ...string) *Query {
	q := b.Insert(table, cols)

	names := []string{}
	for name := range cols {
		names = append(names, name)
	}
	sort.Strings(names)

	lines := []string{}
	for _, name := range names {
		value := cols[name]
		name = b.db.QuoteColumnName(name)
		if e, ok := value.(Expression); ok {
			lines = append(lines, name+"="+e.Build(b.db, q.params))
		} else {
			lines = append(lines, fmt.Sprintf("%v={:p%v}", name, len(q.params)))
			q.params[fmt.Sprintf("p%v", len(q.params))] = value
		}
	}

	if len(constraints) > 0 {
		c := b.quoteColumns(constraints)
		q.sql += " ON CONFLICT (" + c + ") DO UPDATE SET " + strings.Join(lines, ", ")
	} else {
		q.sql += " ON CONFLICT DO UPDATE SET " + strings.Join(lines, ", ")
	}

	return q
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
