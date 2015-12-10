// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dbx

import (
	"testing"
	"io/ioutil"
	"strings"
	"encoding/json"
	"bytes"
	_ "github.com/go-sql-driver/mysql"
)

const (
	TestDSN = "travis:@/ozzo_dbx_test"
	FixtureFile = "testdata/mysql.sql"
)

func TestDB_Open(t *testing.T) {
	if _, err := Open("mysql", TestDSN); err != nil {
		t.Errorf("Failed to open a mysql database: %v", err)
	}

	if _, err := Open("xyz", TestDSN); err == nil {
		t.Error("Using xyz driver should cause an error")
	}

	db, _ := Open("mysql", TestDSN)
	assertNotEqual(t, db.BaseDB, nil, "BaseDB")
	assertNotEqual(t, db.FieldMapper, nil, "MapField")
}

func TestDB_MustOpen(t *testing.T) {
	if _, err := MustOpen("mysql", TestDSN); err != nil {
		t.Errorf("Failed to open a mysql database: %v", err)
	}

	if _, err := MustOpen("mysql", "unknown:x@/test"); err == nil {
		t.Error("Using an invalid DSN should cause an error")
	}
}

func TestDB_Close(t *testing.T) {
	db := getDB()
	if err := db.Close(); err != nil {
		t.Errorf("Failed to close database: %v", err)
	}
}

func TestDB_DriverName(t *testing.T) {
	db := getDB()
	assertEqual(t, db.DriverName(), "mysql")
}

func TestDB_QuoteTableName(t *testing.T) {
	tests := []struct {
		input, output string
	}{
		{"users", "`users`"},
		{"`users`", "`users`"},
		{"(select)", "(select)"},
		{"{{users}}", "{{users}}"},
		{"public.db1.users", "`public`.`db1`.`users`"},
	}
	db := getDB()
	for _, test := range tests {
		result := db.QuoteTableName(test.input)
		if result != test.output {
			t.Errorf("QuoteTableName(%v) = %v, expected %v", test.input, result, test.output)
		}
	}
}

func TestDB_QuoteColumnName(t *testing.T) {
	tests := []struct {
		input, output string
	}{
		{"*", "*"},
		{"users.*", "`users`.*"},
		{"name", "`name`"},
		{"`name`", "`name`"},
		{"(select)", "(select)"},
		{"{{name}}", "{{name}}"},
		{"[[name]]", "[[name]]"},
		{"public.db1.users", "`public`.`db1`.`users`"},
	}
	db := getDB()
	for _, test := range tests {
		result := db.QuoteColumnName(test.input)
		if result != test.output {
			t.Errorf("QuoteColumnName(%v) = %v, expected %v", test.input, result, test.output)
		}
	}
}

