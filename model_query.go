package dbx

import (
	"database/sql"
	"reflect"
	"strings"
	"sync"
)

type TableModel interface {
	TableName() string
}

type ModelQuery struct {
	builder      Builder
	model        *structValue
	fieldMapFunc FieldMapFunc
}

func NewModelQuery(model interface{}, fieldMapFunc FieldMapFunc, builder Builder) *ModelQuery {
	m := newStructValue(model, fieldMapFunc)
	return &ModelQuery{
		builder:      builder,
		model:        m,
		fieldMapFunc: fieldMapFunc,
	}
}

func (q *ModelQuery) Insert(attrs ...string) (sql.Result, error) {
	tableName := q.model.tableName
	cols := q.model.fields(attrs...)
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

type fieldInfo struct {
	name   string
	dbName string
	isPK   bool
	path   []int
}

type structInfo struct {
	nameMap   map[string]*fieldInfo
	dbNameMap map[string]*fieldInfo
	pkNames   []string
}

type structValue struct {
	*structInfo
	value     reflect.Value
	tableName string
}

func newStructValue(model interface{}, mapper FieldMapFunc) *structValue {
	value := reflect.ValueOf(model)
	if value.Kind() != reflect.Ptr || value.Elem().Kind() != reflect.Struct || value.IsNil() {
		return nil
	}
	var tableName string
	if tm, ok := model.(TableModel); ok {
		tableName = tm.TableName()
	} else {
		// todo: generate tableName
	}

	si := getStructInfo(reflect.TypeOf(model), mapper)
	return &structValue{
		structInfo: si,
		value:      value,
		tableName:  tableName,
	}
}

func (s *structValue) pk() map[string]interface{} {
	return nil
}

func (s *structValue) fields(attrs ...string) map[string]interface{} {
	return nil
}

type structInfoMapKey struct {
	t reflect.Type
	m reflect.Value
}

var structInfoMap = make(map[structInfoMapKey]*structInfo)
var muStructInfoMap sync.Mutex

func getStructInfo(a reflect.Type, mapper FieldMapFunc) *structInfo {
	muStructInfoMap.Lock()
	defer muStructInfoMap.Unlock()

	key := structInfoMapKey{a, reflect.ValueOf(mapper)}
	if si, ok := structInfoMap[key]; ok {
		return si
	}

	si := &structInfo{}
	buildStructInfo(si, a, make([]int, 0), "", "", mapper)
	structInfoMap[key] = si

	return si
}

func buildStructInfo(si *structInfo, a reflect.Type, path []int, namePrefix, dbNamePrefix string, mapper FieldMapFunc) {
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
			buildStructInfo(si, ft, path2, concat(namePrefix, name), concat(dbNamePrefix, dbName), mapper)
		} else if dbName != "" {
			// non-anonymous scanner or struct field
			fi := &fieldInfo{
				name:   concat(namePrefix, name),
				dbName: concat(dbNamePrefix, dbName),
				isPK:   false,
				path:   path2,
			}
			si.nameMap[fi.name] = fi
			si.dbNameMap[fi.dbName] = fi
			if isPK {
				si.pkNames = append(si.pkNames, fi.dbName)
			}
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
