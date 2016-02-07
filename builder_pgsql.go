// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dbx

import (
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

// GeneratePlaceholder generates an anonymous parameter placeholder with the given parameter ID.
func (b *PgsqlBuilder) GeneratePlaceholder(i int) string {
	return fmt.Sprintf("$%v", i)
}

// QueryBuilder returns the query builder supporting the current DB.
func (b *PgsqlBuilder) QueryBuilder() QueryBuilder {
	return b.qb
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
	sql := fmt.Sprintf("ALTER TABLE %v ALTER COLUMN %v %v", b.db.QuoteTableName(table), col, typ)
	return b.NewQuery(sql)
}
