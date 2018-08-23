// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbx

import (
	"database/sql"
	"fmt"
	"reflect"
)

// VarTypeError indicates a variable type error when trying to populating a variable with DB result.
type VarTypeError string

// Error returns the error message.
func (s VarTypeError) Error() string {
	return "Invalid variable type: " + string(s)
}

// NullStringMap is a map of sql.NullString that can be used to hold DB query result.
// The map keys correspond to the DB column names, while the map values are their corresponding column values.
type NullStringMap map[string]sql.NullString

// Rows enhances sql.Rows by providing additional data query methods.
// Rows can be obtained by calling Query.Rows(). It is mainly used to populate data row by row.
type Rows struct {
	*sql.Rows
	fieldMapFunc FieldMapFunc
}

// ScanMap populates the current row of data into a NullStringMap.
// Note that the NullStringMap must not be nil, or it will panic.
// The NullStringMap will be populated using column names as keys and their values as
// the corresponding element values.
func (r *Rows) ScanMap(a NullStringMap) error {
	cols, _ := r.Columns()
	var refs []interface{}
	for i := 0; i < len(cols); i++ {
		var t sql.NullString
		refs = append(refs, &t)
	}
	if err := r.Scan(refs...); err != nil {
		return err
	}

	for i, col := range cols {
		a[col] = *refs[i].(*sql.NullString)
	}

	return nil
}

// ScanStruct populates the current row of data into a struct.
// The struct must be given as a pointer.
//
// ScanStruct associates struct fields with DB table columns through a field mapping function.
// It populates a struct field with the data of its associated column.
// Note that only exported struct fields will be populated.
//
// By default, DefaultFieldMapFunc() is used to map struct fields to table columns.
// This function separates each word in a field name with a underscore and turns every letter into lower case.
// For example, "LastName" is mapped to "last_name", "MyID" is mapped to "my_id", and so on.
// To change the default behavior, set DB.FieldMapper with your custom mapping function.
// You may also set Query.FieldMapper to change the behavior for particular queries.
func (r *Rows) ScanStruct(a interface{}) error {
	return r.scanStructV(reflect.ValueOf(a))
}
func (r *Rows) scanStructV(rv reflect.Value) error {
	rv = indirect0(rv, nil, true, true)
	if rv.Kind() != reflect.Struct {
		return VarTypeError("must be a pointer to a struct")
	}
	if !rv.CanSet() {
		return VarTypeError("not settable value")
	}

	si := getStructInfo(rv.Type(), r.fieldMapFunc)

	cols, _ := r.Columns()
	refs := make([]interface{}, len(cols))

	for i, col := range cols {
		if fi, ok := si.dbNameMap[col]; ok {
			refs[i] = fi.getField(rv).Addr().Interface()
		} else {
			refs[i] = &sql.NullString{}
		}
	}

	return r.Scan(refs...)
}

// all populates all rows of query result into a slice of struct or NullStringMap.
// Note that the slice must be given as a pointer.
func (r *Rows) all(slice interface{}) error {
	defer r.Close()

	v := indirect(reflect.ValueOf(slice))

	if v.Kind() != reflect.Slice {
		return VarTypeError("not a slice, must be a slice of struct or NullStringMap")
	}

	if v.IsNil() {
		// create an empty slice
		v.Set(reflect.MakeSlice(v.Type(), 0, 0))
	}

	if !v.CanSet() {
		return VarTypeError("slice not settable")
	}

	// check for a valid element type
	et := v.Type().Elem()
	var si *structInfo
	finV := indirect0(reflect.New(et), nil, false, false)
	switch finV.Kind() {
	default:
		return VarTypeError(fmt.Sprintf("a slice of %s, must be a slice of struct or NullStringMap", finV.Kind()))
	case reflect.Map:
	case reflect.Struct:
		si = getStructInfo(finV.Type(), r.fieldMapFunc)
	}
	cols, err := r.Columns()
	if err != nil {
		return err
	}

	// everything prepared, now scan the result
	for r.Next() {
		ev, err := r.scanRow(et, si, cols)
		if err != nil {
			return err
		}
		newSliceV := reflect.Append(v, ev)
		v.Set(newSliceV)
	}
	return nil
}
func (r *Rows) scanRow(et reflect.Type, si *structInfo, cols []string) (ev reflect.Value, err error) {
	ev = reflect.New(et).Elem()
	evi := indirect(ev)
	if evi.Kind() == reflect.Map {
		if evi.IsNil() {
			evi.Set(reflect.MakeMap(evi.Type()))
		}
		err = r.ScanMap(evi.Interface().(NullStringMap))
	} else {
		refs := make([]interface{}, len(cols))
		for i, col := range cols {
			if fi, ok := si.dbNameMap[col]; ok {
				refs[i] = fi.getField(evi).Addr().Interface()
			} else {
				refs[i] = &sql.NullString{}
			}
		}
		err = r.Scan(refs...)
	}
	return
}

// column populates the given slice with the first column of the query result.
// Note that the slice must be given as a pointer.
func (r *Rows) column(slice interface{}) error {
	defer r.Close()

	v := indirect(reflect.ValueOf(slice))
	if !v.CanSet() {
		return VarTypeError("not settable value")
	}

	if v.Kind() != reflect.Slice {
		return VarTypeError("must be a (pointer to a) slice")
	}

	et := v.Type().Elem()

	cols, _ := r.Columns()
	for r.Next() {
		ev := reflect.New(et)
		refs := make([]interface{}, len(cols))
		for i := range cols {
			if i == 0 {
				refs[i] = ev.Interface()
			} else {
				refs[i] = &sql.NullString{}
			}
		}
		if err := r.Scan(refs...); err != nil {
			return err
		}
		v.Set(reflect.Append(v, ev.Elem()))
	}

	return r.Close()
}

// one populates a single row of query result into a struct or a NullStringMap.
// Note that if a struct is given, it should be a pointer.
func (r *Rows) one(a interface{}) error {
	defer r.Close()

	if !r.Next() {
		if err := r.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}

	rv := indirect(reflect.ValueOf(a))

	if rv.Kind() == reflect.Map {
		if rv.IsNil() {
			if !rv.CanSet() {
				return VarTypeError("not settable value")
			}
			rv.Set(reflect.MakeMap(rv.Type()))
		}
		v, ok := rv.Interface().(NullStringMap)
		if !ok {
			return VarTypeError("must be a NullStringMap")
		}
		return r.ScanMap(v)
	} else {
		return r.scanStructV(rv)
	}
}

// row populates a single row of query result into a list of variables.
func (r *Rows) row(a ...interface{}) error {
	defer r.Close()

	for _, dp := range a {
		if _, ok := dp.(*sql.RawBytes); ok {
			return VarTypeError("RawBytes isn't allowed on Row()")
		}
	}

	if !r.Next() {
		if err := r.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}
	if err := r.Scan(a...); err != nil {
		return err
	}

	return r.Close()
}
