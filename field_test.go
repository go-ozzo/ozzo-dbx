// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dbx

import (
	"testing"
	"reflect"
	"encoding/json"
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
		if r != test.output {
			t.Errorf("MapField(%q) = %q, expected %q", test.input, r, test.output)
		}
	}
}

type FA struct {
	A1 string
	A2 int
}

type FB struct {
	B1 string
}

func TestGetFieldMap(t *testing.T) {
	var a struct {
		X1 string
		FA
		X2 int
		B *FB
		FB `db:"c"`
		c int
	}
	ta := reflect.TypeOf(a)
	r := getFieldMap(ta, DefaultFieldMapFunc)

	v, _ := json.Marshal(r)
	assertEqual(t, string(v), `{"a1":[1,0],"a2":[1,1],"b.b1":[3,0],"c.b1":[4,0],"x1":[0],"x2":[2]}`)
}
