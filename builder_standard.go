// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dbx

// StandardBuilder is the builder that is used by DB for an unknown driver.
type StandardBuilder struct {
	*BaseBuilder
	qb *BaseQueryBuilder
}

// NewStandardBuilder creates a new StandardBuilder instance.
func NewStandardBuilder(db *DB, executor Executor) Builder {
	return &StandardBuilder{
		NewBaseBuilder(db, executor),
		NewBaseQueryBuilder(db),
	}
}

func (b *StandardBuilder) QueryBuilder() QueryBuilder {
	return b.qb
}
