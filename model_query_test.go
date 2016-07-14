package dbx

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Item struct {
	ID   int
	Name string
}

type Item2 struct {
	ID   *int `db:"pk"`
	Name *string
}

func (m *Item2) TableName() string {
	return "item"
}

func TestModelQuery_Insert(t *testing.T) {
	db := getPreparedDB()
	defer db.Close()

	name := "test"
	email := "test@example.com"

	{
		customer := Customer{
			Name:  name,
			Email: email,
		}
		err := db.Model(&customer).Insert()
		if assert.Nil(t, err) {
			assert.Equal(t, 4, customer.ID)
			var c Customer
			db.Select().From("customer").Where(HashExp{"ID": 4}).One(&c)
			assert.Equal(t, name, c.Name)
			assert.Equal(t, email, c.Email)
			assert.Equal(t, 0, c.Status)
			assert.False(t, c.Address.Valid)
		}
	}

	{
		customer := CustomerPtr{
			Name:  name,
			Email: &email,
		}
		err := db.Model(&customer).Insert()
		if assert.Nil(t, err) && assert.NotNil(t, customer.ID) {
			assert.Equal(t, 5, *customer.ID)
			var c CustomerPtr
			db.Select().From("customer").Where(HashExp{"ID": 4}).One(&c)
			assert.Equal(t, name, c.Name)
			if assert.NotNil(t, c.Email) {
				assert.Equal(t, email, *c.Email)
			}
			if assert.NotNil(t, c.Status) {
				assert.Equal(t, 0, *c.Status)
			}
			assert.Nil(t, c.Address)
		}
	}

	{
		customer := CustomerNull{
			Name:  name,
			Email: sql.NullString{email, true},
		}
		err := db.Model(&customer).Insert()
		if assert.Nil(t, err) {
			// potential todo:
			// assert.Equal(t, int64(6), customerNull.ID.Int64)
			var c CustomerNull
			db.Select().From("customer").Where(HashExp{"ID": 4}).One(&c)
			assert.Equal(t, name, c.Name)
			assert.Equal(t, email, c.Email.String)
			if assert.NotNil(t, c.Status) {
				assert.Equal(t, int64(0), c.Status.Int64)
			}
			assert.False(t, c.Address.Valid)
		}
	}

	{
		customer := CustomerEmbedded{
			ID:    100,
			Email: &email,
			InnerCustomer: InnerCustomer{
				Name:   &name,
				Status: sql.NullInt64{1, true},
			},
		}
		err := db.Model(&customer).Insert()
		if assert.Nil(t, err) {
			assert.Equal(t, 100, customer.ID)
			var c CustomerEmbedded
			db.Select().From("customer").Where(HashExp{"ID": 100}).One(&c)
			assert.Equal(t, name, *c.Name)
			assert.Equal(t, email, *c.Email)
			if assert.NotNil(t, c.Status) {
				assert.Equal(t, int64(1), c.Status.Int64)
			}
			assert.False(t, c.Address.Valid)
		}
	}
}

func TestModelQuery_Update(t *testing.T) {
	db := getPreparedDB()
	defer db.Close()

	item := Item{
		ID:   2,
		Name: "test",
	}
	err := db.Model(&item).Update()
	if assert.Nil(t, err) {
		var m Item
		db.Select().From("item").Where(HashExp{"ID": 2}).One(&m)
		assert.Equal(t, "test", m.Name)
	}

	id := 3
	name := "test2"
	item2 := Item2{
		ID:   &id,
		Name: &name,
	}
	err = db.Model(&item2).Update()
	if assert.Nil(t, err) {
		var m Item2
		db.Select().From("item").Where(HashExp{"ID": 3}).One(&m)
		if assert.NotNil(t, m.Name) {
			assert.Equal(t, "test2", *m.Name)
		}
	}
}

func TestModelQuery_Delete(t *testing.T) {
	db := getPreparedDB()
	defer db.Close()

	item := Item{
		ID: 2,
	}
	err := db.Model(&item).Delete()
	if assert.Nil(t, err) {
		var m Item
		err := db.Select().From("item").Where(HashExp{"ID": 2}).One(&m)
		assert.NotNil(t, err)
	}
}
