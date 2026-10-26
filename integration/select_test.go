package dbx_test

import (
	"database/sql"
	"testing"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/stretchr/testify/assert"
)

func TestSelectQuery_Data(t *testing.T) {
	db := getPreparedDB()
	defer func() { _ = db.Close() }()

	q := db.Select("id", "email").From("customer").OrderBy("id")

	var customer Customer
	assert.Nil(t, q.One(&customer))
	assert.Equal(t, customer.Email, "user1@example.com", "customer.Email")

	var customers []Customer
	assert.Nil(t, q.All(&customers))
	assert.Equal(t, len(customers), 3, "len(customers)")

	rows, _ := q.Rows()
	customer.Email = ""
	rows.Next()
	assert.Nil(t, rows.ScanStruct(&customer))
	assert.Equal(t, customer.Email, "user1@example.com", "customer.Email")

	var id, email string
	assert.Nil(t, q.Row(&id, &email))
	assert.Equal(t, id, "1", "id")
	assert.Equal(t, email, "user1@example.com", "email")

	var emails []string
	err := db.Select("email").From("customer").Column(&emails)
	if assert.Nil(t, err) {
		assert.Equal(t, 3, len(emails))
	}

	var e int
	err = db.Select().From("customer").One(&e)
	assert.NotNil(t, err)
	err = db.Select().From("customer").All(&e)
	assert.NotNil(t, err)
}

func TestSelectQuery_Model(t *testing.T) {
	db := getPreparedDB()
	defer func() { _ = db.Close() }()

	{
		// One without specifying FROM
		var customer CustomerPtr
		err := db.Select().OrderBy("id").One(&customer)
		if assert.Nil(t, err) {
			assert.Equal(t, "user1@example.com", *customer.Email)
		}
	}

	{
		// All without specifying FROM
		var customers []CustomerPtr
		err := db.Select().OrderBy("id").All(&customers)
		if assert.Nil(t, err) {
			assert.Equal(t, 3, len(customers))
		}
	}

	{
		// Model without specifying FROM
		var customer CustomerPtr
		err := db.Select().Model(2, &customer)
		if assert.Nil(t, err) {
			assert.Equal(t, "user2@example.com", *customer.Email)
		}
	}

	{
		// Model with WHERE
		var customer CustomerPtr
		err := db.Select().Where(dbx.HashExp{"id": 1}).Model(2, &customer)
		assert.Equal(t, sql.ErrNoRows, err)

		err = db.Select().Where(dbx.HashExp{"id": 2}).Model(2, &customer)
		assert.Nil(t, err)
	}

	{
		// errors
		var i int
		err := db.Select().Model(1, &i)
		assert.Equal(t, dbx.VarTypeError("must be a pointer to a struct"), err)

		var a struct {
			Name string
		}

		err = db.Select().Model(1, &a)
		assert.Equal(t, dbx.MissingPKError, err)
		var b struct {
			ID1 string `db:"pk"`
			ID2 string `db:"pk"`
		}
		err = db.Select().Model(1, &b)
		assert.Equal(t, dbx.CompositePKError, err)
	}
}

func TestSelectQuery_TableNamePointerReceiver(t *testing.T) {
	db := getPreparedDB()
	defer func() { _ = db.Close() }()

	// no From(): the table name comes from the pointer-receiver TableName() for values and slices
	var customers []CustomerPtrRecv
	if assert.Nil(t, db.Select().OrderBy("id").All(&customers)) {
		assert.Len(t, customers, 3)
	}
	var c CustomerPtrRecv
	if assert.Nil(t, db.Select().Where(dbx.HashExp{"id": 2}).One(&c)) {
		assert.Equal(t, "user2", c.Name)
	}
}

func TestSelectQuery_LikeEscapedWildcard(t *testing.T) {
	db := getPreparedDB()
	defer func() { _ = db.Close() }()

	for _, name := range []string{"50% off", "500 off"} {
		_, err := db.Insert("item", dbx.Params{"name": name}).Execute()
		assert.Nil(t, err)
	}

	// % in the value is escaped, so it matches only a literal %
	var names []string
	if assert.Nil(t, db.Select("name").From("item").Where(dbx.Like("name", "50%")).Column(&names)) {
		assert.Equal(t, []string{"50% off"}, names)
	}
}
