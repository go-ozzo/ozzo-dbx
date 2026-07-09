package dbx

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

func TestRows_all_PointerSlice(t *testing.T) {
	db := getPreparedDB()
	defer db.Close()

	var items []*Item
	err := db.NewQuery("SELECT * FROM item ORDER BY id").All(&items)
	if assert.Nil(t, err) {
		assert.True(t, len(items) > 0, "should have items")
		for _, item := range items {
			assert.NotNil(t, item, "each item should be non-nil pointer")
			assert.NotEmpty(t, item.Name, "each item should have a name")
		}
	}
}

func TestRows_all_ValueSlice(t *testing.T) {
	db := getPreparedDB()
	defer db.Close()

	var items []Item
	err := db.NewQuery("SELECT * FROM item ORDER BY id").All(&items)
	if assert.Nil(t, err) {
		assert.True(t, len(items) > 0, "should have items")
	}
}

func TestRows_all_PointerSlice_SameResults(t *testing.T) {
	db := getPreparedDB()
	defer db.Close()

	var ptrs []*Item
	var vals []Item
	db.NewQuery("SELECT * FROM item ORDER BY id").All(&vals)
	db.NewQuery("SELECT * FROM item ORDER BY id").All(&ptrs)

	if assert.Equal(t, len(vals), len(ptrs), "same number of results") {
		for i := range vals {
			assert.Equal(t, vals[i].Name, ptrs[i].Name, "Name should match")
		}
	}
}

func TestRows_all_InvalidTypes(t *testing.T) {
	db := getPreparedDB()
	defer db.Close()

	var strs []*string
	err := db.NewQuery("SELECT * FROM item").All(&strs)
	assert.NotNil(t, err, "should reject []*string")

	var ints []int
	err = db.NewQuery("SELECT * FROM item").All(&ints)
	assert.NotNil(t, err, "should reject []int")
}
