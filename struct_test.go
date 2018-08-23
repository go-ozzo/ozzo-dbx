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

	var c1 Customer = Customer{}

	vc := indirect(reflect.ValueOf(&c1))
	assert.Equal(t, reflect.Struct, vc.Kind())
	assert.Equal(t, reflect.Struct, vc.Type().Kind())
	assert.True(t, vc.CanSet())
}

func Test_structValue_columns(t *testing.T) {
	customer := Customer{
		ID:     1,
		Name:   "abc",
		Status: 2,
		Email:  "abc@example.com",
	}
	sv := newStructValue(&customer, DefaultFieldMapFunc, DefaultTableMapFunc)
	cols := sv.columns(nil, nil)
	assert.Equal(t, map[string]interface{}{"id": 1, "name": "abc", "status": 2, "email": "abc@example.com", "address": sql.NullString{}}, cols)

	cols = sv.columns([]string{"ID", "name"}, nil)
	assert.Equal(t, map[string]interface{}{"id": 1}, cols)

	cols = sv.columns([]string{"ID", "Name"}, []string{"ID"})
	assert.Equal(t, map[string]interface{}{"name": "abc"}, cols)

	cols = sv.columns(nil, []string{"ID", "Address"})
	assert.Equal(t, map[string]interface{}{"name": "abc", "status": 2, "email": "abc@example.com"}, cols)

	sv = newStructValue(&customer, nil, DefaultTableMapFunc)
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
	} {customer, "20"}
	sv := newStructValue(&ev, nil, DefaultTableMapFunc)
	cols := sv.columns([]string{"ID", "Status"}, nil)
	assert.Equal(t, map[string]interface{}{"ID": 1, "Status": "20"}, cols)

	ev2 := struct {
		Status string
		Customer
	}{"20", customer}
	sv = newStructValue(&ev2, nil)
	cols = sv.columns([]string{"ID", "Status"}, nil)
	assert.Equal(t, map[string]interface{}{"ID": 1, "Status": "20"}, cols)
}

type MyCustomer struct{}

type SomeTable struct{}

func (*SomeTable) TableName() string {
	return "strange_name"
}

func Test_getTableName(t *testing.T) {
	{
		var c Customer
		assert.Equal(t, "customer", DefaultTableMapFunc(c))
	}

	{
		var c *Customer
		assert.Equal(t, "customer", DefaultTableMapFunc(c))
	}

	{
		var c MyCustomer
		assert.Equal(t, "my_customer", DefaultTableMapFunc(c))
	}

	{
		var c []Customer
		assert.Equal(t, "customer", DefaultTableMapFunc(c))
	}

	{
		var c *[]Customer
		assert.Equal(t, "customer", DefaultTableMapFunc(c))
	}

	{
		var c []*Customer
		assert.Equal(t, "customer", DefaultTableMapFunc(c))
	}

	{
		var c []MyCustomer
		assert.Equal(t, "my_customer", DefaultTableMapFunc(c))
	}

	{
		var c []CustomerPtr
		assert.Equal(t, "customer", DefaultTableMapFunc(c))
	}

	{
		var c **int
		assert.Equal(t, "", DefaultTableMapFunc(c))
	}

	{
		var c ***[]Customer
		assert.Equal(t, "customer", DefaultTableMapFunc(c))
	}

	{
		func(i interface{}) {
			func(c interface{}) {
				assert.Equal(t, "customer", DefaultTableMapFunc(c))
			}(&i)
		}(&Customer{})
	}

	{
		func(i interface{}) {
			func(c interface{}) {
				assert.Equal(t, "customer", DefaultTableMapFunc(&c))
			}(&i)
		}(&Customer{})
	}

	{
		var c *SomeTable
		assert.Equal(t, "strange_name", DefaultTableMapFunc(c))
	}

	{
		var c **SomeTable
		assert.Equal(t, "strange_name", DefaultTableMapFunc(c))
	}
}

type FA struct {
	A1 string
	A2 int
}

type FB struct {
	B1 string
}
