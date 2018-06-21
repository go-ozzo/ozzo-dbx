// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbx

import (
	"database/sql"
	"errors"
	"io/ioutil"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

const (
	TestDSN     = "travis:@/ozzo_dbx_test?parseTime=true"
	FixtureFile = "testdata/mysql.sql"
)

func TestDB_NewFromDB(t *testing.T) {
	sqlDB, err := sql.Open("mysql", TestDSN)
	if assert.Nil(t, err) {
		db := NewFromDB(sqlDB, "mysql")
		assert.NotNil(t, db.sqlDB)
		assert.NotNil(t, db.FieldMapper)
	}
}

func TestDB_Open(t *testing.T) {
	db, err := Open("mysql", TestDSN)
	assert.Nil(t, err)
	if assert.NotNil(t, db) {
		assert.NotNil(t, db.sqlDB)
		assert.NotNil(t, db.FieldMapper)
		db2 := db.Clone()
		assert.Equal(t, db.driverName, db2.driverName)
	}

	_, err = Open("xyz", TestDSN)
	assert.NotNil(t, err)
}

func TestDB_MustOpen(t *testing.T) {
	_, err := MustOpen("mysql", TestDSN)
	assert.Nil(t, err)

	_, err = MustOpen("mysql", "unknown:x@/test")
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
		assert.Equal(t, test.output, result, test.input)
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
		assert.Equal(t, test.output, result, test.input)
	}
}

