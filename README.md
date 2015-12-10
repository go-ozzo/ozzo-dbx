# ozzo-dbx

[![Build Status](https://travis-ci.org/go-ozzo/ozzo-dbx.svg?branch=master)](https://travis-ci.org/go-ozzo/ozzo-dbx)
[![GoDoc](https://godoc.org/github.com/go-ozzo/ozzo-dbx?status.png)](http://godoc.org/github.com/go-ozzo/ozzo-dbx)

ozzo-dbx is a Go package that enhances the standard `database/sql` package by providing powerful data retrieval methods
as well as DB-agnostic query building capabilities. It has the following features:

* Populating data into structs and NullString maps
* DB-agnostic query building, including SELECT queries, data manipulation queries, and schema manipulation queries
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
q := db.NewQuery("SELECT id, name FROM users LIMIT 10")

var (
	users []struct {
		ID   int
		Name string
	}

	row dbx.NullStringMap

	id   int
	name string

	err error
)

// populate all rows into users struct slice
err = q.All(&users)

// populate the first row into a NullString map
err = q.One(&row)

// populate the first row into id and name
err = q.Row(&id, &name)

// populate data row by row
rows, _ := q.Rows()
for rows.Next() {
	rows.ScanMap(&row)
}
```

When populating a struct, the data of a resulting column will be populated to an exported struct field if the name 
of the field maps to that of the column according to the field mapping function specified by `Query.FieldMapper`. 
For example,

If a resulting column does not have a corresponding struct field, it will be skipped without error. The default
field mapping function separates words in a field name by underscores and turns them into lower case. For example,
a field name `FirstName` will be mapped to the column name `first_name`, and `MyID` to `my_id`.

If a field has a `db` tag, the tag name will be used as the corresponding column name.

Note that only exported fields can be populated with data. Unexported fields are ignored. Anonymous struct fields
will be expanded and populated according to the above rules. Named struct fields will also be expanded, but their
sub-fields will be prefixed with the struct names. In the following example, the `User` type can be used to populate
data from a column named `prof.age`. If the `Profile` field is anonymous, it would be able to receive data from 
from a column named `age` without the prefix `prof.`.

```go
type Profile struct {
    Age int
}

type User struct {
    Prof Profile 
}
```



## Binding Parameters

## Building SELECT Queries

## Building Data Manipulation Queries

## Building Schema Manipulation Queries

## Using Transactions

## Handling Errors

## Logging Executed SQL Statements

## Supporting New Databases
