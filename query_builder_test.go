// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dbx

import (
	"testing"
)

func TestQB_BuildSelect(t *testing.T) {
	tests := []struct {
		cols     []string
		distinct bool
		option   string
		expected string
	}{
		{[]string{}, false, "", "SELECT *"},
		{[]string{}, true, "CALC_ROWS", "SELECT DISTINCT CALC_ROWS *"},
		{[]string{"name", "DOB1"}, false, "", "SELECT `name`, `DOB1`"},
		{[]string{"name As Name", "users.last_name", "u.first1 first"}, false, "", "SELECT `name` AS `Name`, `users`.`last_name`, `u`.`first1` AS `first`"},
	}

	qb := getDB().QueryBuilder()
	for _, test := range tests {
		s := qb.BuildSelect(test.cols, test.distinct, test.option)
		if s != test.expected {
			t.Errorf("BuildSelect(%v, %v, %v) = %q, expected %q", test.cols, test.distinct, test.option, s, test.expected)
		}
	}
}

func TestQB_BuildFrom(t *testing.T) {
	tests := []struct {
		tables   []string
		expected string
	}{
		{[]string{}, ""},
		{[]string{"users"}, "FROM `users`"},
		{[]string{"users", "posts"}, "FROM `users`, `posts`"},
		{[]string{"users u", "posts as p"}, "FROM `users` `u`, `posts` `p`"},
		{[]string{"pub.users p.u", "posts AS p1"}, "FROM `pub`.`users` `p.u`, `posts` `p1`"},
	}

	qb := getDB().QueryBuilder()
	for _, test := range tests {
		s := qb.BuildFrom(test.tables)
		if s != test.expected {
			t.Errorf("BuildFrom(%v) = %q, expected %q", test.tables, s, test.expected)
		}
	}
}

func TestQB_BuildGroupBy(t *testing.T) {
	tests := []struct {
		cols     []string
		expected string
	}{
		{[]string{}, ""},
		{[]string{"name"}, "GROUP BY `name`"},
		{[]string{"name", "age"}, "GROUP BY `name`, `age`"},
	}

	qb := getDB().QueryBuilder()
	for _, test := range tests {
		s := qb.BuildGroupBy(test.cols)
		if s != test.expected {
			t.Errorf("BuildGroupBy(%v) = %q, expected %q", test.cols, s, test.expected)
		}
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
		if s != test.expected {
			t.Errorf("%v: BuildWhere() = %v, expected %v", test.tag, s, test.expected)
		}
		if len(params) != test.count {
			t.Errorf("%v: param count = %v, expected %v", test.tag, len(params), test.count)
		}
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
		if s != test.expected {
			t.Errorf("%v: BuildHaving() = %v, expected %v", test.tag, s, test.expected)
		}
		if len(params) != test.count {
			t.Errorf("%v: param count = %v, expected %v", test.tag, len(params), test.count)
		}
	}
}

func TestQB_BuildOrderBy(t *testing.T) {
	tests := []struct {
		cols     []string
		expected string
	}{
		{[]string{}, ""},
		{[]string{"name"}, "ORDER BY `name`"},
		{[]string{"name ASC", "age DESC", "id desc"}, "ORDER BY `name` ASC, `age` DESC, `id` desc"},
	}
	qb := getDB().QueryBuilder().(*BaseQueryBuilder)
	for _, test := range tests {
		s := qb.BuildOrderBy(test.cols)
		if s != test.expected {
			t.Errorf("BuildOrderBy(%v) = %v, expected %v", test.cols, s, test.expected)
		}
	}
}

func TestQB_BuildLimit(t *testing.T) {
	tests := []struct {
		limit, offset int64
		expected      string
	}{
		{10, -1, "LIMIT 10"},
		{10, 0, "LIMIT 10"},
		{10, 2, "LIMIT 10 OFFSET 2"},
		{0, 2, "LIMIT 0 OFFSET 2"},
		{-1, 2, "LIMIT 9223372036854775807 OFFSET 2"},
		{-1, 0, ""},
	}
	qb := getDB().QueryBuilder().(*BaseQueryBuilder)
	for _, test := range tests {
		s := qb.BuildLimit(test.limit, test.offset)
		if s != test.expected {
			t.Errorf("BuildLimit(%v, %v) = %v, expected %v", test.limit, test.offset, s, test.expected)
		}
	}
}

