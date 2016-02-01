// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dbx

import (
	"testing"
)

func TestSelectQuery(t *testing.T) {
	db := getDB()

	// minimal select query
	q := db.Select().From("users").Build()
	expected := "SELECT * FROM `users`"
	assertEqual(t, q.SQL(), expected, "t1")
	assertEqual(t, len(q.Params()), 0, "t2")

	// a full select query
	q = db.Select("id", "name").
		AndSelect("age").
		Distinct(true).
		SelectOption("CALC").
		From("users").
		Where(NewExp("age>30")).
		AndWhere(NewExp("status=1")).
		OrWhere(NewExp("type=2")).
		InnerJoin("profile", NewExp("user.id=profile.id")).
		OrderBy("age DESC", "type").
		AndOrderBy("id").
		GroupBy("id").
		AndGroupBy("age").
		Having(NewExp("id>10")).
		AndHaving(NewExp("id<20")).
		OrHaving(NewExp("type=3")).
		Limit(10).
		Offset(20).
		Bind(Params{"id": 1}).
		AndBind(Params{"age": 30}).
		Build()

	expected = "SELECT DISTINCT CALC `id`, `name`, `age` FROM `users` INNER JOIN `profile` ON user.id=profile.id WHERE ((age>30) AND (status=1)) OR (type=2) GROUP BY `id`, `age` HAVING ((id>10) AND (id<20)) OR (type=3) ORDER BY `age` DESC, `type`, `id` LIMIT 10 OFFSET 20"
	assertEqual(t, q.SQL(), expected, "t3")
	assertEqual(t, len(q.Params()), 2, "t4")

	// union
	q1 := db.Select().From("users").Build()
	q2 := db.Select().From("posts").Build()
	q = db.Select().From("profiles").Union(q1).UnionAll(q2).Build()
	expected = "(SELECT * FROM `profiles`) UNION (SELECT * FROM `users`) UNION ALL (SELECT * FROM `posts`)"
	assertEqual(t, q.SQL(), expected, "t5")
}

func TestSelectQuery_Data(t *testing.T) {
	db := getPreparedDB()
	defer db.Close()

	q := db.Select("id", "email").From("customer").OrderBy("id")

	var customer Customer
	q.One(&customer)
	assertEqual(t, customer.Email, "user1@example.com", "customer.Email")

	var customers []Customer
	q.All(&customers)
	assertEqual(t, len(customers), 3, "len(customers)")

	rows, _ := q.Rows()
	customer.Email = ""
	rows.one(&customer)
	assertEqual(t, customer.Email, "user1@example.com", "customer.Email")

	var id, email string
	q.Row(&id, &email)
	assertEqual(t, id, "1", "id")
	assertEqual(t, email, "user1@example.com", "email")
}