func TestDB_ProcessSQL(t *testing.T) {
	tests := []struct {
		tag      string
		sql      string   // original SQL
		mysql    string   // expected MySQL version
		postgres string   // expected PostgreSQL version
		oci8     string   // expected OCI version
		params   []string // expected params
	}{
		{
			"normal case",
			`INSERT INTO employee (id, name, age) VALUES ({:id}, {:name}, {:age})`,
			`INSERT INTO employee (id, name, age) VALUES (?, ?, ?)`,
			`INSERT INTO employee (id, name, age) VALUES ($1, $2, $3)`,
			`INSERT INTO employee (id, name, age) VALUES (:p1, :p2, :p3)`,
			[]string{"id", "name", "age"},
		},
		{
			"the same placeholder is used twice",
			`SELECT * FROM employee WHERE first_name LIKE {:keyword} OR last_name LIKE {:keyword}`,
			`SELECT * FROM employee WHERE first_name LIKE ? OR last_name LIKE ?`,
			`SELECT * FROM employee WHERE first_name LIKE $1 OR last_name LIKE $2`,
			`SELECT * FROM employee WHERE first_name LIKE :p1 OR last_name LIKE :p2`,
			[]string{"keyword", "keyword"},
		},
		{
			"non-matching placeholder",
			`SELECT * FROM employee WHERE first_name LIKE "{:key?word}" OR last_name LIKE {:keyword}`,
			`SELECT * FROM employee WHERE first_name LIKE "{:key?word}" OR last_name LIKE ?`,
			`SELECT * FROM employee WHERE first_name LIKE "{:key?word}" OR last_name LIKE $1`,
			`SELECT * FROM employee WHERE first_name LIKE "{:key?word}" OR last_name LIKE :p1`,
			[]string{"keyword"},
		},
		{
			"quote table/column names",
			`SELECT * FROM {{public.user}} WHERE [[user.id]]=1`,
			"SELECT * FROM `public`.`user` WHERE `user`.`id`=1",
			"SELECT * FROM \"public\".\"user\" WHERE \"user\".\"id\"=1",
			"SELECT * FROM \"public\".\"user\" WHERE \"user\".\"id\"=1",
			nil,
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
		assert.Equal(t, test.mysql, s1, test.tag)
		s2, _ := pgsqlDB.processSQL(test.sql)
		assert.Equal(t, test.postgres, s2, test.tag)
		s3, _ := ociDB.processSQL(test.sql)
		assert.Equal(t, test.oci8, s3, test.tag)

		assert.Equal(t, test.params, names, test.tag)
	}
}

func TestDB_Begin(t *testing.T) {
	tests := []struct {
		makeTx func(db *DB) *Tx
		desc   string
	}{
		{
			makeTx: func(db *DB) *Tx {
				tx, _ := db.Begin()
				return tx
			},
			desc: "Begin",
		},
		{
			makeTx: func(db *DB) *Tx {
				sqlTx, _ := db.DB().Begin()
				return db.Wrap(sqlTx)
			},
			desc: "Wrap",
		},
	}

	db := getPreparedDB()

	var (
		lastID int
		name   string
		tx     *Tx
	)
	db.NewQuery("SELECT MAX(id) FROM item").Row(&lastID)

	for _, test := range tests {
		t.Log(test.desc)

		tx = test.makeTx(db)
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
		assert.Equal(t, "name1", name)
		q.Bind(Params{"id": lastID + 2}).Row(&name)
		assert.Equal(t, "name2", name)

		tx = test.makeTx(db)
		_, err3 := tx.NewQuery("DELETE FROM item WHERE id=7").Execute()
		_, err4 := tx.NewQuery("DELETE FROM items WHERE id=7").Execute()
		if err3 == nil && err4 == nil {
			t.Error("Unexpected TX commit")
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}
}

func TestDB_Transactional(t *testing.T) {
	db := getPreparedDB()

	var (
		lastID int
		name   string
	)
	db.NewQuery("SELECT MAX(id) FROM item").Row(&lastID)

	err := db.Transactional(func(tx *Tx) error {
		_, err := tx.Insert("item", Params{
			"name": "name1",
		}).Execute()
		if err != nil {
			return err
		}
		_, err = tx.Insert("item", Params{
			"name": "name2",
		}).Execute()
		if err != nil {
			return err
		}
		return nil
	})

	if assert.Nil(t, err) {
		q := db.NewQuery("SELECT name FROM item WHERE id={:id}")
		q.Bind(Params{"id": lastID + 1}).Row(&name)
		assert.Equal(t, "name1", name)
		q.Bind(Params{"id": lastID + 2}).Row(&name)
		assert.Equal(t, "name2", name)
	}

	err = db.Transactional(func(tx *Tx) error {
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
		db.NewQuery("SELECT name FROM item WHERE id=2").Row(&name)
		assert.Equal(t, "Go in Action", name)
	}

	// Rollback called within Transactional and return error
	err = db.Transactional(func(tx *Tx) error {
		_, err := tx.NewQuery("DELETE FROM item WHERE id=2").Execute()
		if err != nil {
			return err
		}
		_, err = tx.NewQuery("DELETE FROM items WHERE id=2").Execute()
		if err != nil {
			tx.Rollback()
			return err
		}
		return nil
	})
	if assert.NotNil(t, err) {
		db.NewQuery("SELECT name FROM item WHERE id=2").Row(&name)
		assert.Equal(t, "Go in Action", name)
	}

	// Rollback called within Transactional without returning error
	err = db.Transactional(func(tx *Tx) error {
		_, err := tx.NewQuery("DELETE FROM item WHERE id=2").Execute()
		if err != nil {
			return err
		}
		_, err = tx.NewQuery("DELETE FROM items WHERE id=2").Execute()
		if err != nil {
			tx.Rollback()
			return nil
		}
		return nil
	})
	if assert.Nil(t, err) {
		db.NewQuery("SELECT name FROM item WHERE id=2").Row(&name)
		assert.Equal(t, "Go in Action", name)
	}
}

func TestErrors_Error(t *testing.T) {
	errs := Errors{}
	assert.Equal(t, "", errs.Error())
	errs = Errors{errors.New("a")}
	assert.Equal(t, "a", errs.Error())
	errs = Errors{errors.New("a"), errors.New("b")}
	assert.Equal(t, "a\nb", errs.Error())
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
