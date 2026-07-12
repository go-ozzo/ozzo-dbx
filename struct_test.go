// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbx

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultFieldMapFunc(t *testing.T) {
	tests := []struct {
		input, output string
	}{
		{"Name", "name"},
		{"FirstName", "first_name"},
		{"Name0", "name0"},
		{"ID", "id"},
		{"UserID", "user_id"},
		{"User0ID", "user0_id"},
		{"MyURL", "my_url"},
		{"URLPath", "urlpath"},
		{"MyURLPath", "my_urlpath"},
		{"First_Name", "first_name"},
		{"first_name", "first_name"},
		{"_FirstName", "_first_name"},
		{"_First_Name", "_first_name"},
	}
	for _, test := range tests {
		r := DefaultFieldMapFunc(test.input)
		assert.Equal(t, test.output, r, test.input)
	}
}

func Test_concat(t *testing.T) {
	assert.Equal(t, "a.b", concat("a", "b"))
	assert.Equal(t, "a", concat("a", ""))
	assert.Equal(t, "b", concat("", "b"))
}

func Test_parseTag(t *testing.T) {
	name, pk := parseTag("abc")
	assert.Equal(t, "abc", name)
	assert.False(t, pk)

	name, pk = parseTag("pk,abc")
	assert.Equal(t, "abc", name)
	assert.True(t, pk)

	name, pk = parseTag("pk")
	assert.Equal(t, "", name)
	assert.True(t, pk)
}

func Test_indirect(t *testing.T) {
	var a int
	assert.Equal(t, reflect.ValueOf(a).Kind(), indirect(reflect.ValueOf(a)).Kind())
	var b *int
	bi := indirect(reflect.ValueOf(&b))
	assert.Equal(t, reflect.ValueOf(a).Kind(), bi.Kind())
	if assert.NotNil(t, b) {
		assert.Equal(t, 0, *b)
	}
}

func Test_structValue_columns(t *testing.T) {
	customer := Customer{
		ID:     1,
		Name:   "abc",
		Status: 2,
		Email:  "abc@example.com",
	}
	sv := newStructValue(&customer, DefaultFieldMapFunc, GetTableName)
	cols := sv.columns(nil, nil)
	assert.Equal(t, map[string]interface{}{"id": 1, "name": "abc", "status": 2, "email": "abc@example.com", "address": sql.NullString{}}, cols)

	cols = sv.columns([]string{"ID", "name"}, nil)
	assert.Equal(t, map[string]interface{}{"id": 1}, cols)

	cols = sv.columns([]string{"ID", "Name"}, []string{"ID"})
	assert.Equal(t, map[string]interface{}{"name": "abc"}, cols)

	cols = sv.columns(nil, []string{"ID", "Address"})
	assert.Equal(t, map[string]interface{}{"name": "abc", "status": 2, "email": "abc@example.com"}, cols)

	sv = newStructValue(&customer, nil, GetTableName)
	cols = sv.columns([]string{"ID", "Name"}, []string{"ID"})
	assert.Equal(t, map[string]interface{}{"Name": "abc"}, cols)
}

func TestIssue37(t *testing.T) {
	customer := Customer{
		ID:     1,
		Name:   "abc",
		Status: 2,
		Email:  "abc@example.com",
	}
	ev := struct {
		Customer
		Status string
	}{customer, "20"}
	sv := newStructValue(&ev, nil, GetTableName)
	cols := sv.columns([]string{"ID", "Status"}, nil)
	assert.Equal(t, map[string]interface{}{"ID": 1, "Status": "20"}, cols)

	ev2 := struct {
		Status string
		Customer
	}{"20", customer}
	sv = newStructValue(&ev2, nil, GetTableName)
	cols = sv.columns([]string{"ID", "Status"}, nil)
	assert.Equal(t, map[string]interface{}{"ID": 1, "Status": "20"}, cols)
}

type MyCustomer struct{}

func TestGetTableName(t *testing.T) {
	var c1 Customer
	assert.Equal(t, "customer", GetTableName(c1))

	var c2 *Customer
	assert.Equal(t, "customer", GetTableName(c2))

	var c3 MyCustomer
	assert.Equal(t, "my_customer", GetTableName(c3))

	var c4 []Customer
	assert.Equal(t, "customer", GetTableName(c4))

	var c5 *[]Customer
	assert.Equal(t, "customer", GetTableName(c5))

	var c6 []MyCustomer
	assert.Equal(t, "my_customer", GetTableName(c6))

	var c7 []CustomerPtr
	assert.Equal(t, "customer", GetTableName(c7))

	var c8 **int
	assert.Equal(t, "", GetTableName(c8))
}

type FA struct {
	A1 string
	A2 int
}

type FB struct {
	B1 string
}

func TestStructInfo(t *testing.T) {

	mExpect := map[string]string{
		"address": "Address", "email": "Email", "id": "ID", "name": "Name", "status": "Status",
	}

	var c Customer
	si, err := GetStructInfo(&c, DefaultFieldMapFunc)
	if err != nil {
		t.Fatal(err)
	}

	// pk
	sPK := si.PrimaryKeyFieldNames()
	assert.Equal(t, []string{"ID"}, sPK)

	// by db column name
	const tcname = "email"
	fi := si.GetFieldInfoByColumnName(tcname)
	if fi == nil {
		t.Fatalf(`GetFieldInfoByColumnName(%s) == nil`, tcname)
	}
	assert.Equal(t, mExpect[tcname], fi.FieldName())

	// by struct field name
	const tfname = "Email"
	fi = si.GetFieldInfoByFieldName(tfname)
	if fi == nil {
		t.Fatalf(`GetFieldInfoByFieldName(%s) == nil`, tfname)
	}
	assert.Equal(t, tcname, fi.ColumnName())

	// all field mappings
	for _, fi := range si.GetFieldInfo() {
		fldname := fi.FieldName()
		colname := fi.ColumnName()
		v, ok := mExpect[colname]
		if ok {
			assert.Equal(t, v, fldname)
		} else {
			t.Fatalf(`unexpected ColumnName() "%s"`, colname)
		}
	}
}
