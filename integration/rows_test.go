package dbx_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRows_all_PointerSlice(t *testing.T) {
	db := getPreparedDB()
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

	var items []Item
	err := db.NewQuery("SELECT * FROM item ORDER BY id").All(&items)
	if assert.Nil(t, err) {
		assert.True(t, len(items) > 0, "should have items")
	}
}

func TestRows_all_PointerSlice_SameResults(t *testing.T) {
	db := getPreparedDB()
	defer func() { _ = db.Close() }()

	var ptrs []*Item
	var vals []Item
	assert.Nil(t, db.NewQuery("SELECT * FROM item ORDER BY id").All(&vals))
	assert.Nil(t, db.NewQuery("SELECT * FROM item ORDER BY id").All(&ptrs))

	if assert.Equal(t, len(vals), len(ptrs), "same number of results") {
		for i := range vals {
			assert.Equal(t, vals[i].Name, ptrs[i].Name, "Name should match")
		}
	}
}

func TestRows_all_InvalidTypes(t *testing.T) {
	db := getPreparedDB()
	defer func() { _ = db.Close() }()

	var strs []*string
	err := db.NewQuery("SELECT * FROM item").All(&strs)
	assert.NotNil(t, err, "should reject []*string")

	var ints []int
	err = db.NewQuery("SELECT * FROM item").All(&ints)
	assert.NotNil(t, err, "should reject []int")
}

func TestRows_ScanStruct_WithJoinAlias(t *testing.T) {
	db := getPreparedDB()
	defer func() { _ = db.Close() }()

	type OrderItemInfo struct {
		Quantity int
		Subtotal float64
		Name     string
	}

	var info OrderItemInfo
	q := "SELECT oi.quantity, oi.subtotal, i.name FROM order_item oi INNER JOIN item i ON oi.item_id = i.id WHERE oi.order_id = 1 AND oi.item_id = 1"
	err := db.NewQuery(q).One(&info)
	if assert.Nil(t, err) {
		assert.Equal(t, 1, info.Quantity)
		assert.Equal(t, "The Go Programming Language", info.Name)
	}
}

func TestRows_All_WithJoinAlias(t *testing.T) {
	db := getPreparedDB()
	defer func() { _ = db.Close() }()

	type OrderItemInfo struct {
		Quantity int
		Name     string
	}

	var items []OrderItemInfo
	q := "SELECT oi.quantity, i.name FROM order_item oi INNER JOIN item i ON oi.item_id = i.id WHERE oi.order_id = 1 ORDER BY oi.item_id"
	err := db.NewQuery(q).All(&items)
	if assert.Nil(t, err) {
		assert.Equal(t, 2, len(items))
		assert.Equal(t, "The Go Programming Language", items[0].Name)
		assert.Equal(t, "Go in Action", items[1].Name)
	}
}
