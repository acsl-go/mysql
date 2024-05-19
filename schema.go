package mysql

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
)

type Schema struct {
	Name    string
	Fields  []*Field
	Indices []*Index
	Engine  string
	Collate string
	Comment string

	// Private
	dataInfo        *dataSchemaInfo
	dbWrite         *DB
	dbRead          *DB
	stmtInsert      *sql.Stmt
	argsCountInsert int
	sqlFieldList    string
	stmtGet         *sql.Stmt
}

func (sc *Schema) Columns() []string {
	columns := make([]string, len(sc.Fields))
	for i, field := range sc.Fields {
		columns[i] = field.Name
	}
	return columns
}

func (sc *Schema) Field(name string) *Field {
	for _, field := range sc.Fields {
		if field.Name == name {
			return field
		}
	}
	return nil
}

func (sc *Schema) Index(name string) *Index {
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

func (sc *Schema) LoadSchema(ctx context.Context, db *DB) error {
	var dbName string
	if e := db.Ctx.QueryRowContext(ctx, "SELECT DATABASE()").Scan(&dbName); e != nil {
		return errors.Wrap(e, "Get database name failed")
	}

	sc.Fields = make([]*Field, 0)
	sc.Indices = make([]*Index, 0)

	if e := db.Ctx.QueryRowContext(ctx, "SELECT `ENGINE`,`TABLE_COLLATION`,`TABLE_COMMENT` FROM `information_schema`.`TABLES` WHERE `TABLE_SCHEMA` = ? AND `TABLE_NAME` = ?", dbName, sc.Name).Scan(&sc.Engine, &sc.Collate, &sc.Comment); e != nil {
		if e == sql.ErrNoRows {
			return ErrNotFound
		}
		return errors.Wrap(e, "Get table info failed")
	}

	rows, e := db.Ctx.QueryContext(ctx, "SELECT `COLUMN_NAME`,`COLUMN_TYPE`,`IS_NULLABLE`,`COLUMN_DEFAULT`,`COLUMN_COMMENT`,`EXTRA` FROM `information_schema`.`COLUMNS` WHERE `TABLE_SCHEMA` = ? AND `TABLE_NAME` = ?", dbName, sc.Name)
	if e != nil {
		return errors.Wrap(e, "Get table columns failed")
	}

	for rows.Next() {
		var field Field
		var extra, isNullable string
		var defaultValue sql.NullString
		if e := rows.Scan(&field.Name, &field.Type, &isNullable, &defaultValue, &field.Comment, &extra); e != nil {
			return errors.Wrap(e, "Scan table columns failed")
		}
		if extra == "auto_increment" {
			field.AutoIncrement = true
		}
		if isNullable == "YES" {
			field.Nullable = true
		}
		if defaultValue.Valid {
			field.DefaultValue = defaultValue.String
		}
		sc.Fields = append(sc.Fields, &field)
	}

	rows, e = db.Ctx.QueryContext(ctx, "SELECT `INDEX_NAME`,`SEQ_IN_INDEX`,`COLUMN_NAME`,`NON_UNIQUE` FROM `information_schema`.`STATISTICS` WHERE `TABLE_SCHEMA` = ? AND `TABLE_NAME` = ?", dbName, sc.Name)
	if e != nil {
		return errors.Wrap(e, "Get table indexs failed")
	}

	idxMap := make(map[string]int)
	for rows.Next() {
		var idxName string
		var idxColumn string
		var seq, nonUnique int

		if e := rows.Scan(&idxName, &seq, &idxColumn, &nonUnique); e != nil {
			return errors.Wrap(e, "Scan table indexs failed")
		}

		if i, ok := idxMap[idxName]; !ok {
			idxMap[idxName] = len(sc.Indices)
			index := Index{Name: idxName, Columns: []string{idxColumn}}
			if index.Name == "PRIMARY" {
				index.Primary = true
			} else if nonUnique == 0 {
				index.Unique = true
			}
			sc.Indices = append(sc.Indices, &index)
		} else {
			sc.Indices[i].Columns = append(sc.Indices[i].Columns, idxColumn)
		}
	}

	return nil
}

func (sc *Schema) CreateSchema(ctx context.Context, db *DB) error {
	var err error
	var sql string
	var args []interface{}

	sql = "CREATE TABLE IF NOT EXISTS `" + sc.Name + "` ("
	for _, field := range sc.Fields {
		sql += "`" + field.Name + "` " + field.Type
		if field.Nullable {
			sql += " NULL"
		} else {
			sql += " NOT NULL"
		}
		if field.AutoIncrement {
			sql += " AUTO_INCREMENT"
		}
		if field.DefaultValue != "" {
			sql += " DEFAULT " + field.DefaultValue
		}
		if field.Comment != "" {
			sql += " COMMENT '" + escape(field.Comment) + "'"
		}
		sql += ","
	}
	for _, index := range sc.Indices {
		if index.Primary {
			sql += "PRIMARY KEY ("
		} else if index.Unique {
			sql += "UNIQUE KEY `" + index.Name + "` ("
		} else {
			sql += "KEY `" + index.Name + "` ("
		}
		for _, column := range index.Columns {
			sql += "`" + column + "`,"
		}
		sql = sql[:len(sql)-1] + "),"
	}
	sql = sql[:len(sql)-1] + ")"
	if sc.Engine != "" {
		sql += " ENGINE=" + sc.Engine
	}

	if sc.Collate != "" {
		sql += " COLLATE=" + sc.Collate
	}

	if sc.Comment != "" {
		sql += " COMMENT='" + escape(sc.Comment) + "'"
	}

	_, err = db.Ctx.ExecContext(ctx, sql, args...)
	if err != nil {
		return err
	}
	return nil
}

func (sc *Schema) UpdateSchema(ctx context.Context, db *DB) error {
	cur := &Schema{Name: sc.Name}
	e := cur.LoadSchema(ctx, db)
	if e != nil {
		if e == ErrNotFound {
			return sc.CreateSchema(ctx, db)
		}
		return e
	}

	sql := ""
	args := make([]interface{}, 0, 10)

	if sc.Engine != cur.Engine {
		sql += " ENGINE = " + sc.Engine
	}

	if sc.Collate != cur.Collate {
		sql += " COLLATE = " + sc.Collate
	}

	if sc.Comment != cur.Comment {
		sql += " COMMENT = '" + escape(sc.Comment) + "'"
	}

	if sql != "" {
		sql = "ALTER TABLE `" + sc.Name + "`" + sql
		_, e = db.Ctx.ExecContext(ctx, sql, args...)
		if e != nil {
			return e
		}
	}

	for _, field := range cur.Fields {
		if sc.Field(field.Name) == nil {
			sql = "ALTER TABLE `" + sc.Name + "` DROP `" + field.Name + "`"
			_, e = db.Ctx.ExecContext(ctx, sql, args...)
			if e != nil {
				return e
			}
		}
	}

	for _, field := range sc.Fields {
		fd := cur.Field(field.Name)
		sql = ""
		if fd == nil {
			sql = "ALTER TABLE `" + sc.Name + "` ADD `" + field.Name + "` " + field.Type
		} else if !fd.Equal(field) {
			sql = "ALTER TABLE `" + sc.Name + "` MODIFY `" + field.Name + "` " + field.Type
		}
		if sql != "" {
			if field.Nullable {
				sql += " NULL"
			} else {
				sql += " NOT NULL"
			}
			if field.AutoIncrement {
				sql += " AUTO_INCREMENT"
			}
			if field.DefaultValue != "" {
				sql += " DEFAULT " + field.DefaultValue
			}
			if field.Comment != "" {
				sql += " COMMENT '" + escape(field.Comment) + "'"
			}
			_, e = db.Ctx.ExecContext(ctx, sql, args...)
			if e != nil {
				return e
			}
		}
	}

	for _, index := range cur.Indices {
		if sc.Index(index.Name) == nil {
			sql = "ALTER TABLE `" + sc.Name + "` DROP INDEX `" + index.Name + "`"
			_, e = db.Ctx.ExecContext(ctx, sql, args...)
			if e != nil {
				return e
			}
		}
	}

	for _, index := range sc.Indices {
		idx := cur.Index(index.Name)
		sql = ""
		if idx == nil {
			if index.Primary {
				sql = "ALTER TABLE `" + sc.Name + "` ADD PRIMARY KEY ("
			} else if index.Unique {
				sql = "ALTER TABLE `" + sc.Name + "` ADD UNIQUE KEY `" + index.Name + "` ("
			} else {
				sql = "ALTER TABLE `" + sc.Name + "` ADD KEY `" + index.Name + "` ("
			}
		} else if !idx.Equal(index) {
			if index.Primary {
				sql = "ALTER TABLE `" + sc.Name + "` DROP PRIMARY KEY, ADD PRIMARY KEY ("
			} else if index.Unique {
				sql = "ALTER TABLE `" + sc.Name + "` DROP INDEX `" + index.Name + "`, ADD UNIQUE KEY `" + index.Name + "` ("
			} else {
				sql = "ALTER TABLE `" + sc.Name + "` DROP INDEX `" + index.Name + "`, ADD KEY `" + index.Name + "` ("
			}
		}
		if sql != "" {
			for _, column := range index.Columns {
				sql += "`" + column + "`,"
			}
			sql = sql[:len(sql)-1] + ")"
			_, e = db.Ctx.ExecContext(ctx, sql, args...)
			if e != nil {
				return e
			}
		}
	}

	return nil
}
