// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbx

// StandardBuilder is the builder that is used by DB for an unknown driver.
type StandardBuilder struct {
	*BaseBuilder
	qb *BaseQueryBuilder
}

var _ Builder = &StandardBuilder{}

// NewStandardBuilder creates a new StandardBuilder instance.
func NewStandardBuilder(db *DB, executor Executor) Builder {
	return &StandardBuilder{
		NewBaseBuilder(db, executor),
		NewBaseQueryBuilder(db),
	}
}

// QueryBuilder returns the query builder supporting the current DB.
func (b *StandardBuilder) QueryBuilder() QueryBuilder {
	return b.qb
}
