package dbx

type TableModel interface {
	TableName() string
}

type ModelQuery struct {
	model        TableModel
	fieldMapFunc FieldMapFunc
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
