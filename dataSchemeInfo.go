package mysql

import (
	"reflect"
	"strings"
	"sync"
	"time"
)

/*
The column information could be defined in the struct tag with the following format:
`db:"<column_name> <column_type> [options...]"`
The options could be a set of the following:

	pk						- Primary Key
	ai						- Auto Increment
	null					- Nullable
	unsigned				- Unsigned
	def(<value>)			- Default Value
	json					- Mark the column as json data
	yaml					- Mark the column as yaml data
	unique(<index_name>)	- Mark the column as a part of unique index with the given index name
	index(<index_name>)		- Mark the column as a part of index with the given index name
	comment(<comment_text>) - Append comment for the field

The column_name could be omitted, if omitted, the field name will be used as column name.
The column_type could be omitted, if omitted, the type will be determined by the field type, see below.
Only one primary key could exist in a table, if more than one column is marked as primary key, a composite primary key will be created.
The index_name could be omitted, if omitted, the the column name with a prefix('idx_') will be used as index name.
If more than one column is marked as a part of the same index, a composite index will be created.
Only one index could be defined for a column, the `unique` and `index` option could NOT be used together.
For compatibility reason, json column will be treated as text column in MySQL, and decode to json when query.

The column type could be one of the following:

	tinyint(<length>)		- Tiny Integer, the length is optional, if omitted, the default value 4 will be used
	int(<length>)			- Integer, the length is optional, if omitted, the default value 11 will be used
	bigint(<length>)		- Big Integer, the length is optional, if omitted, the default value 20 will be used
	float 					- Float
	double					- Double
	decimal(<l>, <d>)		- Decimal, the length(l) and decimals(d) are optional, if omitted, the default value 10 and 0 will be used
	varchar(<length>)		- Varchar, the length is optional, if omitted, the default value 64 will be used
	text					- Text 64k
	mediumtext				- Medium Text 16M
	longtext				- Long Text 4G
	blob					- Blob 64k
	mediumblob				- Medium Blob 16M
	longblob				- Long Blob 4G
	timestamp				- Timestamp
	datetime				- Datetime

The column type could be omitted, if omitted, the type will be determined by the field type in the struct with the following rules:

	int8, int16, int32						- int(11)
	int, int64,								- bigint(20)
	uint8, uint16, uint32					- int(11) with `unsigned` option
	uint, uint64							- bigint(20) with `unsigned` option
	float32									- float
	float64									- double
	string									- varchar(64)
	[]byte									- blob
	time.Time								- datetime
	other									- Serialized to json and stored as mediumtext in database
*/

const (
	// NONE for None
	NONE = 0

	// Serialize Types
	JSON = 2
	YAML = 3

	// Index Types
	INDEX       = 1
	UNIQUE      = 2
	PRIMARY_KEY = 3
)

var (
	timeTypeKind = reflect.TypeOf(time.Time{}).Kind()
)

type dataSchemaFieldIndex struct {
	IndexType uint8  // pk | index | unique
	IndexName string // index name
}

type dataSchemaField struct {
	Name            string       // Name of the field in struct
	FieldType       reflect.Kind // Type of the field
	FieldIndex      int          // Field index of the struct
	ColumnName      string       // Name of the column in database
	IsPrimaryKey    bool         // pk
	IsAutoincrement bool         // ai
	IsNullable      bool         // null
	DataStoreType   string       // column_type
	DefaultValue    string       // def()
	SerializeMethod uint8        // json | yaml
	SerializerIndex int          // index in serializers
	ElementType     reflect.Kind // Element type of the array
	Indices         []*dataSchemaFieldIndex
	Comment         string // comment()
}

type dataSchemaInfo struct {
	Fields          []*dataSchemaField
	ByColumName     map[string]*dataSchemaField
	AIField         *dataSchemaField
	PKFields        []*dataSchemaField
	SerializerCount int
	Serializers     []*dataSchemaField
	DataType        reflect.Type
}

var dataSchemaCache = sync.Map{}

func escapeOptionParameter(p string) string {
	s := []byte(p)
	d := make([]byte, len(s))
	j := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			d[j] = s[i+1]
			i++
		} else if s[i] == ')' {
			break
		} else {
			d[j] = s[i]
		}
		j++
	}
	return string(d[:j])
}

// Parse option string like x(y), y should ending with ')', character ')' in y could be escaped with a leading slash (\).
// The return values will be: x, y
func parseOption(option string) (string, string) {
	eox := strings.Index(option, "(")
	if eox < 0 {
		return option, ""
	}
	return option[:eox], escapeOptionParameter((option[eox+1:]))
}

