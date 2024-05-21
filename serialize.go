package mysql

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

func SerializeField(t uint8, data interface{}) interface{} {
	switch t {
	case NONE:
		return data
	case JSON:
		b, _ := json.Marshal(data)
		return string(b)
	case YAML:
		b, _ := yaml.Marshal(data)
		return string(b)
	default:
		return ""
	}
}

func DeserializeField(t uint8, data string, v *interface{}) {
	switch t {
	case JSON:
		json.Unmarshal([]byte(data), v)
	case YAML:
		yaml.Unmarshal([]byte(data), v)
	}
}
