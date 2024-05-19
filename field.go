package mysql

type Field struct {
	// Basic information
	Name          string
	Type          string
	Nullable      bool
	AutoIncrement bool
	DefaultValue  string
	Comment       string
}

func (fd *Field) Equal(other *Field) bool {
	if fd.Name != other.Name {
		return false
	}
	if fd.Type != other.Type {
		return false
	}
	if fd.Nullable != other.Nullable {
		return false
	}
	if fd.AutoIncrement != other.AutoIncrement {
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
