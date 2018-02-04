// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbx

import (
	ss "database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type City struct {
	ID   int
	Name string
}

func TestNewQuery(t *testing.T) {
	db := getDB()
	sql := "SELECT * FROM users WHERE id={:id}"
	q := NewQuery(db, db.sqlDB, sql)
	assert.Equal(t, q.SQL(), sql, "q.SQL()")
	assert.Equal(t, q.rawSQL, "SELECT * FROM users WHERE id=?", "q.RawSQL()")

	assert.Equal(t, len(q.Params()), 0, "len(q.Params())@1")
	q.Bind(Params{"id": 1})
	assert.Equal(t, len(q.Params()), 1, "len(q.Params())@2")
}

func TestQuery_Execute(t *testing.T) {
	db := getPreparedDB()
	defer db.Close()

	result, err := db.NewQuery("INSERT INTO item (name) VALUES ('test')").Execute()
	if assert.Nil(t, err) {
		rows, _ := result.RowsAffected()
		assert.Equal(t, rows, int64(1), "Result.RowsAffected()")
		lastID, _ := result.LastInsertId()
		assert.Equal(t, lastID, int64(6), "Result.LastInsertId()")
	}
}

type Customer struct {
	ID      int
	Email   string
	Status  int
	Name    string
	Address ss.NullString
}

func (m Customer) TableName() string {
	return "customer"
}

type CustomerPtr struct {
	ID      *int `db:"pk"`
	Email   *string
	Status  *int
	Name    string
	Address *string
}

func (m CustomerPtr) TableName() string {
	return "customer"
}

type CustomerNull struct {
	ID      ss.NullInt64 `db:"pk,id"`
	Email   ss.NullString
	Status  *ss.NullInt64
	Name    string
	Address ss.NullString
}

func (m CustomerNull) TableName() string {
	return "customer"
}

type CustomerEmbedded struct {
	Id    int
	Email *string
	InnerCustomer
}

func (m CustomerEmbedded) TableName() string {
	return "customer"
}

type CustomerEmbedded2 struct {
	ID    int
	Email *string
	Inner InnerCustomer
}

type InnerCustomer struct {
	Status  ss.NullInt64
	Name    *string
	Address ss.NullString
}