func TestDB_ProcessSQL(t *testing.T) {
	tests := []struct {
		sql      string   // original SQL
		mysql    string   // expected MySQL version
		postgres string   // expected PostgreSQL version
		oci8     string   // expected OCI version
		params   []string // expected params
	}{
		{
			// normal case
			`INSERT INTO employee (id, name, age) VALUES ({:id}, {:name}, {:age})`,
			`INSERT INTO employee (id, name, age) VALUES (?, ?, ?)`,
			`INSERT INTO employee (id, name, age) VALUES ($1, $2, $3)`,
			`INSERT INTO employee (id, name, age) VALUES (:p1, :p2, :p3)`,
			[]string{"id", "name", "age"},
		},
		{
			// the same placeholder is used twice
			`SELECT * FROM employee WHERE first_name LIKE {:keyword} OR last_name LIKE {:keyword}`,
			`SELECT * FROM employee WHERE first_name LIKE ? OR last_name LIKE ?`,
			`SELECT * FROM employee WHERE first_name LIKE $1 OR last_name LIKE $2`,
			`SELECT * FROM employee WHERE first_name LIKE :p1 OR last_name LIKE :p2`,
			[]string{"keyword", "keyword"},
		},
		{
			// non-matching placeholder
			`SELECT * FROM employee WHERE first_name LIKE "{:key?word}" OR last_name LIKE {:keyword}`,
			`SELECT * FROM employee WHERE first_name LIKE "{:key?word}" OR last_name LIKE ?`,
			`SELECT * FROM employee WHERE first_name LIKE "{:key?word}" OR last_name LIKE $1`,
			`SELECT * FROM employee WHERE first_name LIKE "{:key?word}" OR last_name LIKE :p1`,
			[]string{"keyword"},
		},
		{
			// quote table/column names
			`SELECT * FROM {{public.user}} WHERE [[user.id]]=1`,
			"SELECT * FROM `public`.`user` WHERE `user`.`id`=1",
			"SELECT * FROM \"public\".\"user\" WHERE \"user\".\"id\"=1",
			"SELECT * FROM \"public\".\"user\" WHERE \"user\".\"id\"=1",
			[]string{},
		},
	}

	mysqlDB := getDB()
	mysqlDB.Builder = NewMysqlBuilder(nil, nil)
	pgsqlDB := getDB()
	pgsqlDB.Builder = NewPgsqlBuilder(nil, nil)
	ociDB := getDB()
	ociDB.Builder = NewOciBuilder(nil, nil)

	for _, test := range tests {
		s1, names := mysqlDB.processSQL(test.sql)
		if s1 != test.mysql {
			t.Errorf("mysql: %v, expected %v", s1, test.mysql)
		}
		s2, _ := pgsqlDB.processSQL(test.sql)
		if s2 != test.postgres {
			t.Errorf("postgres: %v, expected %v", s2, test.postgres)
		}
		s3, _ := ociDB.processSQL(test.sql)
		if s3 != test.oci8 {
			t.Errorf("oci8: %v, expected %v", s3, test.oci8)
		}

		names1, _ := json.Marshal(names)
		names2, _ := json.Marshal(test.params)
		if bytes.Compare(names1, names2) != 0 {
			t.Errorf("SQL: %v, got %v, expected %v", test.sql, string(names1), string(names2))
		}
	}
}

func TestDB_Begin(t *testing.T) {
	db := getDB()

	var (
		lastID int
		name string
		tx *Tx
	)
	db.NewQuery("SELECT MAX(id) FROM item").Row(&lastID)

	tx, _ = db.Begin()
	_, err1 := tx.Insert("item", Params{
		"name": "name1",
	}).Execute()
	_, err2 := tx.Insert("item", Params{
		"name": "name2",
	}).Execute()
	if err1 == nil && err2 == nil {
		tx.Commit()
	} else {
		t.Errorf("Unexpected TX rollback: %v, %v", err1, err2)
		tx.Rollback()
	}

	q := db.NewQuery("SELECT name FROM item WHERE id={:id}")
	q.Bind(Params{"id": lastID + 1}).Row(&name)
	assertEqual(t, name, "name1", "name1")
	q.Bind(Params{"id": lastID + 2}).Row(&name)
	assertEqual(t, name, "name2", "name2")

	tx, _ = db.Begin()
	_, err3 := tx.NewQuery("DELETE FROM item WHERE id=7").Execute()
	_, err4 := tx.NewQuery("DELETE FROM items WHERE id=7").Execute()
	if err3 == nil && err4 == nil {
		t.Error("Unexpected TX commit")
		tx.Commit()
	} else {
		tx.Rollback()
	}
}

func assertEqual(t *testing.T, actual, expected interface{}, hint ...string) {
	h := "got"
	if len(hint) > 0 {
		h = hint[0] + " ="
	}
	if actual != expected {
		t.Errorf("%v %v, expected %v", h, actual, expected)
	}
}

func assertNotEqual(t *testing.T, actual, expected interface{}, hint ...string) {
	h := "result"
	if len(hint) > 0 {
		h = hint[0]
	}
	if actual == expected {
		t.Errorf("%v should not be %v", h, expected)
	}
}

func getDB() *DB {
	db, err := Open("mysql", TestDSN)
	if err != nil {
		panic(err)
	}
	return db
}

func getPreparedDB() *DB {
	db := getDB()
	s, err := ioutil.ReadFile(FixtureFile)
	if err != nil {
		panic(err)
	}
	lines := strings.Split(string(s), ";")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if _, err := db.NewQuery(line).Execute(); err != nil {
			panic(err)
		}
	}
	return db
}
