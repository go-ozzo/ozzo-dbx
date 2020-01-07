package dbx

import (
	"fmt"
)

// This example shows how to populate DB data in different ways.
func Example_dbQueries() {
	db, _ := Open("mysql", "user:pass@/example")

	// create a new query
	q := db.NewQuery("SELECT id, name FROM users LIMIT 10")

	// fetch all rows into a struct array
	var users []struct {
		ID, Name string
	}
	q.All(&users)

	// fetch a single row into a struct
	var user struct {
		ID, Name string
	}
	q.One(&user)

	// fetch a single row into a string map
	data := NullStringMap{}
	q.One(data)

	// fetch row by row
	rows2, _ := q.Rows()
	for rows2.Next() {
		rows2.ScanStruct(&user)
		// rows.ScanMap(data)
		// rows.Scan(&id, &name)
	}
}

// This example shows how to use query builder to build DB queries.
func Example_queryBuilder() {
	db, _ := Open("mysql", "user:pass@/example")

	// build a SELECT query
	//   SELECT `id`, `name` FROM `users` WHERE `name` LIKE '%Charles%' ORDER BY `id`
	q := db.Select("id", "name").
		From("users").
		Where(Like("name", "Charles")).
		OrderBy("id")

	// fetch all rows into a struct array
	var users []struct {
		ID, Name string
	}
	q.All(&users)

	// build an INSERT query
	//   INSERT INTO `users` (name) VALUES ('James')
	db.Insert("users", Params{
		"name": "James",
	}).Execute()
}

// This example shows how to use query builder in transactions.
func Example_transactions() {
	db, _ := Open("mysql", "user:pass@/example")

	db.Transactional(func(tx *Tx) error {
		_, err := tx.Insert("user", Params{
			"name": "user1",
		}).Execute()
		if err != nil {
			return err
		}
		_, err = tx.Insert("user", Params{
			"name": "user2",
		}).Execute()
		return err
	})
}

type TestCustomer struct {
	ID   string
	Name string
}

// This example shows how to do CRUD operations.
func Example_crudOperations() {
	db, _ := Open("mysql", "user:pass@/example")

	var customer TestCustomer

	// read a customer: SELECT * FROM customer WHERE id=100
	db.Select().Model(100, &customer)

	// create a customer: INSERT INTO customer (name) VALUES ('test')
	db.Model(&customer).Insert()

	// update a customer: UPDATE customer SET name='test' WHERE id=100
	db.Model(&customer).Update()

	// delete a customer: DELETE FROM customer WHERE id=100
	db.Model(&customer).Delete()
}

func ExampleSchemaBuilder() {
	db, _ := Open("mysql", "user:pass@/example")

	db.Insert("users", Params{
		"name": "James",
		"age":  30,
	}).Execute()
}

func ExampleRows_ScanMap() {
	db, _ := Open("mysql", "user:pass@/example")

	user := NullStringMap{}

	sql := "SELECT id, name FROM users LIMIT 10"
	rows, _ := db.NewQuery(sql).Rows()
	for rows.Next() {
		rows.ScanMap(user)
		// ...
	}
}

func ExampleRows_ScanStruct() {
	db, _ := Open("mysql", "user:pass@/example")

	var user struct {
		ID, Name string
	}

	sql := "SELECT id, name FROM users LIMIT 10"
	rows, _ := db.NewQuery(sql).Rows()
	for rows.Next() {
		rows.ScanStruct(&user)
		// ...
	}
}

func ExampleQuery_All() {
	db, _ := Open("mysql", "user:pass@/example")
	sql := "SELECT id, name FROM users LIMIT 10"

	// fetches data into a slice of struct
	var users []struct {
		ID, Name string
	}
	db.NewQuery(sql).All(&users)

	// fetches data into a slice of NullStringMap
	var users2 []NullStringMap
	db.NewQuery(sql).All(&users2)
	for _, user := range users2 {
		fmt.Println(user["id"].String, user["name"].String)
	}
}

func ExampleQuery_One() {
	db, _ := Open("mysql", "user:pass@/example")
	sql := "SELECT id, name FROM users LIMIT 10"

	// fetches data into a struct
	var user struct {
		ID, Name string
	}
	db.NewQuery(sql).One(&user)

	// fetches data into a NullStringMap
	var user2 NullStringMap
	db.NewQuery(sql).All(user2)
	fmt.Println(user2["id"].String, user2["name"].String)
}

func ExampleQuery_Row() {
	db, _ := Open("mysql", "user:pass@/example")
	sql := "SELECT id, name FROM users LIMIT 10"

	// fetches data into a struct
	var (
		id   int
		name string
	)
	db.NewQuery(sql).Row(&id, &name)
}

func ExampleQuery_Rows() {
	var user struct {
		ID, Name string
	}

	db, _ := Open("mysql", "user:pass@/example")
	sql := "SELECT id, name FROM users LIMIT 10"

	rows, _ := db.NewQuery(sql).Rows()
	for rows.Next() {
		rows.ScanStruct(&user)
		// ...
	}
}

func ExampleQuery_Bind() {
	var user struct {
		ID, Name string
	}

	db, _ := Open("mysql", "user:pass@/example")
	sql := "SELECT id, name FROM users WHERE age>{:age} AND status={:status}"

	q := db.NewQuery(sql)
	q.Bind(Params{"age": 30, "status": 1}).One(&user)
}

func ExampleQuery_Prepare() {
	var users1, users2, users3 []struct {
		ID, Name string
	}

	db, _ := Open("mysql", "user:pass@/example")
	sql := "SELECT id, name FROM users WHERE age>{:age} AND status={:status}"

	q := db.NewQuery(sql).Prepare()
	defer q.Close()

	q.Bind(Params{"age": 30, "status": 1}).All(&users1)
	q.Bind(Params{"age": 20, "status": 1}).All(&users2)
	q.Bind(Params{"age": 10, "status": 1}).All(&users3)
}

func ExampleDB() {
	db, _ := Open("mysql", "user:pass@/example")

	// queries data through a plain SQL
	var users []struct {
		ID, Name string
	}
	db.NewQuery("SELECT id, name FROM users WHERE age=30").All(&users)

	// queries data using query builder
	db.Select("id", "name").From("users").Where(HashExp{"age": 30}).All(&users)

	// executes a plain SQL
	db.NewQuery("INSERT INTO users (name) SET ({:name})").Bind(Params{"name": "James"}).Execute()

	// executes a SQL using query builder
	db.Insert("users", Params{"name": "James"}).Execute()
}

func ExampleDB_Open() {
	db, err := Open("mysql", "user:pass@/example")
	if err != nil {
		panic(err)
	}

	var users []NullStringMap
	if err := db.NewQuery("SELECT * FROM users LIMIT 10").All(&users); err != nil {
		panic(err)
	}
}

func ExampleDB_Begin() {
	db, _ := Open("mysql", "user:pass@/example")

	tx, _ := db.Begin()

	_, err1 := tx.Insert("user", Params{
		"name": "user1",
	}).Execute()
	_, err2 := tx.Insert("user", Params{
		"name": "user2",
	}).Execute()

	if err1 == nil && err2 == nil {
		tx.Commit()
	} else {
		tx.Rollback()
	}
}
