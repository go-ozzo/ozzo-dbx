// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbx

import (
	"database/sql"
	"reflect"
	"regexp"
	"strings"
	"sync"
)

type fieldType struct {
	isPK bool
	path []int
}

type structType struct {
	fieldTypes []fieldType
	nameMap    map[string]fieldType
	dbMap      map[string]fieldType
	pk         []string
}

type structValue struct {
	fields    map[string]reflect.Value
	tableName string
}

func (s *structValue) pk() interface{} {
	return nil
}

// FieldMapFunc converts a struct field name into a DB column name.
type FieldMapFunc func(string) string

type fieldMapKey struct {
	t reflect.Type
	m reflect.Value
}

type fieldInfo struct {
	Name    string
	ColName string
	IsPK    bool
	Path    []int
}

type fieldMap map[string]fieldInfo

var (
	// DbTag is the name of the struct tag used to specify the column name for the associated struct field
	DbTag = "db"

	muFieldMap sync.Mutex
	fieldMaps  = make(map[fieldMapKey]fieldMap)
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
func getFieldMap(a reflect.Type, mapper FieldMapFunc) fieldMap {
	muFieldMap.Lock()
	defer muFieldMap.Unlock()

	key := fieldMapKey{a, reflect.ValueOf(mapper)}
	if m, ok := fieldMaps[key]; ok {
		return m
	}

	fm := fieldMap{}
	buildFieldMap(a, make([]int, 0), "", "", fm, mapper)
	fieldMaps[key] = fm

	return fm
}

var scannerType = reflect.TypeOf((*sql.Scanner)(nil)).Elem()

// buildFieldMap is called by getFieldMap recursively to build field map for a struct.
func buildFieldMap(a reflect.Type, path []int, namePrefix, colPrefix string, fm fieldMap, mapper FieldMapFunc) {
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

		colName := tag
		name := field.Name
		if colName == "" && !field.Anonymous {
			colName = field.Name
			if mapper != nil {
				colName = mapper(colName)
			}
		}
		if field.Anonymous {
			name = ""
		}

		if ft.Kind() == reflect.Struct && !reflect.PtrTo(ft).Implements(scannerType) {
			// dive into non-scanner struct
			buildFieldMap(ft, path2, concat(namePrefix, name), concat(colPrefix, colName), fm, mapper)
		} else if colName != "" {
			// non-anonymous scanner or struct field
			colName = concat(colPrefix, colName)
			fm[colName] = fieldInfo{
				Name:    concat(namePrefix, name),
				ColName: colName,
				IsPK:    false,
				Path:    path2,
			}
		}
	}
}

func concat(s1, s2 string) string {
	if s1 == "" {
		return s2
	} else if s2 == "" {
		return s1
	} else {
		return s1 + "." + s2
	}
}

// getStructField returns the reflection value of the field specified by its field map path.
func (fi fieldInfo) getStructField(a reflect.Value) reflect.Value {
	for _, i := range fi.Path {
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
