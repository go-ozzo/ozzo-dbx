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
	value := indirect0(reflect.ValueOf(model), nil, false, true)
	if value.Kind() != reflect.Struct {
		return nil
	}

	return &structValue{
		structInfo: getStructInfo(value.Type(), mapper),
		value:      value,
		tableName:  GetTableName(model),
	}
}

// pk returns the primary key values indexed by the corresponding primary key column names.
func (s *structValue) pk() map[string]interface{} {
	if len(s.pkNames) == 0 {
		return nil
	}
	return s.columns(s.pkNames, nil)
}

// columns returns the struct field values indexed by their corresponding DB column names.
func (s *structValue) columns(include, exclude []string) map[string]interface{} {
	v := make(map[string]interface{}, len(s.nameMap))
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

		if isNestedStruct(ft) {
			// dive into non-scanner struct
			si.build(ft, path2, concat(namePrefix, name), concat(dbNamePrefix, dbName), mapper)
		} else if dbName != "" {
			// non-anonymous scanner or struct field
			fi := &fieldInfo{
				name:   concat(namePrefix, name),
				dbName: concat(dbNamePrefix, dbName),
				path:   path2,
			}
			// a field in an anonymous struct may be shadowed
			if _, ok := si.nameMap[fi.name]; !ok || len(path2) < len(si.nameMap[fi.name].path) {
				si.nameMap[fi.name] = fi
				si.dbNameMap[fi.dbName] = fi
				if isPK {
					si.pkNames = append(si.pkNames, fi.name)
				}
			}
		}
	}
	if len(si.pkNames) == 0 {
		if _, ok := si.nameMap["ID"]; ok {
			si.pkNames = append(si.pkNames, "ID")
		} else if _, ok := si.nameMap["Id"]; ok {
			si.pkNames = append(si.pkNames, "Id")
		}
	}
}

func isNestedStruct(t reflect.Type) bool {
	if t.PkgPath() == "time" && t.Name() == "Time" {
		return false
	}
	return t.Kind() == reflect.Struct && !reflect.PtrTo(t).Implements(scannerType)
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
	return indirect0(v, nil, false, true)
}

// GetTableName returns the table name corresponding to the given model struct or slice of structs.
// Do not call this method in the model's TableName() method, or it will cause infinite loop.
func GetTableName(a interface{}) string {
	if tm, ok := a.(TableModel); ok {
		v := reflect.ValueOf(a)
		if v.Kind() == reflect.Ptr && v.IsNil() {
			a = reflect.New(v.Type().Elem()).Interface()
			return a.(TableModel).TableName()
		}
		return tm.TableName()
	}

	tmt := reflect.TypeOf((*TableModel)(nil)).Elem()
	v := indirect0(reflect.ValueOf(a), func(v reflect.Value) bool {
		// stop once we found something that implements TableModel
		return v.Type().Implements(tmt)

	}, false, false)
	// may well be our early exit got us here
	t := v.Type()
	for t.Kind() == reflect.Ptr && !t.Implements(tmt) {
		t = t.Elem()
	}
	if t.Implements(tmt) {
		return GetTableName(reflect.New(t).Elem().Interface())
	}
	// a slice may have elements of the type we're searching for
	if t.Kind() == reflect.Slice {
		return GetTableName(reflect.New(t.Elem()).Elem().Interface())
	}
	// or a struct where we derive the name from
	if t.Kind() != reflect.Struct {
		// otherwise we can't do anything with that value
		return ""
	}
	return DefaultFieldMapFunc(t.Name())
}

// resolves the given value through pointers and interfaces, returning the actual value found.
// When init == true it initializes empty pointers on the way down. Once the match function returns
// true when called with the current value and type handled, the process is stopped early.
func indirect0(v reflect.Value, match func(v reflect.Value) bool, lastMatching, init bool) reflect.Value {
	// need to use Value.Elem() to get through interfaces
	found := v
	aborted := false
FOR:
	for v.IsValid() {
		// continue?
		if match != nil && match(v) {
			found = v
			aborted = true
			if !lastMatching {
				break FOR
			}
		}
		// inspect type
		k := v.Kind()
		switch {
		case k == reflect.Ptr:
			if v.IsNil() {
				if init && v.CanSet() {
					v.Set(reflect.New(v.Type().Elem()))
				} else {
					v = reflect.New(v.Type().Elem())
				}
			}
			v = v.Elem()
		case k == reflect.Interface:
			if v.IsNil() {
				break FOR
			}
			v = v.Elem()
		case k == reflect.Slice:
			if v.IsNil() {
				if init && v.CanSet() {
					v.Set(reflect.MakeSlice(reflect.SliceOf(v.Type().Elem()), 0, 0))
				} else {
					v = reflect.MakeSlice(reflect.SliceOf(v.Type().Elem()), 0, 0)
				}
			}
			break FOR
		default:
			break FOR
		}
	}

	// done
	if aborted {
		return found
	} else {
		return v
	}
}
