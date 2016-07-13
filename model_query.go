package dbx

import "database/sql"

type TableModel interface {
	TableName() string
}

type ModelQuery struct {
	builder Builder
	model   *structValue
}

func newModelQuery(model interface{}, fieldMapFunc FieldMapFunc, builder Builder) *ModelQuery {
	sv := newStructValue(model, fieldMapFunc)
	if sv == nil {
		// todo: log error
	}
	return &ModelQuery{
		builder: builder,
		model:   sv,
	}
}

func (q *ModelQuery) Insert(attrs ...string) (sql.Result, error) {
	tableName := q.model.tableName
	cols := q.model.fields(attrs...)
	// todo: fill in PK, remove nil PK
	return q.builder.Insert(tableName, Params(cols)).Execute()
}

func (q *ModelQuery) Update(attrs ...string) (sql.Result, error) {
	cols := q.model.fields(attrs...)
	pk := q.model.pk()
	for name := range pk {
		delete(cols, name)
	}
	return q.builder.Update(q.model.tableName, Params(cols), HashExp(pk)).Execute()
}

func (q *ModelQuery) Delete() (sql.Result, error) {
	pk := HashExp(q.model.pk())
	return q.builder.Delete(q.model.tableName, pk).Execute()
}
