# ozzo-dbx

[![Build Status](https://travis-ci.org/go-ozzo/ozzo-dbx.svg?branch=master)](https://travis-ci.org/go-ozzo/ozzo-dbx)
[![GoDoc](https://godoc.org/github.com/go-ozzo/ozzo-dbx?status.png)](http://godoc.org/github.com/go-ozzo/ozzo-dbx)

ozzo-dbx это Go-пакет, который расширяет стандартный `database/sql` пакет путем предоставления мощных методов извлечения данных,
а также позволяет использовать DB-agnostic (независимый от БД) построитель запросов. Он имеет следующие особенности:

* Заполнеие данных в структуры или NullString карты
* Именованные параметры связывания
* DB-agnostic методы построения запросов, включая SELECT, запросы на манипуляцию с данными и запросы на манипуляцию со схемой БД
* Мощный построитель запросов с условиями
* Открытая архитектура, позволяющая с легкостью создавать поддержку для новых баз данных или кастомизировать текущую поддержку 
* Логировать выполненные SQL запросы
* Предоставляет роддержку основных реляционных баз данных

## Требования

Go 1.2 или выше.

## Установка

Выполните данные команды для установки:

```
go get github.com/go-ozzo/ozzo-dbx
```

В дополнение, установите необходимый пакт драйвера базы данных для той базы, которая будет использоваться. Пожалуйста обратитесь к 
[SQL database drivers](https://github.com/golang/go/wiki/SQLDrivers) для получения полного списка. Например, если вы используете
MySQL, вы можете загрузить этот пакет:

```
go get github.com/go-sql-driver/mysql
```

и импортировать его в ваш основной код следующим образом:

```go
import _ "github.com/go-sql-driver/mysql"
```

## Поддерживаемые базы данных

Представленные базы данных уже имеют поддержку из коробки:

* SQLite
* MySQL
* PostgreSQL
* MS SQL Server (2012 или выше)
* Oracle

Другие базы данных также могут работать. Если нет, вы можете создать для него свой построитель, как описано далее в этом документе.


## С чего начать

Представленный ниже код показывает как вы можете использовать пакет для доступа данным базы данных MySQL.

```go
import (
	"fmt"
	"github.com/go-ozzo/ozzo-dbx"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, _ := dbx.Open("mysql", "user:pass@/example")

	// создаем новый запрос
	q := db.NewQuery("SELECT id, name FROM users LIMIT 10")

	// извлекаем все строки в массив структур
	var users []struct {
		ID, Name string
	}
	q.All(&users)

	// извлекаем одну строку в структуру
	var user struct {
		ID, Name string
	}
	q.One(&user)

	// извлекаем строку в строковую карту
	data := dbx.NullStringMap{}
	q.One(data)

	// извлечение строки за строкой
	rows2, _ := q.Rows()
	for rows2.Next() {
		rows2.ScanStruct(&user)
		// rows.ScanMap(data)
		// rows.Scan(&id, &name)
	}
}
```

Следующий пример показывает, как можно использовать возможности построителя запросов из этого пакета.

```go
import (
	"fmt"
	"github.com/go-ozzo/ozzo-dbx"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, _ := dbx.Open("mysql", "user:pass@/example")

	// строим SELECT запрос
	//   SELECT `id`, `name` FROM `users` WHERE `name` LIKE '%Charles%' ORDER BY `id`
	q := db.Select("id", "name").
		From("users").
		Where(dbx.Like("name", "Charles")).
		OrderBy("id")

	// извлекаем все строки в массив структур
	var users []struct {
		ID, Name string
	}
	q.All(&users)

	// строим INSERT запрос
	//   INSERT INTO `users` (`name`) VALUES ('James')
	db.Insert("users", dbx.Params{
		"name": "James",
	}).Execute()
}
```

## Соединение с базой данных

Для соединения с базой данных, используйте `dbx.Open()` таким же образом, как вы могли сделать это при помощи `Open()` метода в `database/sql`.

```go
db, err := dbx.Open("mysql", "user:pass@hostname/db_name")
```

Метод возвращает экземпляр `dbx.DB` который можно использовать для создания и выполнения запросов к БД. Обратите внимание, 
что метод на самом деле не устанавливает соединение до тех пор пока вы не выполните запрос с использованием экземпляра `dbx.DB`. Он так же 
не проверяет корректность имени источника данных. Выполнните `dbx.MustOpen()` для того чтобы убедиться, что имя базы данных правильное.

## Выполнение запросов

Для выполнения SQL запроса, сначала создайте экземпляр `dbx.Query`, и далее создайте запрос при помощи `DB.NewQuery()` с SQL выражением, 
которое нужно исполнить. И затем выполните `Query.Execute()` для исполнения запроса, в случае если запрос не предназначен для извлечения данных.
Например,

```go
q := db.NewQuery("UPDATE users SET status=1 WHERE id=100")
result, err := q.Execute()
```

Если SQL запрос должен вернуть данные (такой как SELECT), следует вызвать один из нижепреведенных методов, 
которые выполнят запрос и сохранят результат в указанной переменной/переменных.

* `Query.All()`: заполняет массив структур или `NullString` карты всеми строками результата.
* `Query.One()`: сохраняет первую строку из результата в структуру или в `NullString` карту.
* `Query.Row()`: заполняет список переменных первой строкой результата, каждя перменная заполняется данными одной из возвращаемых колонок.
* `Query.Rows()`: возвращает экземпляр `dbx.Rows` для возможности извлечения данных строка за строкой.

Например,

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

* `User.id`: not populated because the `db` tag is `-`;
* `User.Type`: not populated because the field is not exported;
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