func TestQuery_Rows(t *testing.T) {
	db := getPreparedDB()
	defer db.Close()

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

	var customers2 []NullStringMap
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
	err = db.NewQuery(sql).Bind(Params{"id": 2}).One(&customer)
	if assert.Nil(t, err) {
		assert.Equal(t, customer.ID, 2, "customer.ID")
		assert.Equal(t, customer.Email, `user2@example.com`, "customer.Email")
		assert.Equal(t, customer.Status, 1, "customer.Status")
	}

	var customerPtr2 CustomerPtr
	sql = `SELECT id, email, address FROM customer WHERE id=2`
	rows2, err := db.sqlDB.Query(sql)
	defer rows2.Close()
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
	err = db.NewQuery(sql).Bind(Params{"id": 2}).One(&customerPtr)
	if assert.Nil(t, err) {
		assert.Equal(t, *customerPtr.ID, 2, "customer.ID")
		assert.Equal(t, *customerPtr.Email, `user2@example.com`, "customer.Email")
		assert.Equal(t, *customerPtr.Status, 1, "customer.Status")
	}

	// struct fields are null types
	var customerNull CustomerNull
	sql = `SELECT * FROM customer WHERE id={:id}`
	err = db.NewQuery(sql).Bind(Params{"id": 2}).One(&customerNull)
	if assert.Nil(t, err) {
		assert.Equal(t, customerNull.ID.Int64, int64(2), "customer.ID")
		assert.Equal(t, customerNull.Email.String, `user2@example.com`, "customer.Email")
		assert.Equal(t, customerNull.Status.Int64, int64(1), "customer.Status")
	}

	// embedded with anonymous struct
	var customerEmbedded CustomerEmbedded
	sql = `SELECT * FROM customer WHERE id={:id}`
	err = db.NewQuery(sql).Bind(Params{"id": 2}).One(&customerEmbedded)
	if assert.Nil(t, err) {
		assert.Equal(t, customerEmbedded.Id, 2, "customer.ID")
		assert.Equal(t, *customerEmbedded.Email, `user2@example.com`, "customer.Email")
		assert.Equal(t, customerEmbedded.Status.Int64, int64(1), "customer.Status")
	}

	// embedded with named struct
	var customerEmbedded2 CustomerEmbedded2
	sql = `SELECT id, email, status as "inner.status" FROM customer WHERE id={:id}`
	err = db.NewQuery(sql).Bind(Params{"id": 2}).One(&customerEmbedded2)
	if assert.Nil(t, err) {
		assert.Equal(t, customerEmbedded2.ID, 2, "customer.ID")
		assert.Equal(t, *customerEmbedded2.Email, `user2@example.com`, "customer.Email")
		assert.Equal(t, customerEmbedded2.Inner.Status.Int64, int64(1), "customer.Status")
	}

	customer2 := NullStringMap{}
	sql = `SELECT * FROM customer WHERE id={:id}`
	err = db.NewQuery(sql).Bind(Params{"id": 1}).One(customer2)
	if assert.Nil(t, err) {
		assert.Equal(t, customer2["id"].String, "1", "customer2[id]")
		assert.Equal(t, customer2["email"].String, `user1@example.com`, "customer2[email]")
		assert.Equal(t, customer2["status"].String, "1", "customer2[status]")
	}

	err = db.NewQuery(sql).Bind(Params{"id": 2}).One(customer)
	assert.NotNil(t, err)

	var customer3 NullStringMap
	err = db.NewQuery(sql).Bind(Params{"id": 2}).One(customer3)
	assert.NotNil(t, err)

	err = db.NewQuery(sql).Bind(Params{"id": 1}).One(&customer3)
	if assert.Nil(t, err) {
		assert.Equal(t, customer3["id"].String, "1", "customer3[id]")
	}

	// Rows
	sql = `SELECT * FROM customer ORDER BY id DESC`
	rows, err := db.NewQuery(sql).Rows()
	if assert.Nil(t, err) {
		s := ""
		for rows.Next() {
			rows.ScanStruct(&customer)
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
	q.Bind(Params{"id": 1}).One(&customer)
	assert.Equal(t, customer.ID, 1, "prepared@1")
	err = q.Bind(Params{"id": 20}).One(&customer)
	assert.Equal(t, err, ss.ErrNoRows, "prepared@2")
	q.Bind(Params{"id": 3}).One(&customer)
	assert.Equal(t, customer.ID, 3, "prepared@3")

	sql = `SELECT name FROM customer WHERE id={:id}`
	var name string
	q = db.NewQuery(sql).Prepare()
	q.Bind(Params{"id": 1}).Row(&name)
	assert.Equal(t, name, "user1", "prepared2@1")
	err = q.Bind(Params{"id": 20}).Row(&name)
	assert.Equal(t, err, ss.ErrNoRows, "prepared2@2")
	q.Bind(Params{"id": 3}).Row(&name)
	assert.Equal(t, name, "user3", "prepared2@3")

	// Query.LastError
	sql = `SELECT * FROM a`
	q = db.NewQuery(sql).Prepare()
	customer.ID = 100
	err = q.Bind(Params{"id": 1}).One(&customer)
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

func TestQuery_logSQL(t *testing.T) {
	db := getDB()
	q := db.NewQuery("SELECT * FROM users WHERE type={:type} AND id={:id}").Bind(Params{"type": "a", "id": 1})
	expected := "SELECT * FROM users WHERE type='a' AND id=1"
	assert.Equal(t, q.logSQL(), expected, "logSQL()")
}

func TestReplacePlaceholders(t *testing.T) {
	tests := []struct {
		ID             string
		Placeholders   []string
		Params         Params
		ExpectedParams string
		HasError       bool
	}{
		{"t1", nil, nil, "null", false},
		{"t2", []string{"id", "name"}, Params{"id": 1, "name": "xyz"}, `[1,"xyz"]`, false},
		{"t3", []string{"id", "name"}, Params{"id": 1}, `null`, true},
		{"t4", []string{"id", "name"}, Params{"id": 1, "name": "xyz", "age": 30}, `[1,"xyz"]`, false},
	}
	for _, test := range tests {
		params, err := replacePlaceholders(test.Placeholders, test.Params)
		result, _ := json.Marshal(params)
		assert.Equal(t, string(result), test.ExpectedParams, "params@"+test.ID)
		assert.Equal(t, err != nil, test.HasError, "error@"+test.ID)
	}
}

func TestIssue6(t *testing.T) {
	db := getPreparedDB()
	q := db.Select("*").From("customer").Where(HashExp{"id": 1})
	var customer Customer
	assert.Equal(t, q.One(&customer), nil)
	assert.Equal(t, 1, customer.ID)
}

type User struct {
	ID      int64
	Email   string
	Created time.Time
	Updated *time.Time
}

func TestIssue13(t *testing.T) {
	db := getPreparedDB()
	var user User
	err := db.Select().From("user").Where(HashExp{"id": 1}).One(&user)
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
