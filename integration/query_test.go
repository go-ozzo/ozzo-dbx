package dbx_test

import (
	ss "database/sql"
	"testing"
	"time"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/stretchr/testify/assert"
)

func TestQuery_Execute(t *testing.T) {
	db := getPreparedDB()
	defer func() { _ = db.Close() }()

	result, err := db.NewQuery("INSERT INTO item (name) VALUES ('test')").Execute()
	if assert.Nil(t, err) {
		rows, _ := result.RowsAffected()
		assert.Equal(t, rows, int64(1), "Result.RowsAffected()")
		lastID, _ := result.LastInsertId()
		assert.Equal(t, lastID, int64(6), "Result.LastInsertId()")
	}
}

func TestQuery_Rows(t *testing.T) {
	db := getPreparedDB()
	defer func() { _ = db.Close() }()

	var (
		sql string
		err error
	)

	// Query.All()
	var customers []Customer
	sql = `SELECT * FROM customer ORDER BY id`
	err = db.NewQuery(sql).All(&customers)
	if assert.Nil(t, err) {
		assert.Equal(t, len(customers), 3, "len(customers)")
		assert.Equal(t, customers[2].ID, 3, "customers[2].ID")
		assert.Equal(t, customers[2].Email, `user3@example.com`, "customers[2].Email")
		assert.Equal(t, customers[2].Status, 2, "customers[2].Status")
	}

	var customers2 []dbx.NullStringMap
	err = db.NewQuery(sql).All(&customers2)
	if assert.Nil(t, err) {
		assert.Equal(t, len(customers2), 3, "len(customers2)")
		assert.Equal(t, customers2[1]["id"].String, "2", "customers2[1][id]")
		assert.Equal(t, customers2[1]["email"].String, `user2@example.com`, "customers2[1][email]")
		assert.Equal(t, customers2[1]["status"].String, "1", "customers2[1][status]")
	}
	err = db.NewQuery(sql).All(customers)
	assert.NotNil(t, err)

	var customers3 []string
	err = db.NewQuery(sql).All(&customers3)
	assert.NotNil(t, err)

	var customers4 string
	err = db.NewQuery(sql).All(&customers4)
	assert.NotNil(t, err)

	var customers5 []Customer
	err = db.NewQuery(`SELECT * FROM customer WHERE id=999`).All(&customers5)
	if assert.Nil(t, err) {
		assert.NotNil(t, customers5)
		assert.Zero(t, len(customers5))
	}

	// One
	var customer Customer
	sql = `SELECT * FROM customer WHERE id={:id}`
	err = db.NewQuery(sql).Bind(dbx.Params{"id": 2}).One(&customer)
	if assert.Nil(t, err) {
		assert.Equal(t, customer.ID, 2, "customer.ID")
		assert.Equal(t, customer.Email, `user2@example.com`, "customer.Email")
		assert.Equal(t, customer.Status, 1, "customer.Status")
	}

	var customerPtr2 CustomerPtr
	sql = `SELECT id, email, address FROM customer WHERE id=2`
	rows2, err := db.DB().Query(sql)
	defer func() { _ = rows2.Close() }()
	assert.Nil(t, err)
	rows2.Next()
	err = rows2.Scan(&customerPtr2.ID, &customerPtr2.Email, &customerPtr2.Address)
	if assert.Nil(t, err) {
		assert.Equal(t, *customerPtr2.ID, 2, "customer.ID")
		assert.Equal(t, *customerPtr2.Email, `user2@example.com`)
		assert.Nil(t, customerPtr2.Address)
	}

	// struct fields are pointers
	var customerPtr CustomerPtr
	sql = `SELECT * FROM customer WHERE id={:id}`
	err = db.NewQuery(sql).Bind(dbx.Params{"id": 2}).One(&customerPtr)
	if assert.Nil(t, err) {
		assert.Equal(t, *customerPtr.ID, 2, "customer.ID")
		assert.Equal(t, *customerPtr.Email, `user2@example.com`, "customer.Email")
		assert.Equal(t, *customerPtr.Status, 1, "customer.Status")
	}

	// struct fields are null types
	var customerNull CustomerNull
	sql = `SELECT * FROM customer WHERE id={:id}`
	err = db.NewQuery(sql).Bind(dbx.Params{"id": 2}).One(&customerNull)
	if assert.Nil(t, err) {
		assert.Equal(t, customerNull.ID.Int64, int64(2), "customer.ID")
		assert.Equal(t, customerNull.Email.String, `user2@example.com`, "customer.Email")
		assert.Equal(t, customerNull.Status.Int64, int64(1), "customer.Status")
	}

	// embedded with anonymous struct
	var customerEmbedded CustomerEmbedded
	sql = `SELECT * FROM customer WHERE id={:id}`
	err = db.NewQuery(sql).Bind(dbx.Params{"id": 2}).One(&customerEmbedded)
	if assert.Nil(t, err) {
		assert.Equal(t, customerEmbedded.Id, 2, "customer.ID")
		assert.Equal(t, *customerEmbedded.Email, `user2@example.com`, "customer.Email")
		assert.Equal(t, customerEmbedded.Status.Int64, int64(1), "customer.Status")
	}

	// embedded with named struct
	var customerEmbedded2 CustomerEmbedded2
	sql = `SELECT id, email, status as "inner.status" FROM customer WHERE id={:id}`
	err = db.NewQuery(sql).Bind(dbx.Params{"id": 2}).One(&customerEmbedded2)
	if assert.Nil(t, err) {
		assert.Equal(t, customerEmbedded2.ID, 2, "customer.ID")
		assert.Equal(t, *customerEmbedded2.Email, `user2@example.com`, "customer.Email")
		assert.Equal(t, customerEmbedded2.Inner.Status.Int64, int64(1), "customer.Status")
	}

	customer2 := dbx.NullStringMap{}
	sql = `SELECT * FROM customer WHERE id={:id}`
	err = db.NewQuery(sql).Bind(dbx.Params{"id": 1}).One(customer2)
	if assert.Nil(t, err) {
		assert.Equal(t, customer2["id"].String, "1", "customer2[id]")
		assert.Equal(t, customer2["email"].String, `user1@example.com`, "customer2[email]")
		assert.Equal(t, customer2["status"].String, "1", "customer2[status]")
	}

	err = db.NewQuery(sql).Bind(dbx.Params{"id": 2}).One(customer)
	assert.NotNil(t, err)

	var customer3 dbx.NullStringMap
	err = db.NewQuery(sql).Bind(dbx.Params{"id": 2}).One(customer3)
	assert.NotNil(t, err)

	err = db.NewQuery(sql).Bind(dbx.Params{"id": 1}).One(&customer3)
	if assert.Nil(t, err) {
		assert.Equal(t, customer3["id"].String, "1", "customer3[id]")
	}

	// Rows
	sql = `SELECT * FROM customer ORDER BY id DESC`
	rows, err := db.NewQuery(sql).Rows()
	if assert.Nil(t, err) {
		s := ""
		for rows.Next() {
			assert.Nil(t, rows.ScanStruct(&customer))
			s += customer.Email + ","
		}
		assert.Equal(t, s, "user3@example.com,user2@example.com,user1@example.com,", "Rows().Next()")
	}

	// FieldMapper
	var a struct {
		MyID string `db:"id"`
		name string
	}
	sql = `SELECT * FROM customer WHERE id=2`
	err = db.NewQuery(sql).One(&a)
	if assert.Nil(t, err) {
		assert.Equal(t, a.MyID, "2", "a.MyID")
		// unexported field is not populated
		assert.Equal(t, a.name, "", "a.name")
	}

	// prepared statement
	sql = `SELECT * FROM customer WHERE id={:id}`
	q := db.NewQuery(sql).Prepare()
	assert.Nil(t, q.Bind(dbx.Params{"id": 1}).One(&customer))
	assert.Equal(t, customer.ID, 1, "prepared@1")
	err = q.Bind(dbx.Params{"id": 20}).One(&customer)
	assert.Equal(t, err, ss.ErrNoRows, "prepared@2")
	assert.Nil(t, q.Bind(dbx.Params{"id": 3}).One(&customer))
	assert.Equal(t, customer.ID, 3, "prepared@3")

	sql = `SELECT name FROM customer WHERE id={:id}`
	var name string
	q = db.NewQuery(sql).Prepare()
	assert.Nil(t, q.Bind(dbx.Params{"id": 1}).Row(&name))
	assert.Equal(t, name, "user1", "prepared2@1")
	err = q.Bind(dbx.Params{"id": 20}).Row(&name)
	assert.Equal(t, err, ss.ErrNoRows, "prepared2@2")
	assert.Nil(t, q.Bind(dbx.Params{"id": 3}).Row(&name))
	assert.Equal(t, name, "user3", "prepared2@3")

	// Query.LastError
	sql = `SELECT * FROM a`
	q = db.NewQuery(sql).Prepare()
	customer.ID = 100
	err = q.Bind(dbx.Params{"id": 1}).One(&customer)
	assert.NotEqual(t, err, nil, "LastError@0")
	assert.Equal(t, customer.ID, 100, "LastError@1")
	assert.Equal(t, q.LastError, nil, "LastError@2")

	// Query.Column
	sql = `SELECT name, id FROM customer ORDER BY id`
	var names []string
	err = db.NewQuery(sql).Column(&names)
	if assert.Nil(t, err) && assert.Equal(t, 3, len(names)) {
		assert.Equal(t, "user1", names[0])
		assert.Equal(t, "user2", names[1])
		assert.Equal(t, "user3", names[2])
	}
	err = db.NewQuery(sql).Column(names)
	assert.NotNil(t, err)
}

type User struct {
	ID      int64
	Email   string
	Created time.Time
	Updated *time.Time
}

func TestIssue6(t *testing.T) {
	db := getPreparedDB()
	q := db.Select("*").From("customer").Where(dbx.HashExp{"id": 1})
	var customer Customer
	assert.Equal(t, q.One(&customer), nil)
	assert.Equal(t, 1, customer.ID)
}

func TestIssue13(t *testing.T) {
	db := getPreparedDB()
	var user User
	err := db.Select().From("user").Where(dbx.HashExp{"id": 1}).One(&user)
	if assert.Nil(t, err) {
		assert.NotZero(t, user.Created)
		assert.Nil(t, user.Updated)
	}

	now := time.Now()

	user2 := User{
		Email:   "now@example.com",
		Created: now,
	}
	err = db.Model(&user2).Insert()
	if assert.Nil(t, err) {
		assert.NotZero(t, user2.ID)
	}

	user3 := User{
		Email:   "now@example.com",
		Created: now,
		Updated: &now,
	}
	err = db.Model(&user3).Insert()
	if assert.Nil(t, err) {
		assert.NotZero(t, user2.ID)
	}
}
