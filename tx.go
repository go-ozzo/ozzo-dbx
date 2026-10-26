// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbx

import (
	"context"
	"database/sql"
	"errors"
	"sync/atomic"
)

// Tx enhances sql.Tx with additional querying methods.
type Tx struct {
	Builder
	tx *sql.Tx
	// ended is set once Commit or Rollback is called, so that Transactional can tell a transaction
	// ended by its callback from one rolled back by database/sql on a canceled context.
	ended atomic.Bool
}

func newTx(builder Builder, tx *sql.Tx) *Tx {
	return &Tx{Builder: builder, tx: tx}
}

// Commit commits the transaction.
func (t *Tx) Commit() error {
	t.ended.Store(true)
	return t.tx.Commit()
}

// Rollback aborts the transaction.
func (t *Tx) Rollback() error {
	t.ended.Store(true)
	return t.tx.Rollback()
}

// commitAfter commits the transaction after the callback of Transactional or TransactionalContext
// returned no error. Commit returns sql.ErrTxDone if the transaction has already ended. That is no
// error if the callback ended it itself. Otherwise database/sql rolled it back because ctx was
// canceled: nothing was committed, and reporting success would hide the lost writes.
func (t *Tx) commitAfter(ctx context.Context) error {
	err := t.tx.Commit()
	if !errors.Is(err, sql.ErrTxDone) {
		return err
	}
	if t.ended.Load() {
		return nil
	}
	if ctx != nil && ctx.Err() != nil {
		return ctx.Err()
	}
	return err
}