func parseFieldTag(field *dataSchemaField, tag string) {
	parts := strings.Split(tag, " ")
	for _, p := range parts {
		if p == "" {
			continue
		}
		if field.ColumnName == "" {
			field.ColumnName = p
			continue
		}
		option, param := parseOption(p)
		switch option {
		case "pk":
			field.IsPrimaryKey = true
			field.Indices = append(field.Indices, &dataSchemaFieldIndex{IndexType: PRIMARY_KEY, IndexName: "PRIMARY"})
		case "ai":
			field.IsAutoincrement = true
		case "null":
			field.IsNullable = true
		case "unsigned":
			field.DataStoreType += " unsigned"
		case "def":
			field.DefaultValue = param
		case "json":
			field.SerializeMethod = JSON
		case "yaml":
			field.SerializeMethod = YAML
		case "unique":
			if param == "" {
				param = "idx_" + field.Name
			}
			field.Indices = append(field.Indices, &dataSchemaFieldIndex{IndexType: UNIQUE, IndexName: param})
		case "index":
			if param == "" {
				param = "idx_" + field.Name
			}
			field.Indices = append(field.Indices, &dataSchemaFieldIndex{IndexType: INDEX, IndexName: param})
		case "comment":
			field.Comment = param
		case "tinyint":
			field.DataStoreType = "tinyint"
			if param != "" {
				field.DataStoreType += "(" + param + ")"
			} else {
				field.DataStoreType += "(4)"
			}
		case "int":
			field.DataStoreType = "int"
			if param != "" {
				field.DataStoreType += "(" + param + ")"
			} else {
				field.DataStoreType += "(11)"
			}
		case "bigint":
			field.DataStoreType = "bigint"
			if param != "" {
				field.DataStoreType += "(" + param + ")"
			} else {
				field.DataStoreType += "(20)"
			}
		case "float":
			field.DataStoreType = "float"
		case "double":
			field.DataStoreType = "double"
		case "decimal":
			field.DataStoreType = "decimal"
			if param != "" {
				field.DataStoreType += "(" + param + ")"
			} else {
				field.DataStoreType += "(10,0)"
			}
		case "varchar":
			field.DataStoreType = "varchar"
			if param != "" {
				field.DataStoreType += "(" + param + ")"
			} else {
				field.DataStoreType += "(64)"
			}
		case "text":
			field.DataStoreType = "text"
		case "mediumtext":
			field.DataStoreType = "mediumtext"
		case "longtext":
			field.DataStoreType = "longtext"
		case "blob":
			field.DataStoreType = "blob"
		case "mediumblob":
			field.DataStoreType = "mediumblob"
		case "longblob":
			field.DataStoreType = "longblob"
		case "timestamp":
			field.DataStoreType = "timestamp"
		case "datetime":
			field.DataStoreType = "datetime"
		}
	}
}

func loadDataSchemaInfo(v reflect.Type) *dataSchemaInfo {
	if pInfo, ok := dataSchemaCache.Load(v); ok {
		return pInfo.(*dataSchemaInfo)
	}
	info := dataSchemaInfo{}
	info.DataType = v
	fieldCount := v.NumField()
	info.Fields = make([]*dataSchemaField, fieldCount)
	info.ByColumName = make(map[string]*dataSchemaField)
	info.PKFields = make([]*dataSchemaField, 0)
	info.SerializerCount = 0
	info.Serializers = make([]*dataSchemaField, 0)
	for i := 0; i < fieldCount; i++ {
		field := v.Field(i)
		if tag, ok := field.Tag.Lookup("db"); ok {
			info.Fields[i] = &dataSchemaField{
				Name:       field.Name,
				FieldType:  field.Type.Kind(),
				FieldIndex: i,
				Indices:    make([]*dataSchemaFieldIndex, 0),
			}
			parseFieldTag(info.Fields[i], tag)
			if info.Fields[i].ColumnName == "" {
				info.Fields[i].ColumnName = field.Name
			}
			if info.Fields[i].DataStoreType == "" {
				info.Fields[i].SerializeMethod = NONE
				switch field.Type.Kind() {
				case reflect.Int8, reflect.Int16, reflect.Int32:
					info.Fields[i].DataStoreType = "int(11)"
				case reflect.Int, reflect.Int64:
					info.Fields[i].DataStoreType = "bigint(20)"
				case reflect.Uint8, reflect.Uint16, reflect.Uint32:
					info.Fields[i].DataStoreType = "int(11) unsigned"
				case reflect.Uint, reflect.Uint64:
					info.Fields[i].DataStoreType = "bigint(20) unsigned"
				case reflect.Float32:
					info.Fields[i].DataStoreType = "float"
				case reflect.Float64:
					info.Fields[i].DataStoreType = "double"
				case reflect.String:
					info.Fields[i].DataStoreType = "varchar(64)"
				case reflect.Slice:
					if field.Type.Elem().Kind() == reflect.Uint8 {
						info.Fields[i].DataStoreType = "blob"
					} else {
						info.Fields[i].DataStoreType = "mediumtext"
						info.Fields[i].SerializeMethod = JSON
					}
				case timeTypeKind:
					info.Fields[i].DataStoreType = "datetime"
				default:
					info.Fields[i].DataStoreType = "mediumtext"
					info.Fields[i].SerializeMethod = JSON
				}
			}
			info.ByColumName[info.Fields[i].ColumnName] = info.Fields[i]
			if info.Fields[i].IsAutoincrement {
				info.AIField = info.Fields[i]
			}
			if info.Fields[i].IsPrimaryKey {
				info.PKFields = append(info.PKFields, info.Fields[i])
			}
			if info.Fields[i].SerializeMethod != NONE {
				info.Fields[i].SerializerIndex = info.SerializerCount
				info.Serializers = append(info.Serializers, info.Fields[i])
				info.SerializerCount++
			}
		}
	}
	pInfo, _ := dataSchemaCache.LoadOrStore(v, &info)
	return pInfo.(*dataSchemaInfo)
}
