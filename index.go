package mysql

type Index struct {
	Name    string
	Columns []string
	Primary bool
	Unique  bool
}

func (idx *Index) Equal(other *Index) bool {
	if idx.Primary != other.Primary {
		return false
	}
	if !idx.Primary && idx.Name != other.Name {
		return false
	}
	if idx.Unique != other.Unique {
		return false
	}
	if len(idx.Columns) != len(other.Columns) {
		return false
	}
	for i, column := range idx.Columns {
		if column != other.Columns[i] {
			return false
		}
	}
	return true
}
