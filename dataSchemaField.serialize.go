package mysql

import (
	"encoding/json"
	"reflect"

	"gopkg.in/yaml.v3"
)

func (field *dataSchemaField) Serialize(elem reflect.Value) any {
	switch field.SerializeMethod {
	case NONE:
		return elem.Field(field.FieldIndex).Interface()
	case JSON:
		b, _ := json.Marshal(elem.Field(field.FieldIndex).Interface())
		return string(b)
	case YAML:
		b, _ := yaml.Marshal(elem.Field(field.FieldIndex).Interface())
		return string(b)
	default:
		return ""
	}
}

func (field *dataSchemaField) Deserialize(elem reflect.Value, data string) {
	switch field.SerializeMethod {
	case JSON:
		json.Unmarshal([]byte(data), elem.Field(field.FieldIndex).Addr().Interface())
	case YAML:
		yaml.Unmarshal([]byte(data), elem.Field(field.FieldIndex).Addr().Interface())
	}
}
