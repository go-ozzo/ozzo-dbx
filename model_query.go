package dbx

import "reflect"

type TableModel interface {
	TableName() string
}

type ModelQuery struct {
	tableName    string
	model        interface{}
	fieldMapFunc FieldMapFunc
}

func NewModelQuery(model interface{}, fmf FieldMapFunc) *ModelQuery {
	mq := &ModelQuery{
		model:        model,
		fieldMapFunc: fmf,
	}
	tm, ok := model.(TableModel)
	if ok {
		mq.tableName = tm.TableName()
	} else {
		t := reflect.TypeOf(model)
		if t.Kind() == reflect.Interface || t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		// todo
		mq.tableName = t.Name()
	}
	return mq
}

func (q *ModelQuery) Create(attrs ...string) error {
	/*
		1. determine table name
		2. determine field names and values
		3. call Insert()
	*/
	return nil
}

func (q *ModelQuery) Update(attrs ...string) error {
	return nil
}

func (q *ModelQuery) Delete() error {
	return nil
}
