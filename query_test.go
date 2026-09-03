// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbx

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewQuery(t *testing.T) {
	db := NewFromDB(nil, "mysql")
	sql := "SELECT * FROM users WHERE id={:id}"
	q := NewQuery(db, db.sqlDB, sql)
	assert.Equal(t, q.SQL(), sql, "q.SQL()")
	assert.Equal(t, q.rawSQL, "SELECT * FROM users WHERE id=?", "q.RawSQL()")

	assert.Equal(t, len(q.Params()), 0, "len(q.Params())@1")
	q.Bind(Params{"id": 1})
	assert.Equal(t, len(q.Params()), 1, "len(q.Params())@2")
}

func TestQuery_logSQL(t *testing.T) {
	db := NewFromDB(nil, "mysql")
	q := db.NewQuery("SELECT * FROM users WHERE type={:type} AND id={:id}").Bind(Params{"type": "a", "id": 1})
	expected := "SELECT * FROM users WHERE type=\"a\" AND id=1"
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
