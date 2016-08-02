package dbx

import (
	"errors"
	"reflect"
)

type (
	// TableModel is the interface that should be implemented by models which have unconventional table names.
	TableModel interface {
		TableName() string
	}

	// ModelQuery represents a query associated with a struct model.
	ModelQuery struct {
		builder   Builder
		model     *structValue
		exclude   []string
		lastError error
	}
)

var (
	MissingPKError   = errors.New("missing primary key declaration")
	CompositePKError = errors.New("composite primary key is not supported")
)

func newModelQuery(model interface{}, fieldMapFunc FieldMapFunc, builder Builder) *ModelQuery {
	q := &ModelQuery{
		builder: builder,
		model:   newStructValue(model, fieldMapFunc),
	}
	if q.model == nil {
		q.lastError = VarTypeError("must be a pointer to a struct representing the model")
	}
	return q
}

// Exclude excludes the specified struct fields from being inserted/updated into the DB table.
func (q *ModelQuery) Exclude(attrs ...string) *ModelQuery {
	q.exclude = attrs
	return q
}

// Insert inserts a row in the table using the struct model associated with this query.
//
// By default, it inserts *all* public fields into the table, including those nil or empty ones.
// You may pass a list of the fields to this method to indicate that only those fields should be inserted.
// You may also call Exclude to exclude some fields from being inserted.
//
// If a model has an empty primary key, it is considered auto-incremental and the corresponding struct
// field will be filled with the generated primary key value after a successful insertion.
func (q *ModelQuery) Insert(attrs ...string) error {
	if q.lastError != nil {
		return q.lastError
	}
	tableName := q.model.tableName
	cols := q.model.columns(attrs, q.exclude)
	pkName := ""
	for name, value := range q.model.pk() {
		if isAutoInc(value) {
			delete(cols, name)
			pkName = name
			break
		}
	}

	result, err := q.builder.Insert(tableName, Params(cols)).Execute()
	if err == nil && pkName != "" {
		pkValue, err := result.LastInsertId()
		if err != nil {
			return err
		}
		indirect(q.model.dbNameMap[pkName].getField(q.model.value)).SetInt(pkValue)
	}
	return err
}

func isAutoInc(value interface{}) bool {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Ptr:
		return v.IsNil() || isAutoInc(v.Elem())
	case reflect.Invalid:
		return true
	}
	return false
}

// Update updates a row in the table using the struct model associated with this query.
// The row being updated has the same primary key as specified by the model.
//
// By default, it updates *all* public fields in the table, including those nil or empty ones.
// You may pass a list of the fields to this method to indicate that only those fields should be updated.
// You may also call Exclude to exclude some fields from being updated.
func (q *ModelQuery) Update(attrs ...string) error {
	if q.lastError != nil {
		return q.lastError
	}
	pk := q.model.pk()
	if len(pk) == 0 {
		return MissingPKError
	}

	cols := q.model.columns(attrs, q.exclude)
	for name := range pk {
		delete(cols, name)
	}
	_, err := q.builder.Update(q.model.tableName, Params(cols), HashExp(pk)).Execute()
	return err
}

// Delete deletes a row in the table using the primary key specified by the struct model associated with this query.
func (q *ModelQuery) Delete() error {
	if q.lastError != nil {
		return q.lastError
	}
	pk := q.model.pk()
	if len(pk) == 0 {
		return MissingPKError
	}
	_, err := q.builder.Delete(q.model.tableName, HashExp(pk)).Execute()
	return err
}
