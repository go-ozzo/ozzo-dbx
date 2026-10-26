package dbx_test

import (
	"testing"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/stretchr/testify/assert"
)

func TestMysqlBuilder_RenameColumn(t *testing.T) {
	db := getPreparedDB()
	defer func() { _ = db.Close() }()

	b := dbx.NewMysqlBuilder(db, db.DB())
	db.Builder = b

	// This requires SHOW CREATE TABLE against real MySQL
	q := b.RenameColumn("customer", "email", "e")
	assert.Equal(t, q.SQL(), "ALTER TABLE `customer` CHANGE `email` `e` varchar(128) NOT NULL")
}

func TestMysqlBuilder_Upsert(t *testing.T) {
	db := getPreparedDB()
	defer func() { _ = db.Close() }()

	// existing key: the row is updated instead of failing with a duplicate-key error
	_, err := db.Upsert("customer", dbx.Params{
		"id":     1,
		"email":  "new1@example.com",
		"name":   dbx.NewExp("UPPER({:name})", dbx.Params{"name": "renamed"}),
		"status": 5,
	}, "id").Execute()
	if assert.Nil(t, err) {
		var c Customer
		assert.Nil(t, db.Select().From("customer").Where(dbx.HashExp{"id": 1}).One(&c))
		assert.Equal(t, "new1@example.com", c.Email)
		assert.Equal(t, "RENAMED", c.Name)
		assert.Equal(t, 5, c.Status)
		assert.Equal(t, "address1", c.Address.String)
	}

	// new key: the row is inserted
	_, err = db.Upsert("customer", dbx.Params{"id": 10, "email": "u10@example.com", "name": "u10"}, "id").Execute()
	if assert.Nil(t, err) {
		var c Customer
		assert.Nil(t, db.Select().From("customer").Where(dbx.HashExp{"id": 10}).One(&c))
		assert.Equal(t, "u10", c.Name)
	}

	// any unique key fires ON DUPLICATE KEY UPDATE; the constraint column keeps the row's value
	_, err = db.Upsert("account", dbx.Params{"id": 99, "code": "A1", "name": "renamed"}, "id").Execute()
	if assert.Nil(t, err) {
		var a []Account
		assert.Nil(t, db.Select().From("account").All(&a))
		if assert.Len(t, a, 1) {
			assert.Equal(t, 1, a[0].ID)
			assert.Equal(t, "renamed", a[0].Name)
		}
	}

	// nothing to update besides the constraint columns
	_, err = db.Upsert("account", dbx.Params{"id": 1}, "id").Execute()
	assert.NotNil(t, err)
}
