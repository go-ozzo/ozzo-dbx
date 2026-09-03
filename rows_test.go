package dbx

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

	// DefaultFieldMapFunc maps "LongName" -> "long_name", "ShortDesc" -> "short_desc"
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
