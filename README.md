# ozzo-dbx

[![Build Status](https://travis-ci.org/go-ozzo/ozzo-dbx.svg?branch=master)](https://travis-ci.org/go-ozzo/ozzo-dbx)
[![GoDoc](https://godoc.org/github.com/go-ozzo/ozzo-dbx?status.png)](http://godoc.org/github.com/go-ozzo/ozzo-dbx)

ozzo-dbx is a Go package that enhances the standard `database/sql` package by providing powerful data retrieval methods
as well as DB-agnostic query building capabilities. It has the following features:

* Populating data into structs and NullString maps
* Named parameter binding
* DB-agnostic query building methods, including SELECT queries, data manipulation queries, and schema manipulation queries
* Powerful query condition building
* Open architecture allowing addition of new database support or customization of existing support
* Logging executed SQL statements
* Supporting major relational databases

## Requirements

Go 1.2 or above.

## Installation

Run the following command to install the package:

```
go get github.com/go-ozzo/ozzo-dbx
```

In addition, install the specific DB driver package for the kind of database to be used. Please refer to
[SQL database drivers](https://github.com/golang/go/wiki/SQLDrivers) for a complete list. For example, if you are
using MySQL, you may install the following package:

```
go get github.com/go-sql-driver/mysql
```

and import it in your main code like the following:

```go
import _ "github.com/go-sql-driver/mysql"
```

## Supported Databases

The following databases are supported out of box:

* SQLite
* MySQL
* PostgreSQL
* MS SQL Server (2012 or above)
* Oracle

Other databases may also work. If not, you may create a builder for it, as explained later in this document. 


## Getting Started

The following code snippet shows how you can use this package in order to access data from a MySQL database.

```go
import (
	"fmt"
	"github.com/go-ozzo/ozzo-dbx"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, _ := dbx.Open("mysql", "user:pass@/example")

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
	data := dbx.NullStringMap{}
	q.One(data)

	// fetch row by row
	rows2, _ := q.Rows()
	for rows2.Next() {
		rows2.ScanStruct(&user)
		// rows.ScanMap(data)
		// rows.Scan(&id, &name)
	}
}
```

And the following example shows how to use the query building capability of this package.

```go
import (
	"fmt"
	"github.com/go-ozzo/ozzo-dbx"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, _ := dbx.Open("mysql", "user:pass@/example")

	// build a SELECT query
	//   SELECT `id`, `name` FROM `users` WHERE `name` LIKE '%Charles%' ORDER BY `id`
	q := db.Select("id", "name").
		From("users").
		Where(dbx.Like("name", "Charles")).
		OrderBy("id")

	// fetch all rows into a struct array
	var users []struct {
		ID, Name string
	}
	q.All(&users)

	// build an INSERT query
	//   INSERT INTO `users` (`name`) VALUES ('James')
	db.Insert("users", dbx.Params{
		"name": "James",
	}).Execute()
}
```

## Connecting to Database

To connect to a database, call `dbx.Open()` in the same way as you would do with the `Open()` method in `database/sql`.

```go
db, err := dbx.Open("mysql", "user:pass@hostname/db_name")
```

The method returns a `dbx.DB` instance which can be used to create and execute DB queries. Note that the method 
does not really establish a connection until a query is made using the returned `dbx.DB` instance. It also
does not check the correctness of the data source name either. Call `dbx.MustOpen()` to make sure the data 
source name is correct.

## Executing Queries

To execute a SQL statement, first create a `dbx.Query` instance by calling `DB.NewQuery()` with the SQL statement
to be executed. And then call `Query.Execute()` to execute the query if the query is not meant to retrieving data.
For example,

```go
q := db.NewQuery("UPDATE users SET status=1 WHERE id=100")
result, err := q.Execute()
```

If the SQL statement does retrieve data (e.g. a SELECT statement), one of the following methods should be called, 
which will execute the query and populate the result into the specified variable(s).

* `Query.All()`: populate all rows of the result into a slice of structs or `NullString` maps.
* `Query.One()`: populate the first row of the result into a struct or a `NullString` map.
* `Query.Row()`: populate the first row of the result into a list of variables, one for each returning column.
* `Query.Rows()`: returns a `dbx.Rows` instance to allow retrieving data row by row.

For example,

```go
type User struct {
	ID   int
	Name string
}

var (
	users []User
	user User

	row dbx.NullStringMap

	id   int
	name string

	err error
)

q := db.NewQuery("SELECT id, name FROM users LIMIT 10")

// populate all rows into a User slice
err = q.All(&users)
fmt.Println(users[0].ID, users[0].Name)

// populate the first row into a User struct
err = q.One(&user)
fmt.Println(user.ID, user.Name)

// populate the first row into a NullString map
err = q.One(&row)
fmt.Println(row["id"], row["name"])

// populate the first row into id and name
err = q.Row(&id, &name)

// populate data row by row
rows, _ := q.Rows()
for rows.Next() {
	rows.ScanMap(&row)
}
```

When populating a struct, the following rules are used to determine which columns should go into which struct fields:

* Only exported struct fields can be populated.
* A field receives data if its name is mapped to a column according to the field mapping function `Query.FieldMapper`.
  The default field mapping function separates words in a field name by underscores and turns them into lower case.
  For example, a field name `FirstName` will be mapped to the column name `first_name`, and `MyID` to `my_id`.
* If a field has a `db` tag, the tag value will be used as the corresponding column name. If the `db` tag is a dash `-`,
  it means the field should NOT be populated.
* For anonymous fields that are of struct type, they will be expanded and their component fields will be populated
  according to the rules described above.
* For named fields that are of struct type, they will also be expanded. But their component fields will be prefixed
  with the struct names when being populated.

The following example shows how fields are populated according to the rules above:

```go
type User struct {
	id     int
	Type   int `db:"-"`
	MyName string `db:"name"`
	Prof   Profile
}

type Profile struct {
	Age int
}
```

* `User.id`: not populated because the field is not exported;
* `User.Type`: not populated because the `db` tag is `-`;
* `User.MyName`: to be populated from the `name` column, according to the `db` tag;
* `Profile.Age`: to be populated from the `prof.age` column, since `Prof` is a named field of struct type
  and its fields will be prefixed with `prof.`.

Note that if a column in the result does not have a corresponding struct field, it will be ignored. Similarly,
if a struct field does not have a corresponding column in the result, it will not be populated.

## Binding Parameters

A SQL statement is usually parameterized with dynamic values. For example, you may want to select the user record
according to the user ID received from the client. Parameter binding should be used in this case, and it is almost
always preferred for security reason. Unlike `database/sql` which does anonymous parameter binding, `ozzo-dbx` uses
named parameter binding. For example,

```go
q := db.NewQuery("SELECT id, name FROM users WHERE id={:id}")
q.Bind(dbx.Params{"id": 100})
q.One(&user)
```

The above example will select the user record whose `id` is 100. The method `Query.Bind()` binds a set
of named parameters to a SQL statement which contains parameter placeholders in the format of `{:ParamName}`.

If a SQL statement needs to be executed multiple times with different parameter values, it should be prepared
to improve the performance. For example,

```go
q := db.NewQuery("SELECT id, name FROM users WHERE id={:id}")
q.Prepare()

q.Bind(dbx.Params{"id": 100})
q.One(&user)

q.Bind(dbx.Params{"id": 200})
q.One(&user)

// ...
```

Note that anonymous parameter binding is not supported as it will mess up with named parameters.

## Building Queries

Instead of writing plain SQLs, `ozzo-dbx` allows you to build SQLs programmatically, which often leads to cleaner,
more secure, and DB-agnostic code. You can build three types of queries: the SELECT queries, the data manipulation
queries, and the schema manipulation queries.

### Building SELECT Queries

Building a SELECT query starts by calling `DB.Select()`. You can build different clauses of a SELECT query using
the corresponding query building methods. For example,

```go
db, _ := dbx.Open("mysql", "user:pass@/example")
db.Select("id", "name").
	From("users").
	Where(dbx.HashExp{"id": 100}).
	One(&user)
```

The above code will generate and execute the following SQL statement:

```sql
SELECT `id`, `name`
FROM `users`
WHERE `id`={:p0}
```

Notice how the table and column names are properly quoted according to the currently using database type.
And parameter binding is used to populate the value of `p0` in the `WHERE` clause.

`dbx-ozzo` supports very flexible and powerful query condition building which can be used to build SQL clauses
such as `WHERE`, `HAVING`, etc. For example,

```go
// id=100
dbx.NewExp("id={:id}", Params{"id": 100})

// id=100 AND status=1
dbx.HashExp{"id": 100, "status": 1}

// status=1 OR age>30
dbx.Or(dbx.HashExp{"status": 1}, dbx.NewExp("age>30"))

// name LIKE '%admin%' AND name LIKE '%example%'
dbx.Like("name", "admin", "example")
```

### Building Data Manipulation Queries

Data manipulation queries are those changing the data in the database, such as INSERT, UPDATE, DELETE statements.
Such queries can be built by calling the corresponding methods of `DB`. For example,

```go
db, _ := dbx.Open("mysql", "user:pass@/example")

// INSERT INTO `users` (`name`, `email`) VALUES ({:p0}, {:p1})
db.Insert("users", dbx.Params{
	"name": "James",
	"email": "james@example.com",
}).Execute()

// UPDATE `users` SET `status`={:p0} WHERE `id`={:p1}
db.Update("users", dbx.Params{"status": 1}, dbx.HashExp{"id": 100}).Execute()

// DELETE FROM `users` WHERE `status`={:p0}
db.Delete("users", dbx.HashExp{"status": 2}).Execute()
```

When building data manipulation queries, remember to call `Execute()` at the end to execute the queries.

### Building Schema Manipulation Queries

Schema manipulation queries are those changing the database schema, such as creating a new table, adding a new column.
These queries can be built by calling the corresponding methods of `DB`. For example,

```go
db, _ := dbx.Open("mysql", "user:pass@/example")

// CREATE TABLE `users` (`id` int primary key, `name` varchar(255))
q := db.CreateTable("users", map[string]string{
	"id": "int primary key",
	"name": "varchar(255)",
})
q.Execute()
```

## Quoting Table and Column Names

Databases vary in quoting table and column names. To allow writing DB-agnostic SQLs, ozzo-dbx introduces a special
syntax in quoting table and column names. A word enclosed within `{{` and `}}` is treated as a table name and will
be quoted according to the particular DB driver. Similarly, a word enclosed within `[[` and `]]` is treated as a 
column name and will be quoted accordingly as well. For example, when working with a MySQL database, the following
query will be properly quoted:

```go
// SELECT * FROM `users` WHERE `status`=1
q := db.NewQuery("SELECT * FROM {{users}} WHERE [[status]]=1")
```

Note that if a table or column name contains a prefix, it will still be properly quoted. For example, `{{public.users}}`
will be quoted as `"public"."users"` for PostgreSQL.

## Using Transactions

You can use all aforementioned query execution and building methods with transaction. For example,

```go
db, _ := dbx.Open("mysql", "user:pass@/example")

tx, _ := db.Begin()

_, err1 := tx.Insert("users", dbx.Params{
	"name": "user1",
}).Execute()
_, err2 := tx.Insert("users", dbx.Params{
	"name": "user2",
}).Execute()

if err1 == nil && err2 == nil {
	tx.Commit()
} else {
	tx.Rollback()
}
```

## Logging Executed SQL Statements

When `DB.LogFunc` is configured with a compatible log function, all SQL statements being executed will be logged.
The following example shows how to configure the logger using the standard `log` package:

```go
import (
	"fmt"
	"log"
	"github.com/go-ozzo/ozzo-dbx"
)

func main() {
	db, _ := dbx.Open("mysql", "user:pass@/example")
	db.LogFunc = log.Printf

	// ...
)
```

And the following example shows how to use the `ozzo-log` package which allows logging message severities and categories:

```go
import (
	"fmt"
	"github.com/go-ozzo/ozzo-dbx"
	"github.com/go-ozzo/ozzo-log"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	logger := log.NewLogger()
	logger.Targets = []log.Target{log.NewConsoleTarget()}
	logger.Open()

	db, _ := dbx.Open("mysql", "user:pass@/example")
	db.LogFunc = logger.Info

	// ...
)
```

## Supporting New Databases

While `ozzo-dbx` provides out-of-box support for most major relational databases, its open architecture
allows you to add support for new databases. The effort of adding support for a new database involves:

* Create a struct that implements the `QueryBuilder` interface. You may use `BaseQueryBuilder` directly or extend it
  via composition.
* Create a struct that implements the `Builder` interface. You may extend `BaseBuilder` via composition.
* Write an `init()` function to register the new builder in `dbx.BuilderFuncMap`.
