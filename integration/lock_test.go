package dbx_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mysqlLockWaitTimeout = 1205

// queueItem reads the item table as a queue of pending rows.
type queueItem struct {
	ID   int
	Name string
}

func (queueItem) TableName() string { return "item" }

// getLockTestDB returns a freshly prepared DB whose connections are never reused,
// so a session variable set in one test never leaks into another.
func getLockTestDB(t *testing.T) *dbx.DB {
	db := getPreparedDB()
	db.DB().SetMaxIdleConns(0)
	db.DB().SetMaxOpenConns(4)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// beginWithLockTimeout starts a transaction whose lock waits time out after one second.
func beginWithLockTimeout(t *testing.T, db *dbx.DB) *dbx.Tx {
	tx, err := db.Begin()
	require.NoError(t, err)
	t.Cleanup(func() { _ = tx.Rollback() })
	_, err = tx.NewQuery("SET SESSION innodb_lock_wait_timeout=1").Execute()
	require.NoError(t, err)
	return tx
}

func assertLockWaitTimeout(t *testing.T, err error) {
	t.Helper()
	var me *mysql.MySQLError
	if assert.True(t, errors.As(err, &me), "want a MySQL lock wait timeout, got %v", err) {
		assert.Equal(t, uint16(mysqlLockWaitTimeout), me.Number)
	}
}

func TestSelectQuery_ForUpdate(t *testing.T) {
	db := getLockTestDB(t)
	tx1 := beginWithLockTimeout(t, db)
	tx2 := beginWithLockTimeout(t, db)

	var c Customer
	require.NoError(t, tx1.Select().Where(dbx.HashExp{"id": 1}).ForUpdate().One(&c))
	assert.Equal(t, "user1", c.Name)

	// the row is locked by tx1: tx2 waits and times out
	var c2 Customer
	assertLockWaitTimeout(t, tx2.Select().Where(dbx.HashExp{"id": 1}).ForUpdate().One(&c2))
}

func TestSelectQuery_ForUpdateSkipLocked(t *testing.T) {
	db := getLockTestDB(t)
	tx1 := beginWithLockTimeout(t, db)
	tx2 := beginWithLockTimeout(t, db)

	var c Customer
	require.NoError(t, tx1.Select().Where(dbx.HashExp{"id": 1}).ForUpdate().One(&c))

	// tx2 skips the row locked by tx1 instead of waiting
	var customers []Customer
	require.NoError(t, tx2.Select().Where(dbx.Between("id", 1, 3)).OrderBy("id").ForUpdateSkipLocked().All(&customers))
	if assert.Len(t, customers, 2) {
		assert.Equal(t, 2, customers[0].ID)
		assert.Equal(t, 3, customers[1].ID)
	}
}

func TestSelectQuery_ForShare(t *testing.T) {
	db := getLockTestDB(t)
	tx1 := beginWithLockTimeout(t, db)
	tx2 := beginWithLockTimeout(t, db)

	// shared locks do not conflict with each other
	var c1, c2 Customer
	require.NoError(t, tx1.Select().Where(dbx.HashExp{"id": 1}).ForShare().One(&c1))
	require.NoError(t, tx2.Select().Where(dbx.HashExp{"id": 1}).ForShare().One(&c2))

	// but a write waits for tx1's shared lock and times out
	_, err := tx2.Update("customer", dbx.Params{"name": "changed"}, dbx.HashExp{"id": 1}).Execute()
	assertLockWaitTimeout(t, err)
}

func TestSelectQuery_LockErrors(t *testing.T) {
	db := getLockTestDB(t)
	tx, err := db.Begin()
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()

	var customers []Customer
	other := tx.Select().From("customer").Build()
	assert.NotNil(t, tx.Select().From("customer").Union(other).ForUpdate().All(&customers))
	assert.NotNil(t, tx.Select("status").From("customer").GroupBy("status").ForUpdate().Column(&[]int{}))
	assert.NotNil(t, tx.Select("status").From("customer").Distinct(true).ForShare().Column(&[]int{}))

	// outside a transaction the lock would be released at once
	var c Customer
	assert.NotNil(t, db.Select().Where(dbx.HashExp{"id": 1}).ForUpdate().One(&c))
	assert.NotNil(t, db.Select().ForUpdateSkipLocked().All(&customers))
}

// TestSelectQuery_ForUpdateSkipLockedConcurrentClaimers is the job-queue pattern: two workers claim
// pending rows at the same time, each in its own transaction, and never get the same row.
func TestSelectQuery_ForUpdateSkipLockedConcurrentClaimers(t *testing.T) {
	db := getLockTestDB(t)
	// a safety net against a hang, not a synchronization mechanism
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	claim := func(tx *dbx.Tx) ([]queueItem, error) {
		var items []queueItem
		err := tx.Select().OrderBy("id").Limit(3).ForUpdateSkipLocked().WithContext(ctx).All(&items)
		return items, err
	}

	var (
		wg             sync.WaitGroup
		claimedA       []queueItem
		claimedB       []queueItem
		errA, errB     error
		aRead, bRead   = make(chan struct{}), make(chan struct{})
		closeA, closeB sync.Once
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		defer closeA.Do(func() { close(aRead) })
		errA = db.TransactionalContext(ctx, nil, func(tx *dbx.Tx) error {
			var err error
			if claimedA, err = claim(tx); err != nil {
				return err
			}
			closeA.Do(func() { close(aRead) })
			// keep the rows locked until B has claimed its rows
			select {
			case <-bRead:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}()
	go func() {
		defer wg.Done()
		defer closeB.Do(func() { close(bRead) })
		select {
		case <-aRead:
		case <-ctx.Done():
			errB = ctx.Err()
			return
		}
		errB = db.TransactionalContext(ctx, nil, func(tx *dbx.Tx) error {
			var err error
			claimedB, err = claim(tx)
			closeB.Do(func() { close(bRead) })
			return err
		})
	}()
	wg.Wait()

	require.NoError(t, errA)
	require.NoError(t, errB)
	idsA := map[int]bool{}
	for _, item := range claimedA {
		idsA[item.ID] = true
	}
	// item has 5 rows: A claims 3, B gets the 2 that A has not locked
	assert.Len(t, claimedA, 3)
	assert.Len(t, claimedB, 2)
	for _, item := range claimedB {
		assert.False(t, idsA[item.ID], "row %d claimed by both workers", item.ID)
	}
}
