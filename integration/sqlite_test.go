package dbx_test

import (
	"database/sql/driver"
	"sync"
	"sync/atomic"
	"testing"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"modernc.org/sqlite"
)

// These tests run on an in-memory SQLite database (modernc.org/sqlite, no CGo), so they need
// no server and run wherever the integration suite runs.

var (
	sqliteCallsOnce sync.Once
	sqliteCalls     atomic.Int64
)

// getSQLiteDB opens a private in-memory database; one connection, so every query sees it.
func getSQLiteDB(t *testing.T) *dbx.DB {
	// dbx_next() returns the number of times it has been called: a non-deterministic function
	// that shows how often an expression is evaluated
	sqliteCallsOnce.Do(func() {
		err := sqlite.RegisterScalarFunction("dbx_next", 0, func(*sqlite.FunctionContext, []driver.Value) (driver.Value, error) {
			return sqliteCalls.Add(1), nil
		})
		if err != nil {
			panic(err)
		}
	})
	db, err := dbx.Open("sqlite", ":memory:")
	require.NoError(t, err)
	db.DB().SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestSqliteBuilder_UpsertSingleEvaluation(t *testing.T) {
	db := getSQLiteDB(t)
	_, err := db.NewQuery("CREATE TABLE tag (id INTEGER PRIMARY KEY, v INTEGER NOT NULL)").Execute()
	require.NoError(t, err)
	_, err = db.Insert("tag", dbx.Params{"id": 1, "v": 100}).Execute()
	require.NoError(t, err)

	// conflict: the expression is evaluated once, and the update stores the proposed value
	sqliteCalls.Store(0)
	_, err = db.Upsert("tag", dbx.Params{"id": 1, "v": dbx.NewExp("dbx_next()")}, "id").Execute()
	require.NoError(t, err)
	var v int64
	require.NoError(t, db.Select("v").From("tag").Where(dbx.HashExp{"id": 1}).Row(&v))
	assert.Equal(t, int64(1), sqliteCalls.Load())
	assert.Equal(t, int64(1), v)

	// no conflict: evaluated once, inserted
	_, err = db.Upsert("tag", dbx.Params{"id": 2, "v": dbx.NewExp("dbx_next()")}, "id").Execute()
	require.NoError(t, err)
	require.NoError(t, db.Select("v").From("tag").Where(dbx.HashExp{"id": 2}).Row(&v))
	assert.Equal(t, int64(2), sqliteCalls.Load())
	assert.Equal(t, int64(2), v)
}

func TestSqliteBuilder_UpsertKeepsOtherColumns(t *testing.T) {
	db := getSQLiteDB(t)
	_, err := db.NewQuery("CREATE TABLE note (id INTEGER PRIMARY KEY, title TEXT, body TEXT)").Execute()
	require.NoError(t, err)
	_, err = db.Insert("note", dbx.Params{"id": 1, "title": "t1", "body": "b1"}).Execute()
	require.NoError(t, err)

	// only the passed columns are updated from the proposed row; the constraint column is not
	_, err = db.Upsert("note", dbx.Params{"id": 1, "title": "t2"}, "id").Execute()
	require.NoError(t, err)
	var title, body string
	require.NoError(t, db.Select("title", "body").From("note").Where(dbx.HashExp{"id": 1}).Row(&title, &body))
	assert.Equal(t, "t2", title)
	assert.Equal(t, "b1", body)
}
