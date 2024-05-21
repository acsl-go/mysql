package mysql

import (
	"reflect"
	"time"
)

// TAG FORMAT: `<name>` or `<name>(<value>)`
// Value should surround by round brackets, the right bracket could be escaped by `\\`

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

The column_name could be omitted, if omitted, the field name will be used as column name and automatic convert to snake format.
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

type tagItem struct {
	Name  string
	Value string
}

func jumpSpace(tag string, i int) int {
	for i < len(tag) && tag[i] == ' ' {
		i++
	}
	return i
}

func readValue(tag string, i int) (string, int) {
	o := ""
	for i < len(tag) {
		if tag[i] == ')' {
			break
		}
		if tag[i] == '\\' {
			i++
		}
		o += string(tag[i])
		i++
	}
	if i >= len(tag) {
		panic("tag format error (2): " + tag)
	}
	i++
	return o, i
}

func readName(tag string, i int) (string, int) {
	o := ""
	for i < len(tag) {
		if tag[i] == ' ' || tag[i] == '(' {
			break
		}
		if tag[i] == '\\' {
			i++
		}
		o += string(tag[i])
		i++
	}
	return o, i
}

func parseTagArguments(tag string) []tagItem {
	var items []tagItem
	i := 0
	for i < len(tag) {
		i = jumpSpace(tag, i)
		if i >= len(tag) {
			break
		}
		if tag[i] == '(' {
			if len(items) == 0 {
				panic("tag format error (1): " + tag)
			}
			items[len(items)-1].Value, i = readValue(tag, i+1)
		} else {
			newItem := tagItem{}
			newItem.Name, i = readName(tag, i)
			items = append(items, newItem)
		}
	}
	return items
}

func (fd *Field) FromTag(tag string) {
	tagItems := parseTagArguments(tag)
	for _, item := range tagItems {
		if fd.Name == "" {
			fd.Name = item.Name
			continue
		}
		switch item.Name {
		case "pk":
			fd.IsPrimaryKey = true
			fd.Indices = append(fd.Indices, &FieldIndexDecl{IndexType: PRIMARY_KEY, IndexName: "PRIMARY"})
		case "ai":
			fd.IsAutoIncrement = true
		case "null":
			fd.IsNullable = true
		case "unsigned":
			if fd.Type == "" {
				panic("unsigned must follow a type")
			}
			fd.Type += " unsigned"
		case "def":
			fd.DefaultValue = item.Value
		case "json":
			fd.SerializeMethod = JSON
		case "yaml":
			fd.SerializeMethod = YAML
		case "unique":
			if item.Value == "" {
				item.Value = "idx_" + fd.Name
			}
			fd.Indices = append(fd.Indices, &FieldIndexDecl{IndexType: UNIQUE, IndexName: item.Value})
		case "index":
			if item.Value == "" {
				item.Value = "idx_" + fd.Name
			}
			fd.Indices = append(fd.Indices, &FieldIndexDecl{IndexType: INDEX, IndexName: item.Value})
		case "comment":
			fd.Comment = item.Value
		case "tinyint":
			fd.Type = "tinyint"
			if item.Value != "" {
				fd.Type += "(" + item.Value + ")"
			} else {
				fd.Type += "(4)"
			}
		case "int":
			fd.Type = "int"
			if item.Value != "" {
				fd.Type += "(" + item.Value + ")"
			} else {
				fd.Type += "(11)"
			}
		case "bigint":
			fd.Type = "bigint"
			if item.Value != "" {
				fd.Type += "(" + item.Value + ")"
			} else {
				fd.Type += "(20)"
			}
		case "float":
			fd.Type = "float"
		case "double":
			fd.Type = "double"
		case "decimal":
			fd.Type = "decimal"
			if item.Value != "" {
				fd.Type += "(" + item.Value + ")"
			} else {
				fd.Type += "(32,8)"
			}
		case "varchar":
			fd.Type = "varchar"
			if item.Value != "" {
				fd.Type += "(" + item.Value + ")"
			} else {
				fd.Type += "(64)"
			}
		case "text":
			fd.Type = "text"
		case "mediumtext":
			fd.Type = "mediumtext"
		case "longtext":
			fd.Type = "longtext"
		case "blob":
			fd.Type = "blob"
		case "mediumblob":
			fd.Type = "mediumblob"
		case "longblob":
			fd.Type = "longblob"
		case "timestamp":
			fd.Type = "timestamp"
		case "datetime":
			fd.Type = "datetime"
		}
	}
}

func (fd *Field) CompleteWithType(t reflect.Type) {
	if fd.Name == "" {
		fd.Name = camelToSnake(t.Name())
	}
	if fd.Type == "" {
		switch t.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32:
			fd.Type = "int(11)"
		case reflect.Int, reflect.Int64:
			fd.Type = "bigint(20)"
		case reflect.Uint8, reflect.Uint16, reflect.Uint32:
			fd.Type = "int(11) unsigned"
		case reflect.Uint, reflect.Uint64:
			fd.Type = "bigint(20) unsigned"
		case reflect.Float32:
			fd.Type = "float"
		case reflect.Float64:
			fd.Type = "double"
		case reflect.String:
			fd.Type = "varchar(64)"
		case reflect.Slice:
			if t.Elem().Kind() == reflect.Uint8 {
				fd.Type = "blob"
			} else {
				fd.Type = "mediumtext"
				fd.SerializeMethod = JSON
			}
		case timeTypeKind:
			fd.Type = "datetime"
		default:
			fd.Type = "mediumtext"
			fd.SerializeMethod = JSON
		}
	}
}
