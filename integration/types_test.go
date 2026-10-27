package dbx_test

import "database/sql"

// Test types for integration tests.

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
	Id    int //nolint:revive // "Id", not "ID", on purpose: it covers the "Id" primary-key fallback in StructInfo.build
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

type Account struct {
	ID      int
	Code    string
	Name    string
	Created string
}

type AccountPeriod struct {
	ID        int
	AccountID int
	Period    string
	Amount    int
	Created   sql.NullString
}

// CustomerPtrRecv declares TableName with a pointer receiver.
type CustomerPtrRecv struct {
	ID    int
	Email string
	Name  string
}

func (m *CustomerPtrRecv) TableName() string {
	return "customer"
}
