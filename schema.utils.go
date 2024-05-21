package mysql

func (sc *Schema[T]) generateIndices() {
	sc.Indices = make([]*Index, 0)
	for _, field := range sc.Fields {
		for _, indexDecl := range field.Indices {
			ok := false
			for _, indexItem := range sc.Indices {
				if indexItem.Name == indexDecl.IndexName {
					indexItem.Columns = append(indexItem.Columns, field.Name)
					ok = true
					break
				}
			}
			if !ok {
				sc.Indices = append(sc.Indices, &Index{
					Name:    indexDecl.IndexName,
					Primary: indexDecl.IndexType == PRIMARY_KEY,
					Unique:  indexDecl.IndexType == UNIQUE,
					Columns: []string{field.Name},
				})
			}
		}
	}
}

func (sc *Schema[T]) generateFieldMap() {
	sc.FieldsByColumn = make(map[string]*Field)
	for _, field := range sc.Fields {
		sc.FieldsByColumn[field.Name] = field
	}
}
