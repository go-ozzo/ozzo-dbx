// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbx

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	db := NewFromDB(nil, "mysql")
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
	db := NewFromDB(nil, "mysql")
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

	mysqlDB := NewFromDB(nil, "mysql")
	mysqlDB.Builder = NewMysqlBuilder(nil, nil)
	pgsqlDB := NewFromDB(nil, "postgres")
	pgsqlDB.Builder = NewPgsqlBuilder(nil, nil)
	ociDB := NewFromDB(nil, "oci8")
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

func TestErrors_Error(t *testing.T) {
	errs := Errors{}
	assert.Equal(t, "", errs.Error())
	errs = Errors{errors.New("a")}
	assert.Equal(t, "a", errs.Error())
	errs = Errors{errors.New("a"), errors.New("b")}
	assert.Equal(t, "a\nb", errs.Error())
}

// Naming according to issue 49 ( https://github.com/go-ozzo/ozzo-dbx/issues/49 )

type ArtistDAO struct{}

func (ArtistDAO) TableName() string {
	return "artists"
}

func Test_TableNameWithPrefix(t *testing.T) {
	db := NewFromDB(nil, "mysql")
	db.TableMapper = func(a interface{}) string {
		return "tbl_" + GetTableName(a)
	}
	assert.Equal(t, "tbl_artists", db.TableMapper(ArtistDAO{}))
}
