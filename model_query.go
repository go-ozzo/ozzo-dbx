package dbx

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
)

type (
	// TableModel is the interface that should be implemented by models which have unconventional table names.
	TableModel interface {
		TableName() string
	}

	// ModelQuery represents a query associated with a struct model.
	ModelQuery struct {
		db        *DB
		ctx       context.Context
		builder   Builder
		model     *structValue
		exclude   []string
		lastError error
	}
)

var (
	MissingPKError   = errors.New("missing primary key declaration")        //nolint:staticcheck // public API, renaming would break backward compatibility
	CompositePKError = errors.New("composite primary key is not supported") //nolint:staticcheck // public API, renaming would break backward compatibility
)

func NewModelQuery(model interface{}, fieldMapFunc FieldMapFunc, db *DB, builder Builder) *ModelQuery {
	q := &ModelQuery{
		db:      db,
		ctx:     db.ctx,
		builder: builder,
		model:   newStructValue(model, fieldMapFunc, db.TableMapper),
	}
	if q.model == nil {
		q.lastError = VarTypeError("must be a pointer to a struct representing the model")
	}
	return q
}

// Context returns the context associated with the query.
func (q *ModelQuery) Context() context.Context {
	return q.ctx
}

// WithContext associates a context with the query.
func (q *ModelQuery) WithContext(ctx context.Context) *ModelQuery {
	q.ctx = ctx
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
	cols := q.model.columns(attrs, q.exclude)
	pkName := ""
	for name, value := range q.model.pk() {
		if isAutoInc(value) {
			delete(cols, name)
			pkName = name
			break
		}
	}

	if pkName == "" {
		_, err := q.builder.Insert(q.model.tableName, Params(cols)).WithContext(q.ctx).Execute()
		return err
	}

	// handle auto-incremental PK
	query := q.builder.Insert(q.model.tableName, Params(cols)).WithContext(q.ctx)
	pkValue, err := insertAndReturnPK(q.db, query, pkName)
	if err != nil {
		return err
	}

	pkField := indirect(q.model.dbNameMap[pkName].getField(q.model.value))
	switch pkField.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		pkField.SetUint(uint64(pkValue))
	default:
		pkField.SetInt(pkValue)
	}

	return nil
}

func insertAndReturnPK(db *DB, query *Query, pkName string) (int64, error) {
	if db.DriverName() != "postgres" && db.DriverName() != "pgx" {
		result, err := query.Execute()
		if err != nil {
			return 0, err
		}
		return result.LastInsertId()
	}

	// specially handle postgres (lib/pq) as it doesn't support LastInsertId
	returning := fmt.Sprintf(" RETURNING %s", db.QuoteColumnName(pkName))
	query.sql += returning
	query.rawSQL += returning
	var pkValue int64
	err := query.Row(&pkValue)
	return pkValue, err
}

func isAutoInc(value interface{}) bool {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Ptr: //nolint:govet // reflect.Ptr is valid, govet toolchain mismatch
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
	_, err := q.builder.Update(q.model.tableName, Params(cols), HashExp(pk)).WithContext(q.ctx).Execute()
	return err
}

// Upsert inserts a row in the table using the struct model associated with this query, or updates
// the existing row if the insertion conflicts with it. The primary key is the conflict target.
//
// By default, it inserts and updates *all* public fields, including those nil or empty ones.
// You may pass a list of the fields to this method to indicate that only those fields should be
// inserted and updated. You may also call Exclude to exclude some fields from both the insertion
// and the update (e.g. a creation timestamp that must not change once the row exists).
// The primary key is never updated, and attrs and Exclude do not apply to it.
//
// If a model has a single-column primary key with an empty value, it is considered auto-incremental
// and is left out of the insertion. Unlike Insert, Upsert does not fill the primary key field after
// the statement runs, whether a row was inserted or updated.
//
// Upsert relies on Builder.Upsert, so it works on MySQL, PostgreSQL and SQLite and returns an error
// on other databases. On MySQL the update also happens on a conflict with any other unique index of
// the table: the matched row keeps its primary key, which then differs from the model's.
// The update copies the proposed row (see Builder.Upsert), so BEFORE INSERT triggers apply to it.
// Upsert returns an error if no field is left to update besides the primary key.
//
// The insertion part must be valid on its own even when the row already exists: the database
// checks NOT NULL constraints before it detects the conflict. So when you pass a list of fields,
// it must include every NOT NULL column that has no default value.
func (q *ModelQuery) Upsert(attrs ...string) error {
	if q.lastError != nil {
		return q.lastError
	}
	pk := q.model.pk()
	if len(pk) == 0 {
		return MissingPKError
	}

	cols := q.model.columns(attrs, q.exclude)
	constraints := make([]string, 0, len(pk))
	for name, value := range pk {
		if len(pk) == 1 && isAutoInc(value) {
			delete(cols, name)
		} else {
			cols[name] = value
		}
		constraints = append(constraints, name)
	}
	sort.Strings(constraints)

	_, err := q.builder.Upsert(q.model.tableName, Params(cols), constraints...).WithContext(q.ctx).Execute()
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
	_, err := q.builder.Delete(q.model.tableName, HashExp(pk)).WithContext(q.ctx).Execute()
	return err
}
