// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQB_BuildSelect(t *testing.T) {
	tests := []struct {
		tag      string
		cols     []string
		distinct bool
		option   string
		expected string
	}{
		{"empty", []string{}, false, "", "SELECT *"},
		{"empty distinct", []string{}, true, "CALC_ROWS", "SELECT DISTINCT CALC_ROWS *"},
		{"multi-columns", []string{"name", "DOB1"}, false, "", "SELECT `name`, `DOB1`"},
		{"aliased columns", []string{"name As Name", "users.last_name", "u.first1 first"}, false, "", "SELECT `name` AS `Name`, `users`.`last_name`, `u`.`first1` AS `first`"},
	}

	db := getDB()
	qb := db.QueryBuilder()
	for _, test := range tests {
		s := qb.BuildSelect(test.cols, test.distinct, test.option)
		assert.Equal(t, test.expected, s, test.tag)
	}
	assert.Equal(t, qb.(*BaseQueryBuilder).DB(), db)
}

func TestQB_BuildFrom(t *testing.T) {
	tests := []struct {
		tag      string
		tables   []string
		expected string
	}{
		{"empty", []string{}, ""},
		{"single table", []string{"users"}, "FROM `users`"},
		{"multiple tables", []string{"users", "posts"}, "FROM `users`, `posts`"},
		{"table alias", []string{"users u", "posts as p"}, "FROM `users` `u`, `posts` `p`"},
		{"table prefix and alias", []string{"pub.users p.u", "posts AS p1"}, "FROM `pub`.`users` `p.u`, `posts` `p1`"},
	}

	qb := getDB().QueryBuilder()
	for _, test := range tests {
		s := qb.BuildFrom(test.tables)
		assert.Equal(t, test.expected, s, test.tag)
	}
}

func TestQB_BuildGroupBy(t *testing.T) {
	tests := []struct {
		tag      string
		cols     []string
		expected string
	}{
		{"empty", []string{}, ""},
		{"single column", []string{"name"}, "GROUP BY `name`"},
		{"multiple columns", []string{"name", "age"}, "GROUP BY `name`, `age`"},
	}

	qb := getDB().QueryBuilder()
	for _, test := range tests {
		s := qb.BuildGroupBy(test.cols)
		assert.Equal(t, test.expected, s, test.tag)
	}
}

func TestQB_BuildWhere(t *testing.T) {
	tests := []struct {
		exp      Expression
		expected string
		count    int
		tag      string
	}{
		{HashExp{"age": 30, "dept": "marketing"}, "WHERE `age`={:p0} AND `dept`={:p1}", 2, "t1"},
		{nil, "", 0, "t2"},
		{NewExp(""), "", 0, "t3"},
	}

	qb := getDB().QueryBuilder()
	for _, test := range tests {
		params := Params{}
		s := qb.BuildWhere(test.exp, params)
		assert.Equal(t, test.expected, s, test.tag)
		assert.Equal(t, test.count, len(params), test.tag)
	}
}

func TestQB_BuildHaving(t *testing.T) {
	tests := []struct {
		exp      Expression
		expected string
		count    int
		tag      string
	}{
		{HashExp{"age": 30, "dept": "marketing"}, "HAVING `age`={:p0} AND `dept`={:p1}", 2, "t1"},
		{nil, "", 0, "t2"},
		{NewExp(""), "", 0, "t3"},
	}

	qb := getDB().QueryBuilder()
	for _, test := range tests {
		params := Params{}
		s := qb.BuildHaving(test.exp, params)
		assert.Equal(t, test.expected, s, test.tag)
		assert.Equal(t, test.count, len(params), test.tag)
	}
}

func TestQB_BuildOrderBy(t *testing.T) {
	tests := []struct {
		tag      string
		cols     []string
		expected string
	}{
		{"empty", []string{}, ""},
		{"single column", []string{"name"}, "ORDER BY `name`"},
		{"multiple columns", []string{"name ASC", "age DESC", "id desc"}, "ORDER BY `name` ASC, `age` DESC, `id` desc"},
	}
	qb := getDB().QueryBuilder().(*BaseQueryBuilder)
	for _, test := range tests {
		s := qb.BuildOrderBy(test.cols)
		assert.Equal(t, test.expected, s, test.tag)
	}
}

