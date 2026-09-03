package dbx_test

import (
	"context"
	"database/sql"
	"testing"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/stretchr/testify/assert"
)

func TestDB_NewFromDB(t *testing.T) {
	sqlDB, err := sql.Open("mysql", TestDSN)
	if assert.Nil(t, err) {
		db := dbx.NewFromDB(sqlDB, "mysql")
		assert.NotNil(t, db.DB())
		assert.NotNil(t, db.FieldMapper)
	}
}

func TestDB_Open(t *testing.T) {
	db, err := dbx.Open("mysql", TestDSN)
	assert.Nil(t, err)
	if assert.NotNil(t, db) {
		assert.NotNil(t, db.DB())
		assert.NotNil(t, db.FieldMapper)
		db2 := db.Clone()
		assert.NotEqual(t, db, db2)
		assert.Equal(t, db.DriverName(), db2.DriverName())
		ctx := context.Background()
		db3 := db.WithContext(ctx)
		assert.Equal(t, ctx, db3.Context())
		assert.NotEqual(t, db, db3)
	}

	_, err = dbx.Open("xyz", TestDSN)
	assert.NotNil(t, err)
}

func TestDB_MustOpen(t *testing.T) {
	_, err := dbx.MustOpen("mysql", TestDSN)
	assert.Nil(t, err)

	_, err = dbx.MustOpen("mysql", "unknown:x@/test")
	assert.NotNil(t, err)
}

func TestDB_Close(t *testing.T) {
	db := getDB()
	assert.Nil(t, db.Close())
}

func TestDB_DriverName(t *testing.T) {
	db := getDB()
	assert.Equal(t, "mysql", db.DriverName())
}

func TestDB_Begin(t *testing.T) {
	tests := []struct {
		makeTx func(db *dbx.DB) *dbx.Tx
		desc   string
	}{
		{
			makeTx: func(db *dbx.DB) *dbx.Tx {
				tx, _ := db.Begin()
				return tx
			},
			desc: "Begin",
		},
		{
			makeTx: func(db *dbx.DB) *dbx.Tx {
				sqlTx, _ := db.DB().Begin()
				return db.Wrap(sqlTx)
			},
			desc: "Wrap",
		},
		{
			makeTx: func(db *dbx.DB) *dbx.Tx {
				tx, _ := db.BeginTx(context.Background(), nil)
				return tx
			},
			desc: "BeginTx",
		},
	}

	db := getPreparedDB()

	var (
		lastID int
		name   string
		tx     *dbx.Tx
	)
	err := db.NewQuery("SELECT MAX(id) FROM item").Row(&lastID)
	assert.Nil(t, err)

	for _, test := range tests {
		t.Log(test.desc)

		tx = test.makeTx(db)
		_, err1 := tx.Insert("item", dbx.Params{
			"name": "name1",
		}).Execute()
		_, err2 := tx.Insert("item", dbx.Params{
			"name": "name2",
		}).Execute()
		if err1 == nil && err2 == nil {
			assert.Nil(t, tx.Commit())
		} else {
			t.Errorf("Unexpected TX rollback: %v, %v", err1, err2)
			_ = tx.Rollback()
		}

		q := db.NewQuery("SELECT name FROM item WHERE id={:id}")
		assert.Nil(t, q.Bind(dbx.Params{"id": lastID + 1}).Row(&name))
		assert.Equal(t, "name1", name)
		assert.Nil(t, q.Bind(dbx.Params{"id": lastID + 2}).Row(&name))
		assert.Equal(t, "name2", name)

		tx = test.makeTx(db)
		_, err3 := tx.NewQuery("DELETE FROM item WHERE id=7").Execute()
		_, err4 := tx.NewQuery("DELETE FROM items WHERE id=7").Execute()
		if err3 == nil && err4 == nil {
			t.Error("Unexpected TX commit")
			assert.Nil(t, tx.Commit())
		} else {
			_ = tx.Rollback()
		}
	}
}

func TestDB_Transactional(t *testing.T) {
	db := getPreparedDB()

	var (
		lastID int
		name   string
	)
	assert.Nil(t, db.NewQuery("SELECT MAX(id) FROM item").Row(&lastID))

	err := db.Transactional(func(tx *dbx.Tx) error {
		_, err := tx.Insert("item", dbx.Params{
			"name": "name1",
		}).Execute()
		if err != nil {
			return err
		}
		_, err = tx.Insert("item", dbx.Params{
			"name": "name2",
		}).Execute()
		if err != nil {
			return err
		}
		return nil
	})

	if assert.Nil(t, err) {
		q := db.NewQuery("SELECT name FROM item WHERE id={:id}")
		assert.Nil(t, q.Bind(dbx.Params{"id": lastID + 1}).Row(&name))
		assert.Equal(t, "name1", name)
		assert.Nil(t, q.Bind(dbx.Params{"id": lastID + 2}).Row(&name))
		assert.Equal(t, "name2", name)
	}

	err = db.Transactional(func(tx *dbx.Tx) error {
		_, err := tx.NewQuery("DELETE FROM item WHERE id=2").Execute()
		if err != nil {
			return err
		}
		_, err = tx.NewQuery("DELETE FROM items WHERE id=2").Execute()
		if err != nil {
			return err
		}
		return nil
	})
	if assert.NotNil(t, err) {
		assert.Nil(t, db.NewQuery("SELECT name FROM item WHERE id=2").Row(&name))
		assert.Equal(t, "Go in Action", name)
	}

	// Rollback called within Transactional and return error
	err = db.Transactional(func(tx *dbx.Tx) error {
		_, err := tx.NewQuery("DELETE FROM item WHERE id=2").Execute()
		if err != nil {
			return err
		}
		_, err = tx.NewQuery("DELETE FROM items WHERE id=2").Execute()
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		return nil
	})
	if assert.NotNil(t, err) {
		assert.Nil(t, db.NewQuery("SELECT name FROM item WHERE id=2").Row(&name))
		assert.Equal(t, "Go in Action", name)
	}

	// Rollback called within Transactional without returning error
	err = db.Transactional(func(tx *dbx.Tx) error {
		_, err := tx.NewQuery("DELETE FROM item WHERE id=2").Execute()
		if err != nil {
			return err
		}
		_, err = tx.NewQuery("DELETE FROM items WHERE id=2").Execute()
		if err != nil {
			_ = tx.Rollback()
			return nil
		}
		return nil
	})
	if assert.Nil(t, err) {
		assert.Nil(t, db.NewQuery("SELECT name FROM item WHERE id=2").Row(&name))
		assert.Equal(t, "Go in Action", name)
	}
}
