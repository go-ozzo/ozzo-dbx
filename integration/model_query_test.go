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

func TestModelQuery_Upsert(t *testing.T) {
	db := getPreparedDB()
	defer func() { _ = db.Close() }()

	{
		// PK-only table, existing PK: the row is updated
		customer := Customer{ID: 2, Email: "upd@example.com", Name: "upd", Status: 7}
		if assert.Nil(t, db.Model(&customer).Upsert()) {
			var c Customer
			assert.Nil(t, db.Select().From("customer").Where(dbx.HashExp{"id": 2}).One(&c))
			assert.Equal(t, "upd@example.com", c.Email)
			assert.Equal(t, "upd", c.Name)
			assert.Equal(t, 7, c.Status)
		}
	}

	{
		// PK-only table, zero PK: a new row is inserted; the struct PK is not filled
		customer := Customer{Email: "ins@example.com", Name: "ins"}
		if assert.Nil(t, db.Model(&customer).Upsert()) {
			assert.Equal(t, 0, customer.ID)
			var c Customer
			assert.Nil(t, db.Select().From("customer").Where(dbx.HashExp{"email": "ins@example.com"}).One(&c))
			assert.Equal(t, 4, c.ID)
			assert.Equal(t, "ins", c.Name)
		}
	}

	{
		// selected fields only: the others keep their values
		customer := Customer{ID: 3, Email: "new3@example.com", Name: "only-name", Status: 9}
		if assert.Nil(t, db.Model(&customer).Upsert("Name", "Email")) {
			var c Customer
			assert.Nil(t, db.Select().From("customer").Where(dbx.HashExp{"id": 3}).One(&c))
			assert.Equal(t, "only-name", c.Name)
			assert.Equal(t, "new3@example.com", c.Email)
			assert.Equal(t, 2, c.Status)
			assert.Equal(t, "address3", c.Address.String)
		}

		// the insert part must be valid on its own, even for an existing row: leaving out the
		// NOT NULL column email (no default) fails in strict mode, and the row is unchanged
		customer = Customer{ID: 3, Name: "no-email"}
		assert.NotNil(t, db.Model(&customer).Upsert("Name"))
		var c Customer
		assert.Nil(t, db.Select().From("customer").Where(dbx.HashExp{"id": 3}).One(&c))
		assert.Equal(t, "only-name", c.Name)
	}

	{
		// a unique key besides the PK: the existing row is updated and keeps its own PK
		account := Account{ID: 99, Code: "A1", Name: "renamed", Created: "c2"}
		if assert.Nil(t, db.Model(&account).Exclude("Created").Upsert()) {
			var a []Account
			assert.Nil(t, db.Select().From("account").All(&a))
			if assert.Len(t, a, 1) {
				assert.Equal(t, 1, a[0].ID)
				assert.Equal(t, "renamed", a[0].Name)
				assert.Equal(t, "c1", a[0].Created)
			}
		}
	}

	{
		// composite unique key + Exclude: excluded fields are neither inserted nor updated
		period := AccountPeriod{AccountID: 1, Period: "2026-10", Amount: 20, Created: sql.NullString{String: "c2", Valid: true}}
		assert.Nil(t, db.Model(&period).Exclude("Created").Upsert())
		period = AccountPeriod{AccountID: 1, Period: "2026-09", Amount: 30, Created: sql.NullString{String: "c3", Valid: true}}
		assert.Nil(t, db.Model(&period).Exclude("Created").Upsert())

		var p []AccountPeriod
		assert.Nil(t, db.Select().From("account_period").OrderBy("period").All(&p))
		if assert.Len(t, p, 2) {
			assert.Equal(t, "2026-09", p[0].Period)
			assert.Equal(t, 1, p[0].ID)
			assert.Equal(t, 30, p[0].Amount)
			assert.Equal(t, "c1", p[0].Created.String)
			assert.Equal(t, "2026-10", p[1].Period)
			assert.Equal(t, 20, p[1].Amount)
			// the excluded column was not inserted either
			assert.False(t, p[1].Created.Valid)
		}
	}

	{
		// errors
		assert.NotNil(t, db.Model(&Customer{ID: 1}).Upsert("ID"))
		assert.Equal(t, dbx.MissingPKError, db.Model(&Item{Name: "x"}).Upsert())
		var a int
		assert.NotNil(t, db.Model(&a).Upsert())
	}
}
