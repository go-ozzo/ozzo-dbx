package dbx

import "errors"

type TableModel interface {
	TableName() string
}

type ModelQuery struct {
	builder   Builder
	model     *structValue
	exclude   []string
	lastError error
}

func newModelQuery(model interface{}, fieldMapFunc FieldMapFunc, builder Builder) *ModelQuery {
	sv := newStructValue(model, fieldMapFunc)
	q := &ModelQuery{
		builder: builder,
		model:   sv,
	}
	if sv == nil {
		q.lastError = errors.New("The model must be specified as a pointer to the model struct.")
	}
	return q
}

func (q *ModelQuery) Exclude(attrs ...string) *ModelQuery {
	q.exclude = attrs
	return q
}

func (q *ModelQuery) Insert(attrs ...string) error {
	if q.lastError != nil {
		return q.lastError
	}
	tableName := q.model.tableName
	cols := q.model.columns(attrs, q.exclude)
	pk := q.model.pk()
	ai := "" // auto-incremental column
	for pkc := range pk {
		if col, ok := cols[pkc]; ok {
			if col == nil || col == 0 {
				delete(cols, pkc)
				ai = pkc
			}
		}
	}

	result, err := q.builder.Insert(tableName, Params(cols)).Execute()
	if err == nil && ai != "" {
		pkValue, err := result.LastInsertId()
		if err != nil {
			return err
		}
		q.model.dbNameMap[ai].getField(q.model.value).SetInt(pkValue)
	}
	return err
}

func (q *ModelQuery) Update(attrs ...string) error {
	if q.lastError != nil {
		return q.lastError
	}
	cols := q.model.columns(attrs, q.exclude)
	pk := q.model.pk()
	for name := range pk {
		delete(cols, name)
	}
	_, err := q.builder.Update(q.model.tableName, Params(cols), HashExp(pk)).Execute()
	return err
}

func (q *ModelQuery) Delete() error {
	if q.lastError != nil {
		return q.lastError
	}
	pk := q.model.pk()
	_, err := q.builder.Delete(q.model.tableName, HashExp(pk)).Execute()
	return err
}
