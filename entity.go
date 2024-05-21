package mysql

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type entityField struct {
	FieldIndex      int
	ColumnName      string
	FieldSchema     *Field
	SerializeMethod uint8
}

type Entity[T interface{}] struct {
	fields         []*entityField
	tableNameStr   string
	columnNamesStr string
	dbRead         *DB
}

var (
	entityCache = make(map[reflect.Type]interface{})
)

func GetEntity[T interface{}, S interface{}](schema *Schema[S]) *Entity[T] {
	t := reflect.TypeOf((*T)(nil)).Elem()
	if entity, ok := entityCache[t]; ok {
		return entity.(*Entity[T])
	}
	entity := &Entity[T]{
		fields:         make([]*entityField, 0),
		tableNameStr:   schema.Name,
		columnNamesStr: "",
		dbRead:         schema.dbRead,
	}
	for i := 0; i < t.NumField(); i++ {
		fieldType := t.Field(i)
		if tag, ok := fieldType.Tag.Lookup("db"); ok {
			field := &entityField{FieldIndex: i}
			if tag != "" {
				tag_parts := strings.Split(tag, " ")
				field.ColumnName = tag_parts[0]
			}
			if field.ColumnName == "" {
				field.ColumnName = camelToSnake(fieldType.Name)
			}
			fs, ok := schema.FieldsByColumn[field.ColumnName]
			if !ok {
				panic("field of " + t.Name() + " not found in schema: " + field.ColumnName)
			}
			field.FieldSchema = fs
			field.SerializeMethod = fs.SerializeMethod
			entity.fields = append(entity.fields, field)
			entity.columnNamesStr += "`" + field.ColumnName + "`,"
		}
	}
	entity.columnNamesStr = entity.columnNamesStr[:len(entity.columnNamesStr)-1] // remove last comma
	entityCache[t] = entity
	return entity
}

type rowLike interface {
	Scan(dest ...interface{}) error
}

func (ent *Entity[T]) scan(r rowLike, data *T) error {
	val := reflect.ValueOf(data).Elem()
	args := make([]interface{}, len(ent.fields))
	for i, field := range ent.fields {
		if field.SerializeMethod == JSON || field.SerializeMethod == YAML {
			args[i] = new(string)
		} else {
			args[i] = val.Field(field.FieldIndex).Addr().Interface()
		}
	}
	e := r.Scan(args...)
	if e != nil {
		return errors.Wrap(e, "scan failed")
	}
	for i, field := range ent.fields {
		if field.SerializeMethod == JSON {
			json.Unmarshal([]byte(args[i].(string)), val.Field(field.FieldIndex).Addr().Interface())
		} else if field.SerializeMethod == YAML {
			yaml.Unmarshal([]byte(args[i].(string)), val.Field(field.FieldIndex).Addr().Interface())
		}
	}
	return nil
}
