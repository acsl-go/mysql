package mysql

type FieldIndexDecl struct {
	IndexType uint8  // pk | index | unique
	IndexName string // index name
}

type Field struct {
	// Basic information
	Name            string // Column name
	Type            string // Column type in SQL format
	IsPrimaryKey    bool
	IsAutoIncrement bool
	IsNullable      bool
	DefaultValue    string // Default value in SQL format
	Comment         string
	SerializeMethod uint8 // json | yaml | none
	Indices         []*FieldIndexDecl

	EntityIndex int // Index in the entity
}

func (fd *Field) Equal(other *Field) bool {
	if fd.Name != other.Name {
		return false
	}
	if fd.Type != other.Type {
		return false
	}
	if fd.IsNullable != other.IsNullable {
		return false
	}
	if fd.IsAutoIncrement != other.IsAutoIncrement {
		return false
	}
	defVal1 := fd.DefaultValue
	defVal2 := other.DefaultValue
	if defVal1 == "NULL" {
		defVal1 = ""
	}
	if defVal2 == "NULL" {
		defVal2 = ""
	}
	if defVal1 != defVal2 {
		return false
	}
	if fd.Comment != other.Comment {
		return false
	}
	return true
}
