package mysql

import "database/sql"

type Schema[T struct{}] struct {
	Name           string
	Fields         []*Field
	Indices        []*Index
	Engine         string
	Collate        string
	Comment        string
	FieldsByColumn map[string]*Field

	aiField *Field

	dbWrite *DB
	dbRead  *DB

	insertStmt      *sql.Stmt
	insertArgFields []*Field
	updateAllStmt   *sql.Stmt
	updateAllFields []*Field
	primaryWhere    string
	primaryFields   []*Field

	entity *Entity[T]
}

func (sc *Schema[T]) Columns() []string {
	columns := make([]string, len(sc.Fields))
	for i, field := range sc.Fields {
		columns[i] = field.Name
	}
	return columns
}

func (sc *Schema[T]) Field(name string) *Field {
	for _, field := range sc.Fields {
		if field.Name == name {
			return field
		}
	}
	return nil
}

func (sc *Schema[T]) Index(name string) *Index {
	if name == "PRIMARY" {
		name = ""
	}
	for _, index := range sc.Indices {
		if index.Name == name || (name == "" && index.Primary) {
			return index
		}
	}
	return nil
}