func TestQB_BuildOrderByAndLimit(t *testing.T) {
	qb := getDB().QueryBuilder()

	sql := qb.BuildOrderByAndLimit("SELECT *", []string{"name"}, 10, 2)
	expected := "SELECT *\nORDER BY `name`\nLIMIT 10 OFFSET 2"
	assertEqual(t, sql, expected, "t1")

	sql = qb.BuildOrderByAndLimit("SELECT *", nil, -1, -1)
	expected = "SELECT *"
	assertEqual(t, sql, expected, "t2")

	sql = qb.BuildOrderByAndLimit("SELECT *", []string{"name"}, -1, -1)
	expected = "SELECT *\nORDER BY `name`"
	assertEqual(t, sql, expected, "t3")

	sql = qb.BuildOrderByAndLimit("SELECT *", nil, 10, -1)
	expected = "SELECT *\nLIMIT 10"
	assertEqual(t, sql, expected, "t4")
}

func TestQB_BuildJoin(t *testing.T) {
	qb := getDB().QueryBuilder()

	params := Params{}
	ji := JoinInfo{"LEFT JOIN", "users u", NewExp("id=u.id", Params{"id":1})}
	sql := qb.BuildJoin([]JoinInfo{ji}, params)
	expected := "LEFT JOIN `users` `u` ON id=u.id"
	assertEqual(t, sql, expected, "BuildJoin@1")
	assertEqual(t, len(params), 1, "len(params)@1")

	params = Params{}
	ji = JoinInfo{"INNER JOIN", "users", nil}
	sql = qb.BuildJoin([]JoinInfo{ji}, params)
	expected = "INNER JOIN `users`"
	assertEqual(t, sql, expected, "BuildJoin@2")
	assertEqual(t, len(params), 0, "len(params)@2")

	sql = qb.BuildJoin([]JoinInfo{}, nil)
	expected = ""
	assertEqual(t, sql, expected, "BuildJoin@3")

	ji = JoinInfo{"INNER JOIN", "users", nil}
	ji2 := JoinInfo{"LEFT JOIN", "posts", nil}
	sql = qb.BuildJoin([]JoinInfo{ji, ji2}, nil)
	expected = "INNER JOIN `users`\nLEFT JOIN `posts`"
	assertEqual(t, sql, expected, "BuildJoin@3")
}

func TestQB_BuildUnion(t *testing.T) {
	db := getDB()
	qb := db.QueryBuilder()

	params := Params{}
	ui := UnionInfo{false, db.NewQuery("SELECT names").Bind(Params{"id":1})}
	sql := qb.BuildUnion([]UnionInfo{ui}, params)
	expected := "UNION (SELECT names)"
	assertEqual(t, sql, expected, "BuildUnion@1")
	assertEqual(t, len(params), 1, "len(params)@1")

	params = Params{}
	ui = UnionInfo{true, db.NewQuery("SELECT names")}
	sql = qb.BuildUnion([]UnionInfo{ui}, params)
	expected = "UNION ALL (SELECT names)"
	assertEqual(t, sql, expected, "BuildUnion@2")
	assertEqual(t, len(params), 0, "len(params)@2")

	sql = qb.BuildUnion([]UnionInfo{}, nil)
	expected = ""
	assertEqual(t, sql, expected, "BuildUnion@3")

	ui = UnionInfo{true, db.NewQuery("SELECT names")}
	ui2 := UnionInfo{false, db.NewQuery("SELECT ages")}
	sql = qb.BuildUnion([]UnionInfo{ui, ui2}, nil)
	expected = "UNION ALL (SELECT names)\nUNION (SELECT ages)"
	assertEqual(t, sql, expected, "BuildUnion@4")
}
