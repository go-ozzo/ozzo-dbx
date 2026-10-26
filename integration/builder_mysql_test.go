package dbx_test

import (
	"testing"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestMysqlBuilder_UpsertNullConstraintColumn(t *testing.T) {
	db := getPreparedDB()
	defer func() { _ = db.Close() }()

	// a UNIQUE index allows several NULLs, so a NULL in the constraint column never conflicts:
	// both upserts insert a row
	for i := 0; i < 2; i++ {
		_, err := db.Upsert("device", dbx.Params{"serial": nil, "name": "unnamed"}, "serial").Execute()
		assert.Nil(t, err)
	}
	var n int
	assert.Nil(t, db.Select("COUNT(*)").From("device").Row(&n))
	assert.Equal(t, 2, n)
}

func TestMysqlBuilder_UpsertBeforeInsertTrigger(t *testing.T) {
	db := getPreparedDB()
	defer func() { _ = db.Close() }()

	// contact has a BEFORE INSERT trigger that lower-cases email. The update part copies the
	// proposed row, so the triggered value is stored on the update path too.
	_, err := db.Insert("contact", dbx.Params{"id": 1, "email": "First@Example.com"}).Execute()
	require.NoError(t, err)
	_, err = db.Upsert("contact", dbx.Params{"id": 1, "email": "MIXED@Example.com"}, "id").Execute()
	require.NoError(t, err)

	var email string
	require.NoError(t, db.Select("email").From("contact").Where(dbx.HashExp{"id": 1}).Row(&email))
	assert.Equal(t, "mixed@example.com", email)
}

func TestMysqlBuilder_UpsertSingleEvaluation(t *testing.T) {
	db := getPreparedDB()
	defer func() { _ = db.Close() }()

	// dbx_next_tag() counts its calls in a session variable; a transaction keeps one connection
	_, err := db.NewQuery("DROP FUNCTION IF EXISTS dbx_next_tag").Execute()
	require.NoError(t, err)
	_, err = db.NewQuery("CREATE FUNCTION dbx_next_tag() RETURNS INT NOT DETERMINISTIC NO SQL " +
		"BEGIN SET @dbx_calls = IFNULL(@dbx_calls, 0) + 1; RETURN @dbx_calls; END").Execute()
	require.NoError(t, err)
	defer func() { _, _ = db.NewQuery("DROP FUNCTION IF EXISTS dbx_next_tag").Execute() }()

	tx, err := db.Begin()
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()
	_, err = tx.NewQuery("SET @dbx_calls = 0").Execute()
	require.NoError(t, err)

	// customer 1 exists: the expression is evaluated once and its value is stored
	_, err = tx.Upsert("customer", dbx.Params{
		"id":     1,
		"email":  "user1@example.com",
		"status": dbx.NewExp("dbx_next_tag()"),
	}, "id").Execute()
	require.NoError(t, err)

	var calls, status int
	require.NoError(t, tx.NewQuery("SELECT @dbx_calls").Row(&calls))
	require.NoError(t, tx.Select("status").From("customer").Where(dbx.HashExp{"id": 1}).Row(&status))
	assert.Equal(t, 1, calls)
	assert.Equal(t, 1, status)
}
