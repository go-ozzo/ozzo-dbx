package dbx

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// txDriverState records what a fake driver's transactions did.
type txDriverState struct {
	mu         sync.Mutex
	commits    int
	rollbacks  int
	rolledBack chan struct{} // closed on the first rollback
	once       sync.Once
}

type txConnector struct{ s *txDriverState }

func (c txConnector) Connect(context.Context) (driver.Conn, error) { return txConn(c), nil }
func (c txConnector) Driver() driver.Driver                        { return txDriver(c) }

type txDriver struct{ s *txDriverState }

func (d txDriver) Open(string) (driver.Conn, error) { return txConn(d), nil }

type txConn struct{ s *txDriverState }

func (c txConn) Prepare(string) (driver.Stmt, error) { return txStmt{}, nil }
func (c txConn) Close() error                        { return nil }
func (c txConn) Begin() (driver.Tx, error)           { return txTx(c), nil }
func (c txConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return txTx(c), nil
}

type txStmt struct{}

func (txStmt) Close() error                               { return nil }
func (txStmt) NumInput() int                              { return -1 }
func (txStmt) Exec([]driver.Value) (driver.Result, error) { return driver.RowsAffected(1), nil }
func (txStmt) Query([]driver.Value) (driver.Rows, error)  { return nil, errors.New("not supported") }

type txTx struct{ s *txDriverState }

func (t txTx) Commit() error {
	t.s.mu.Lock()
	defer t.s.mu.Unlock()
	t.s.commits++
	return nil
}

func (t txTx) Rollback() error {
	t.s.mu.Lock()
	t.s.rollbacks++
	t.s.mu.Unlock()
	t.s.once.Do(func() { close(t.s.rolledBack) })
	return nil
}

func newTxTestDB() (*DB, *txDriverState) {
	s := &txDriverState{rolledBack: make(chan struct{})}
	return NewFromDB(sql.OpenDB(txConnector{s}), "mysql"), s
}

func (s *txDriverState) counts() (int, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.commits, s.rollbacks
}

func TestTransactional_CanceledContext(t *testing.T) {
	// f succeeds, but its context is canceled before Transactional commits: database/sql rolls
	// the transaction back, so nothing is committed and success must not be reported
	cancelThenWait := func(cancel context.CancelFunc, s *txDriverState) func(*Tx) error {
		return func(tx *Tx) error {
			if _, err := tx.Insert("item", Params{"name": "x"}).Execute(); err != nil {
				return err
			}
			cancel()
			<-s.rolledBack
			return nil
		}
	}

	t.Run("TransactionalContext", func(t *testing.T) {
		db, s := newTxTestDB()
		ctx, cancel := context.WithCancel(context.Background())
		err := db.TransactionalContext(ctx, nil, cancelThenWait(cancel, s))
		assert.ErrorIs(t, err, context.Canceled)
		commits, _ := s.counts()
		assert.Equal(t, 0, commits)
	})

	t.Run("Transactional with DB context", func(t *testing.T) {
		db, s := newTxTestDB()
		ctx, cancel := context.WithCancel(context.Background())
		err := db.WithContext(ctx).Transactional(cancelThenWait(cancel, s))
		assert.ErrorIs(t, err, context.Canceled)
		commits, _ := s.counts()
		assert.Equal(t, 0, commits)
	})
}

func TestTransactional_EndedByCallback(t *testing.T) {
	for name, run := range map[string]func(db *DB, f func(*Tx) error) error{
		"Transactional": func(db *DB, f func(*Tx) error) error { return db.Transactional(f) },
		"TransactionalContext": func(db *DB, f func(*Tx) error) error {
			return db.TransactionalContext(context.Background(), nil, f)
		},
	} {
		t.Run(name, func(t *testing.T) {
			// committed normally
			db, s := newTxTestDB()
			assert.Nil(t, run(db, func(*Tx) error { return nil }))
			commits, rollbacks := s.counts()
			assert.Equal(t, 1, commits)
			assert.Equal(t, 0, rollbacks)

			// f commits itself and returns nil: no error
			db, s = newTxTestDB()
			assert.Nil(t, run(db, func(tx *Tx) error { return tx.Commit() }))
			commits, _ = s.counts()
			assert.Equal(t, 1, commits)

			// f rolls back itself and returns nil: no error (existing behavior)
			db, s = newTxTestDB()
			assert.Nil(t, run(db, func(tx *Tx) error { return tx.Rollback() }))
			commits, rollbacks = s.counts()
			assert.Equal(t, 0, commits)
			assert.Equal(t, 1, rollbacks)

			// f returns an error: rolled back, the error is returned
			db, s = newTxTestDB()
			errF := errors.New("f failed")
			assert.Equal(t, errF, run(db, func(*Tx) error { return errF }))
			commits, rollbacks = s.counts()
			assert.Equal(t, 0, commits)
			assert.Equal(t, 1, rollbacks)
		})
	}
}

func TestErrors_Unwrap(t *testing.T) {
	errF := errors.New("f failed")
	err := error(Errors{errF, sql.ErrTxDone})
	assert.ErrorIs(t, err, errF)
	assert.ErrorIs(t, err, sql.ErrTxDone)
	assert.Equal(t, "f failed\nsql: transaction has already been committed or rolled back", err.Error())
}
