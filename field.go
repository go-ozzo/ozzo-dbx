// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbx

import (
	"reflect"
	"regexp"
	"strings"
	"sync"
)

// FieldMapFunc converts a struct field name into a DB column name.
type FieldMapFunc func(string) string

type fieldMapKey struct {
	t reflect.Type
	m reflect.Value
}

var (
	// DbTag is the name of the struct tag used to specify the column name for the associated struct field
	DbTag = "db"

	muFieldMap sync.Mutex
	fieldMap   = make(map[fieldMapKey]map[string][]int)
	fieldRegex = regexp.MustCompile(`([^A-Z_])([A-Z])`)
)

// DefaultFieldMapFunc maps a field name to a DB column name.
// The mapping rule set by this method is that words in a field name will be separated by underscores
// and the name will be turned into lower case. For example, "FirstName" maps to "first_name", and "MyID" becomes "my_id".
// See DB.FieldMapper for more details.
func DefaultFieldMapFunc(f string) string {
	return strings.ToLower(fieldRegex.ReplaceAllString(f, "${1}_$2"))
}

// getFieldMap builds a field map for a struct.
// The map returned will have field names as keys and field positions as values.
// Only exported fields are considered. For anonymous fields that are structs,
// their exported fields will be included in the map recursively.
// See TestGetFieldMap() for an example.
func getFieldMap(a reflect.Type, mapper FieldMapFunc) map[string][]int {
	muFieldMap.Lock()
	defer muFieldMap.Unlock()

	key := fieldMapKey{a, reflect.ValueOf(mapper)}
	if m, ok := fieldMap[key]; ok {
		return m
	}

	fields := make(map[string][]int)
	buildFieldMap(a, make([]int, 0), "", fields, mapper)
	fieldMap[key] = fields

	return fields
}

// buildFieldMap is called by getFieldMap recursively to build field map for a struct.
func buildFieldMap(a reflect.Type, path []int, prefix string, fields map[string][]int, mapper FieldMapFunc) {
	n := a.NumField()
	for i := 0; i < n; i++ {
		field := a.Field(i)
		tag := field.Tag.Get(DbTag)

		// only handle anonymous or exported fields
		if !field.Anonymous && field.PkgPath != "" || tag == "-" {
			continue
		}

		path2 := make([]int, len(path), len(path)+1)
		copy(path2, path)
		path2 = append(path2, i)

		ft := field.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}

		name := tag
		if name == "" && !field.Anonymous {
			name = field.Name
			if mapper != nil {
				name = mapper(name)
			}
		}
		if ft.Kind() != reflect.Struct {
			if name != "" {
				if prefix != "" {
					fields[prefix+"."+name] = path2
				} else {
					fields[name] = path2
				}
			}
			continue
		}

		if name == "" {
			buildFieldMap(ft, path2, prefix, fields, mapper)
		} else {
			p := name
			if prefix != "" {
				p = prefix + "." + p
			}
			buildFieldMap(ft, path2, p, fields, mapper)
		}
	}
}

// getStructField returns the reflection value of the field specified by its field map path.
func getStructField(a reflect.Value, path []int) reflect.Value {
	for _, i := range path {
		a = indirect(a.Field(i))
	}
	return a
}

// indirect dereferences pointers and returns the actual value it points to.
// If a pointer is nil, it will be initialized with a new value.
func indirect(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}
