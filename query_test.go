// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dbx

import (
	"encoding/json"
	"testing"
)

type City struct {
	ID   int
	Name string
}

func TestNewQuery(t *testing.T) {
	db := getDB()
	sql := "SELECT * FROM users WHERE id={:id}"
	q := NewQuery(db, db.sqlDB, sql)
	assertEqual(t, q.SQL(), sql, "q.SQL()")
	assertEqual(t, q.rawSQL, "SELECT * FROM users WHERE id=?", "q.RawSQL()")

	assertEqual(t, len(q.Params()), 0, "len(q.Params())@1")
	q.Bind(Params{"id": 1})
	assertEqual(t, len(q.Params()), 1, "len(q.Params())@2")
}

func TestQuery_Execute(t *testing.T) {
	db := getPreparedDB()
	defer db.Close()

	result, err := db.NewQuery("INSERT INTO item (name) VALUES ('test')").Execute()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	rows, _ := result.RowsAffected()
	assertEqual(t, rows, int64(1), "Result.RowsAffected()")
	lastID, _ := result.LastInsertId()
	assertEqual(t, lastID, int64(6), "Result.LastInsertId()")
}

type Customer struct {
	ID     int
	Email  string
	Status int
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
	if err := db.NewQuery(sql).All(&customers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else {
		assertEqual(t, len(customers), 3, "len(customers)")
		assertEqual(t, customers[2].ID, 3, "customers[2].ID")
		assertEqual(t, customers[2].Email, `user3@example.com`, "customers[2].Email")
		assertEqual(t, customers[2].Status, 2, "customers[2].Status")
	}
	var customers2 []NullStringMap
	if err := db.NewQuery(sql).All(&customers2); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else {
		assertEqual(t, len(customers2), 3, "len(customers2)")
		assertEqual(t, customers2[1]["id"].String, "2", "customers2[1][id]")
		assertEqual(t, customers2[1]["email"].String, `user2@example.com`, "customers2[1][email]")
		assertEqual(t, customers2[1]["status"].String, "1", "customers2[1][status]")
	}
	if err := db.NewQuery(sql).All(customers); err == nil {
		t.Error("Error expected when a non-pointer is used in All()")
	}
	var customers3 []string
	if err := db.NewQuery(sql).All(&customers3); err == nil {
		t.Error("Error expected when a slice of unsupported type is used in All()")
	}
	var customers4 string
	if err := db.NewQuery(sql).All(&customers4); err == nil {
		t.Error("Error expected when a non-slice is used in All()")
	}

	// One
	var customer Customer
	sql = `SELECT * FROM customer WHERE id={:id}`
	if err := db.NewQuery(sql).Bind(Params{"id": 2}).One(&customer); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else {
		assertEqual(t, customer.ID, 2, "customer.ID")
		assertEqual(t, customer.Email, `user2@example.com`, "customer.Email")
		assertEqual(t, customer.Status, 1, "customer.Status")
	}
	customer2 := NullStringMap{}
	if err := db.NewQuery(sql).Bind(Params{"id": 1}).One(customer2); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else {
		assertEqual(t, customer2["id"].String, "1", "customer2[id]")
		assertEqual(t, customer2["email"].String, `user1@example.com`, "customer2[email]")
		assertEqual(t, customer2["status"].String, "1", "customer2[status]")
	}
	if err := db.NewQuery(sql).Bind(Params{"id": 2}).One(customer); err == nil {
		t.Error("Error expected when a non-pointer is used in One()")
	}
	var customer3 NullStringMap
	if err := db.NewQuery(sql).Bind(Params{"id": 2}).One(customer3); err == nil {
		t.Error("Error expected when a nil NullStringMap is used One()")
	}
	if err := db.NewQuery(sql).Bind(Params{"id": 1}).One(&customer3); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else {
		assertEqual(t, customer3["id"].String, "1", "customer3[id]")
	}

	// Rows
	sql = `SELECT * FROM customer ORDER BY id DESC`
	rows, err := db.NewQuery(sql).Rows()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else {
		s := ""
		for rows.Next() {
			rows.ScanStruct(&customer)
			s += customer.Email + ","
		}
		assertEqual(t, s, "user3@example.com,user2@example.com,user1@example.com,", "Rows().Next()")
	}

	// FieldMapper
	var a struct {
		MyID string `db:"id"`
		name string
	}
	sql = `SELECT * FROM customer WHERE id=2`
	if err := db.NewQuery(sql).One(&a); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else {
		assertEqual(t, a.MyID, "2", "a.MyID")
		// unexported field is not populated
		assertEqual(t, a.name, "", "a.name")
	}

	// prepared statement
	sql = `SELECT * FROM customer WHERE id={:id}`
	q := db.NewQuery(sql).Prepare()
	q.Bind(Params{"id": 1}).One(&customer)
	assertEqual(t, customer.ID, 1, "prepared@1")
	q.Bind(Params{"id": 2}).One(&customer)
	assertEqual(t, customer.ID, 2, "prepared@2")

	// Query.LastError
	sql = `SELECT * FROM a`
	q = db.NewQuery(sql).Prepare()
	customer.ID = 100
	q.Bind(Params{"id": 1}).One(&customer)
	assertEqual(t, customer.ID, 100, "LastError@1")
	assertNotEqual(t, q.LastError, nil, "LastError@2")
}

func TestQuery_logSQL(t *testing.T) {
	db := getDB()
	q := db.NewQuery("SELECT * FROM users WHERE type={:type} AND id={:id}").Bind(Params{"type": "a", "id": 1})
	expected := "SELECT * FROM users WHERE type='a' AND id=1"
	assertEqual(t, q.logSQL(), expected, "logSQL()")
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
		assertEqual(t, string(result), test.ExpectedParams, "params@"+test.ID)
		assertEqual(t, err != nil, test.HasError, "error@"+test.ID)
	}
}

func TestIssue6(t *testing.T) {
	db := getDB()
	q := db.Select("*").From("customer").Where(HashExp{"id": 1})
	var customer Customer
	assertEqual(t, q.One(&customer), nil)
	assertEqual(t, 1, customer.ID)
}
