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

type (
	// FieldMapFunc converts a struct field name into a DB column name.
	FieldMapFunc func(string) string

	structInfo struct {
		nameMap   map[string]*fieldInfo // mapping from struct field names to field infos
		dbNameMap map[string]*fieldInfo // mapping from db column names to field infos
		pkNames   []string              // struct field names representing PKs
	}

	structValue struct {
		*structInfo
		value     reflect.Value // the struct value
		tableName string        // the db table name for the struct
	}

	fieldInfo struct {
		name   string // field name
		dbName string // db column name
		path   []int  // index path to the struct field reflection
	}

	structInfoMapKey struct {
		t reflect.Type
		m reflect.Value
	}
)

var (
	// DbTag is the name of the struct tag used to specify the column name for the associated struct field
	DbTag = "db"

	fieldRegex      = regexp.MustCompile(`([^A-Z_])([A-Z])`)
	scannerType     = reflect.TypeOf((*sql.Scanner)(nil)).Elem()
	structInfoMap   = make(map[structInfoMapKey]*structInfo)
	muStructInfoMap sync.Mutex
)

// DefaultFieldMapFunc maps a field name to a DB column name.
// The mapping rule set by this method is that words in a field name will be separated by underscores
// and the name will be turned into lower case. For example, "FirstName" maps to "first_name", and "MyID" becomes "my_id".
// See DB.FieldMapper for more details.
func DefaultFieldMapFunc(f string) string {
	return strings.ToLower(fieldRegex.ReplaceAllString(f, "${1}_$2"))
}

func getStructInfo(a reflect.Type, mapper FieldMapFunc) *structInfo {
	muStructInfoMap.Lock()
	defer muStructInfoMap.Unlock()

	key := structInfoMapKey{a, reflect.ValueOf(mapper)}
	if si, ok := structInfoMap[key]; ok {
		return si
	}

	si := &structInfo{
		nameMap:   map[string]*fieldInfo{},
		dbNameMap: map[string]*fieldInfo{},
	}
	si.build(a, make([]int, 0), "", "", mapper)
	structInfoMap[key] = si

	return si
}

func newStructValue(model interface{}, mapper FieldMapFunc) *structValue {
	value := reflect.ValueOf(model)
	if value.Kind() != reflect.Ptr || value.Elem().Kind() != reflect.Struct || value.IsNil() {
		return nil
	}
	t := reflect.TypeOf(model).Elem()
	var tableName string
	if tm, ok := model.(TableModel); ok {
		tableName = tm.TableName()
	} else {
		tableName = DefaultFieldMapFunc(t.Name())
	}

	si := getStructInfo(t, mapper)
	return &structValue{
		structInfo: si,
		value:      value.Elem(),
		tableName:  tableName,
	}
}

func (s *structValue) pk() map[string]interface{} {
	return s.columns(s.pkNames, nil)
}

func (s *structValue) columns(include, exclude []string) map[string]interface{} {
	v := map[string]interface{}{}
	if len(include) == 0 {
		for _, fi := range s.nameMap {
			v[fi.dbName] = fi.getValue(s.value)
		}
	} else {
		for _, attr := range include {
			if fi, ok := s.nameMap[attr]; ok {
				v[fi.dbName] = fi.getValue(s.value)
			}
		}
	}
	if len(exclude) > 0 {
		for _, name := range exclude {
			if fi, ok := s.nameMap[name]; ok {
				delete(v, fi.dbName)
			}
		}
	}
	return v
}

// getValue returns the field value for the given struct value.
func (fi *fieldInfo) getValue(a reflect.Value) interface{} {
	for _, i := range fi.path {
		a = a.Field(i)
		if a.Kind() == reflect.Ptr {
			if a.IsNil() {
				return nil
			}
			a = a.Elem()
		}
	}
	return a.Interface()
}

// getField returns the reflection value of the field for the given struct value.
func (fi *fieldInfo) getField(a reflect.Value) reflect.Value {
	i := 0
	for ; i < len(fi.path)-1; i++ {
		a = indirect(a.Field(fi.path[i]))
	}
	return a.Field(fi.path[i])
}

func (si *structInfo) build(a reflect.Type, path []int, namePrefix, dbNamePrefix string, mapper FieldMapFunc) {
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

		name := field.Name
		dbName, isPK := parseTag(tag)
		if dbName == "" && !field.Anonymous {
			if mapper != nil {
				dbName = mapper(field.Name)
			} else {
				dbName = field.Name
			}
		}
		if field.Anonymous {
			name = ""
		}

		if ft.Kind() == reflect.Struct && !reflect.PtrTo(ft).Implements(scannerType) {
			// dive into non-scanner struct
			si.build(ft, path2, concat(namePrefix, name), concat(dbNamePrefix, dbName), mapper)
		} else if dbName != "" {
			// non-anonymous scanner or struct field
			fi := &fieldInfo{
				name:   concat(namePrefix, name),
				dbName: concat(dbNamePrefix, dbName),
				path:   path2,
			}
			si.nameMap[fi.name] = fi
			si.dbNameMap[fi.dbName] = fi
			if isPK {
				si.pkNames = append(si.pkNames, fi.name)
			}
		}
	}
	if len(si.pkNames) == 0 {
		if _, ok := si.nameMap["ID"]; ok {
			si.pkNames = append(si.pkNames, "ID")
		}
	}
}

func parseTag(tag string) (string, bool) {
	if tag == "pk" {
		return "", true
	}
	if strings.HasPrefix(tag, "pk,") {
		return tag[3:], true
	}
	return tag, false
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
