package dbx

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Item struct {
	ID2  int
	Name string
}

func TestModelQuery_Insert(t *testing.T) {
	db := getPreparedDB()
	defer db.Close()

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
			db.Select().From("customer").Where(HashExp{"ID": 4}).One(&c)
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
		// inserting with null-typed fields
		customer := CustomerNull{
			Name:  name,
			Email: sql.NullString{email, true},
		}
		err := db.Model(&customer).Insert()
		if assert.Nil(t, err) {
			// potential todo: need to check if the field implements sql.Scanner
			// assert.Equal(t, int64(6), customer.ID.Int64)
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
		// inserting with embedded structures
		customer := CustomerEmbedded{
			Id:    100,
			Email: &email,
			InnerCustomer: InnerCustomer{
				Name:   &name,
				Status: sql.NullInt64{1, true},
			},
		}
		err := db.Model(&customer).Insert()
		if assert.Nil(t, err) {
			assert.Equal(t, 100, customer.Id)
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
			db.Select().From("customer").Where(HashExp{"ID": 101}).One(&c)
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
	defer db.Close()

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
			db.Select().From("customer").Where(HashExp{"ID": id}).One(&c)
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
		assert.Equal(t, MissingPKError, err)
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
			db.Select().From("customer").Where(HashExp{"ID": id}).One(&c)
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
			db.Select().From("customer").Where(HashExp{"ID": id}).One(&c)
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
	defer db.Close()

	customer := Customer{
		ID: 2,
	}
	err := db.Model(&customer).Delete()
	if assert.Nil(t, err) {
		var m Customer
		err := db.Select().From("customer").Where(HashExp{"ID": 2}).One(&m)
		assert.NotNil(t, err)
	}

	{
		// deleting without primary keys
		item2 := Item{
			Name: "",
		}
		err := db.Model(&item2).Delete()
		assert.Equal(t, MissingPKError, err)
	}

	var a int
	assert.NotNil(t, db.Model(&a).Delete())
}
