package dbx

import (
	"reflect"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
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

// resolveColumn simulates the column resolution logic used in ScanStruct and all.
// It tries exact match first, then strips the table alias prefix (first dot) as fallback.
func resolveColumn(dbNameMap map[string]*FieldInfo, col string) (*FieldInfo, bool) {
	if fi, ok := dbNameMap[col]; ok {
		return fi, true
	}
	if dotIdx := strings.Index(col, "."); dotIdx >= 0 {
		if fi, ok := dbNameMap[col[dotIdx+1:]]; ok {
			return fi, true
		}
	}
	return nil, false
}

func TestResolveColumn_StripTablePrefix(t *testing.T) {
	type Address struct {
		City  string `db:"city"`
		State string `db:"state"`
	}
	type Person struct {
		Name    string  `db:"name"`
		Email   string  `db:"email"`
		Address Address `db:"address"`
	}

	si := getStructInfo(reflect.TypeOf(Person{}), nil)

	tests := []struct {
		name      string
		col       string
		wantMatch bool
		wantField string
	}{
		{
			name:      "exact match without prefix",
			col:       "name",
			wantMatch: true,
			wantField: "name",
		},
		{
			name:      "table alias prefix stripped",
			col:       "t.name",
			wantMatch: true,
			wantField: "name",
		},
		{
			name:      "nested struct exact match",
			col:       "address.city",
			wantMatch: true,
			wantField: "address.city",
		},
		{
			name:      "table alias prefix with nested struct",
			col:       "src.address.city",
			wantMatch: true,
			wantField: "address.city",
		},
		{
			name:      "unknown column no prefix",
			col:       "nonexistent",
			wantMatch: false,
		},
		{
			name:      "unknown column with prefix",
			col:       "t.nonexistent",
			wantMatch: false,
		},
		{
			name:      "schema.table.column strips first prefix",
			col:       "myschema.name",
			wantMatch: true,
			wantField: "name",
		},
		{
			name:      "email exact match",
			col:       "email",
			wantMatch: true,
			wantField: "email",
		},
		{
			name:      "email with table prefix",
			col:       "p.email",
			wantMatch: true,
			wantField: "email",
		},
		{
			name:      "nested state with table prefix",
			col:       "p.address.state",
			wantMatch: true,
			wantField: "address.state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fi, ok := resolveColumn(si.dbNameMap, tt.col)
			assert.Equal(t, tt.wantMatch, ok, "match result for %q", tt.col)
			if tt.wantMatch && ok {
				assert.Equal(t, tt.wantField, fi.dbName, "resolved field for %q", tt.col)
			}
		})
	}
}

func TestResolveColumn_CustomFieldMapper(t *testing.T) {
	type Product struct {
		ID        int `db:"pk,id"`
		LongName  string
		ShortDesc string
	}

	si := getStructInfo(reflect.TypeOf(Product{}), DefaultFieldMapFunc)

	// DefaultFieldMapFunc maps "LongName" → "long_name", "ShortDesc" → "short_desc"
	tests := []struct {
		col       string
		wantMatch bool
		wantField string
	}{
		{"long_name", true, "long_name"},
		{"t.long_name", true, "long_name"},
		{"short_desc", true, "short_desc"},
		{"src.short_desc", true, "short_desc"},
		{"id", true, "id"},
		{"p.id", true, "id"},
	}

	for _, tt := range tests {
		t.Run(tt.col, func(t *testing.T) {
			fi, ok := resolveColumn(si.dbNameMap, tt.col)
			assert.Equal(t, tt.wantMatch, ok)
			if tt.wantMatch && ok {
				assert.Equal(t, tt.wantField, fi.dbName)
			}
		})
	}
}

// Integration tests using MySQL (run in CI with getPreparedDB).
// These verify the full path: SQL query with JOIN alias → ScanStruct/All → correct field population.

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
