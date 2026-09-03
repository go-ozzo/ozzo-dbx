package dbx_test

import (
	"database/sql"
	"testing"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/stretchr/testify/assert"
)

func TestModelQuery_Insert(t *testing.T) {
	db := getPreparedDB()
	defer func() { _ = db.Close() }()

	name := "test"
	email := "test@example.com"

	{
		// inserting normally
		customer := Customer{
			Name:  name,
			Email: email,
		}
		err := db.Model(&customer).Insert()
		if assert.Nil(t, err) {
			assert.Equal(t, 4, customer.ID)
			var c Customer
			assert.Nil(t, db.Select().From("customer").Where(dbx.HashExp{"ID": 4}).One(&c))
			assert.Equal(t, name, c.Name)
			assert.Equal(t, email, c.Email)
			assert.Equal(t, 0, c.Status)
			assert.False(t, c.Address.Valid)
		}
	}

	{
		// inserting with pointer-typed fields
		customer := CustomerPtr{
			Name:  name,
			Email: &email,
		}
		err := db.Model(&customer).Insert()
		if assert.Nil(t, err) && assert.NotNil(t, customer.ID) {
			assert.Equal(t, 5, *customer.ID)
			var c CustomerPtr
			assert.Nil(t, db.Select().From("customer").Where(dbx.HashExp{"ID": 4}).One(&c))
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
		// inserting with null-typed fields
		customer := CustomerNull{
			Name:  name,
			Email: sql.NullString{String: email, Valid: true},
		}
		err := db.Model(&customer).Insert()
		if assert.Nil(t, err) {
			var c CustomerNull
			assert.Nil(t, db.Select().From("customer").Where(dbx.HashExp{"ID": 4}).One(&c))
			assert.Equal(t, name, c.Name)
			assert.Equal(t, email, c.Email.String)
			if assert.NotNil(t, c.Status) {
				assert.Equal(t, int64(0), c.Status.Int64)
			}
			assert.False(t, c.Address.Valid)
		}
	}

	{
		// inserting with embedded structures
		customer := CustomerEmbedded{
			Id:    100,
			Email: &email,
			InnerCustomer: InnerCustomer{
				Name:   &name,
				Status: sql.NullInt64{Int64: 1, Valid: true},
			},
		}
		err := db.Model(&customer).Insert()
		if assert.Nil(t, err) {
			assert.Equal(t, 100, customer.Id)
			var c CustomerEmbedded
			assert.Nil(t, db.Select().From("customer").Where(dbx.HashExp{"ID": 100}).One(&c))
			assert.Equal(t, name, *c.Name)
			assert.Equal(t, email, *c.Email)
			if assert.NotNil(t, c.Status) {
				assert.Equal(t, int64(1), c.Status.Int64)
			}
			assert.False(t, c.Address.Valid)
		}
	}

	{
		// inserting with include/exclude fields
		customer := Customer{
			Name:   name,
			Email:  email,
			Status: 1,
		}
		err := db.Model(&customer).Exclude("Name").Insert("Name", "Email")
		if assert.Nil(t, err) {
			assert.Equal(t, 101, customer.ID)
			var c Customer
			db.Select().From("customer").Where(dbx.HashExp{"ID": 101}).One(&c) //nolint:errcheck // NULL name -> string scan error expected
			assert.Equal(t, "", c.Name)
			assert.Equal(t, email, c.Email)
			assert.Equal(t, 0, c.Status)
			assert.False(t, c.Address.Valid)
		}
	}

	var a int
	assert.NotNil(t, db.Model(&a).Insert())
}

func TestModelQuery_Update(t *testing.T) {
	db := getPreparedDB()
	defer func() { _ = db.Close() }()

	id := 2
	name := "test"
	email := "test@example.com"
	{
		// updating normally
		customer := Customer{
			ID:    id,
			Name:  name,
			Email: email,
		}
		err := db.Model(&customer).Update()
		if assert.Nil(t, err) {
			var c Customer
			assert.Nil(t, db.Select().From("customer").Where(dbx.HashExp{"ID": id}).One(&c))
			assert.Equal(t, name, c.Name)
			assert.Equal(t, email, c.Email)
			assert.Equal(t, 0, c.Status)
		}
	}

	{
		// updating without primary keys
		item2 := Item{
			Name: name,
		}
		err := db.Model(&item2).Update()
		assert.Equal(t, dbx.MissingPKError, err)
	}

	{
		// updating all fields
		customer := CustomerPtr{
			ID:    &id,
			Name:  name,
			Email: &email,
		}
		err := db.Model(&customer).Update()
		if assert.Nil(t, err) {
			assert.Equal(t, id, *customer.ID)
			var c CustomerPtr
			assert.Nil(t, db.Select().From("customer").Where(dbx.HashExp{"ID": id}).One(&c))
			assert.Equal(t, name, c.Name)
			if assert.NotNil(t, c.Email) {
				assert.Equal(t, email, *c.Email)
			}
			assert.Nil(t, c.Status)
		}
	}

	{
		// updating selected fields only
		id = 3
		customer := CustomerPtr{
			ID:    &id,
			Name:  name,
			Email: &email,
		}
		err := db.Model(&customer).Update("Name", "Email")
		if assert.Nil(t, err) {
			assert.Equal(t, id, *customer.ID)
			var c CustomerPtr
			assert.Nil(t, db.Select().From("customer").Where(dbx.HashExp{"ID": id}).One(&c))
			assert.Equal(t, name, c.Name)
			if assert.NotNil(t, c.Email) {
				assert.Equal(t, email, *c.Email)
			}
			if assert.NotNil(t, c.Status) {
				assert.Equal(t, 2, *c.Status)
			}
		}
	}

	{
		// updating non-struct
		var a int
		assert.NotNil(t, db.Model(&a).Update())
	}
}

func TestModelQuery_Delete(t *testing.T) {
	db := getPreparedDB()
	defer func() { _ = db.Close() }()

	customer := Customer{
		ID: 2,
	}
	err := db.Model(&customer).Delete()
	if assert.Nil(t, err) {
		var m Customer
		err := db.Select().From("customer").Where(dbx.HashExp{"ID": 2}).One(&m)
		assert.NotNil(t, err)
	}

	{
		// deleting without primary keys
		item2 := Item{
			Name: "",
		}
		err := db.Model(&item2).Delete()
		assert.Equal(t, dbx.MissingPKError, err)
	}

	var a int
	assert.NotNil(t, db.Model(&a).Delete())
}
