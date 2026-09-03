package dbx

import "database/sql"

// Test types shared across unit test files.

type Customer struct {
	ID      int
	Email   string
	Status  int
	Name    string
	Address sql.NullString
}

func (m Customer) TableName() string {
	return "customer"
}

type CustomerPtr struct {
	ID      *int `db:"pk"`
	Email   *string
	Status  *int
	Name    string
	Address *string
}

func (m CustomerPtr) TableName() string {
	return "customer"
}

type CustomerNull struct {
	ID      sql.NullInt64 `db:"pk,id"`
	Email   sql.NullString
	Status  *sql.NullInt64
	Name    string
	Address sql.NullString
}

func (m CustomerNull) TableName() string {
	return "customer"
}

type CustomerEmbedded struct {
	Id    int
	Email *string
	InnerCustomer
}

func (m CustomerEmbedded) TableName() string {
	return "customer"
}

type CustomerEmbedded2 struct {
	ID    int
	Email *string
	Inner InnerCustomer
}

type InnerCustomer struct {
	Status  sql.NullInt64
	Name    *string
	Address sql.NullString
}

type City struct {
	ID   int
	Name string
}

type Item struct {
	ID2  int
	Name string
}
