// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dbx

import (
	"fmt"
)

// MysqlBuilder is the builder for PostgreSQL databases.
type PgsqlBuilder struct {
	*BaseBuilder
	qb *BaseQueryBuilder
}

// NewPgsqlBuilder creates a new PgsqlBuilder instance.
func NewPgsqlBuilder(db *DB, executor Executor) Builder {
	return &PgsqlBuilder{
		NewBaseBuilder(db, executor),
		NewBaseQueryBuilder(db),
	}
}

func (b *PgsqlBuilder) GeneratePlaceholder(i int) string {
	return fmt.Sprintf("$%v", i)
}

func (b *PgsqlBuilder) QueryBuilder() QueryBuilder {
	return b.qb
}

func (b *PgsqlBuilder) DropIndex(table, name string) *Query {
	sql := fmt.Sprintf("DROP INDEX %v", b.db.QuoteColumnName(name))
	return b.NewQuery(sql)
}

func (b *PgsqlBuilder) RenameTable(oldName, newName string) *Query {
	sql := fmt.Sprintf("ALTER TABLE %v RENAME TO %v", b.db.QuoteTableName(oldName), b.db.QuoteTableName(newName))
	return b.NewQuery(sql)
}

func (b *PgsqlBuilder) AlterColumn(table, col, typ string) *Query {
	col = b.db.QuoteColumnName(col)
	sql := fmt.Sprintf("ALTER TABLE %v ALTER COLUMN %v %v", b.db.QuoteTableName(table), col, typ)
	return b.NewQuery(sql)
}
