package dbx

import (
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
	item := Item{
		Name: name,
	}
	err := db.Model(&item).Insert()
	if assert.Nil(t, err) {
		assert.Equal(t, 6, item.ID)
	}

	item2 := Item2{
		Name: &name,
	}
	err = db.Model(&item2).Insert()
	if assert.Nil(t, err) && assert.NotNil(t, item2.ID) {
		assert.Equal(t, 7, *item2.ID)
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
