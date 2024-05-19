package mysql

import (
	"reflect"
)

func followPointer(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		return followPointer(v.Elem())
	}
	return v
}

func (sc *Schema) Reflect(v any) bool {
	rv := reflect.ValueOf(v)
	elem := followPointer(rv)

	if elem.Kind() != reflect.Struct /* || elem.IsNil() || !elem.IsValid()*/ {
		return false
	}

	schema := loadDataSchemaInfo(reflect.TypeOf(elem.Interface()))

	sc.Fields = make([]*Field, 0, len(schema.Fields))
	sc.Indices = make([]*Index, 0, len(schema.Fields))

	for i := 0; i < len(schema.Fields); i++ {
		field := schema.Fields[i]
		if field == nil {
			continue
		}
		sc.Fields = append(sc.Fields, &Field{
			Name:          field.ColumnName,
			Type:          field.DataStoreType,
			Nullable:      field.IsNullable,
			AutoIncrement: field.IsAutoincrement,
			DefaultValue:  field.DefaultValue,
			Comment:       field.Comment,
		})

		for _, indexDecl := range field.Indices {
			ok := false
			for _, indexItem := range sc.Indices {
				if indexItem.Name == indexDecl.IndexName {
					indexItem.Columns = append(indexItem.Columns, field.ColumnName)
					ok = true
					break
				}
			}
			if !ok {
				sc.Indices = append(sc.Indices, &Index{
					Name:    indexDecl.IndexName,
					Primary: indexDecl.IndexType == PRIMARY_KEY,
					Unique:  indexDecl.IndexType == UNIQUE,
					Columns: []string{field.ColumnName},
				})
			}
		}
	}

	sc.dataInfo = schema
	return true
}
