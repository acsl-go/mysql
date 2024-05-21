package mysql

import "reflect"

func (sc *Schema[T]) fromType(t reflect.Type) {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		panic("FromType: t must be a struct type")
	}
	sc.Fields = make([]*Field, 0)
	sc.primaryWhere = ""
	for i := 0; i < t.NumField(); i++ {
		fieldType := t.Field(i)
		tag, ok := fieldType.Tag.Lookup("db")
		if ok { // Only process fields with a db tag
			field := &Field{
				Indices:     make([]*FieldIndexDecl, 0),
				EntityIndex: i,
			}
			field.FromTag(tag)
			field.CompleteWithType(fieldType.Type)
			sc.Fields = append(sc.Fields, field)
			if field.IsAutoIncrement {
				sc.aiField = field
			}
			if field.IsPrimaryKey {
				sc.primaryFields = append(sc.primaryFields, field)
				sc.primaryWhere += field.Name + " = ? AND "
			}
		}
	}
	if len(sc.primaryWhere) > 5 {
		sc.primaryWhere = sc.primaryWhere[:len(sc.primaryWhere)-5]
	}
	sc.generateIndices()
	sc.generateFieldMap()
	sc.entity = GetEntity[T](sc)
}