func TestQB_BuildLimit(t *testing.T) {
	tests := []struct {
		tag           string
		limit, offset int64
		expected      string
	}{
		{"t1", 10, -1, "LIMIT 10"},
		{"t2", 10, 0, "LIMIT 10"},
		{"t3", 10, 2, "LIMIT 10 OFFSET 2"},
		{"t4", 0, 2, "LIMIT 0 OFFSET 2"},
		{"t5", -1, 2, "LIMIT 9223372036854775807 OFFSET 2"},
		{"t6", -1, 0, ""},
	}
	qb := getDB().QueryBuilder().(*BaseQueryBuilder)
	for _, test := range tests {
		s := qb.BuildLimit(test.limit, test.offset)
		assert.Equal(t, test.expected, s, test.tag)
	}
}

func TestQB_BuildOrderByAndLimit(t *testing.T) {
	qb := getDB().QueryBuilder()

	sql := qb.BuildOrderByAndLimit("SELECT *", []string{"name"}, 10, 2)
	expected := "SELECT * ORDER BY `name` LIMIT 10 OFFSET 2"
	assert.Equal(t, sql, expected, "t1")

	sql = qb.BuildOrderByAndLimit("SELECT *", nil, -1, -1)
	expected = "SELECT *"
	assert.Equal(t, sql, expected, "t2")

	sql = qb.BuildOrderByAndLimit("SELECT *", []string{"name"}, -1, -1)
	expected = "SELECT * ORDER BY `name`"
	assert.Equal(t, sql, expected, "t3")

	sql = qb.BuildOrderByAndLimit("SELECT *", nil, 10, -1)
	expected = "SELECT * LIMIT 10"
	assert.Equal(t, sql, expected, "t4")
}

func TestQB_BuildJoin(t *testing.T) {
	qb := getDB().QueryBuilder()

	params := Params{}
	ji := JoinInfo{"LEFT JOIN", "users u", NewExp("id=u.id", Params{"id": 1})}
	sql := qb.BuildJoin([]JoinInfo{ji}, params)
	expected := "LEFT JOIN `users` `u` ON id=u.id"
	assert.Equal(t, sql, expected, "BuildJoin@1")
	assert.Equal(t, len(params), 1, "len(params)@1")

	params = Params{}
	ji = JoinInfo{"INNER JOIN", "users", nil}
	sql = qb.BuildJoin([]JoinInfo{ji}, params)
	expected = "INNER JOIN `users`"
	assert.Equal(t, sql, expected, "BuildJoin@2")
	assert.Equal(t, len(params), 0, "len(params)@2")

	sql = qb.BuildJoin([]JoinInfo{}, nil)
	expected = ""
	assert.Equal(t, sql, expected, "BuildJoin@3")

	ji = JoinInfo{"INNER JOIN", "users", nil}
	ji2 := JoinInfo{"LEFT JOIN", "posts", nil}
	sql = qb.BuildJoin([]JoinInfo{ji, ji2}, nil)
	expected = "INNER JOIN `users` LEFT JOIN `posts`"
	assert.Equal(t, sql, expected, "BuildJoin@3")
}

func TestQB_BuildUnion(t *testing.T) {
	db := getDB()
	qb := db.QueryBuilder()

	params := Params{}
	ui := UnionInfo{false, db.NewQuery("SELECT names").Bind(Params{"id": 1})}
	sql := qb.BuildUnion([]UnionInfo{ui}, params)
	expected := "UNION (SELECT names)"
	assert.Equal(t, sql, expected, "BuildUnion@1")
	assert.Equal(t, len(params), 1, "len(params)@1")

	params = Params{}
	ui = UnionInfo{true, db.NewQuery("SELECT names")}
	sql = qb.BuildUnion([]UnionInfo{ui}, params)
	expected = "UNION ALL (SELECT names)"
	assert.Equal(t, sql, expected, "BuildUnion@2")
	assert.Equal(t, len(params), 0, "len(params)@2")

	sql = qb.BuildUnion([]UnionInfo{}, nil)
	expected = ""
	assert.Equal(t, sql, expected, "BuildUnion@3")

	ui = UnionInfo{true, db.NewQuery("SELECT names")}
	ui2 := UnionInfo{false, db.NewQuery("SELECT ages")}
	sql = qb.BuildUnion([]UnionInfo{ui, ui2}, nil)
	expected = "UNION ALL (SELECT names) UNION (SELECT ages)"
	assert.Equal(t, sql, expected, "BuildUnion@4")
}
