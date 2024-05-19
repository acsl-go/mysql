package mysql

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
)

// Update updates a row in the table.
// It's possible to update a certain columns by providing the column names.
func (sc *Schema) Update(ctx context.Context, v any, columns ...string) (int64, error) {
	if sc.dataInfo == nil {
		return 0, ErrNoDataInfo
	}

	if sc.dbWrite == nil {
		return 0, ErrNotReady
	}

	if len(sc.dataInfo.PKFields) == 0 {
		return 0, ErrNoPrimaryKey
	}

	rv := reflect.ValueOf(v)
	elem := followPointer(rv)

	if elem.Kind() != reflect.Struct {
		return 0, ErrInvalidData
	}

	sql := "UPDATE `" + sc.Name + "` SET "
	args := make([]any, 0, len(sc.dataInfo.Fields))

	if len(columns) == 0 {
		// Update all
		for _, field := range sc.dataInfo.Fields {
			if field.IsPrimaryKey {
				continue
			}
			sql += "`" + field.ColumnName + "` = ?,"
			args = append(args, field.Serialize(elem))
		}
		sql = sql[:len(sql)-1] // Remove last comma
	} else {
		// Update certain columns
		for _, column := range columns {
			field, ok := sc.dataInfo.ByColumName[column]
			if !ok {
				return 0, errors.New("Unknown column: " + column)
			}
			if field.IsPrimaryKey {
				return 0, errors.New("Cannot update primary key: " + column)
			}
			sql += "`" + column + "` = ?,"
			args = append(args, field.Serialize(elem))
		}
		sql = sql[:len(sql)-1] // Remove last comma
	}

	sql += " WHERE "
	for _, field := range sc.dataInfo.PKFields {
		sql += "`" + field.ColumnName + "` = ? AND "
		args = append(args, field.Serialize(elem))
	}
	sql = sql[:len(sql)-5] // Remove last " AND "

	r, e := sc.dbWrite.Ctx.ExecContext(ctx, sql, args...)
	if e != nil {
		return 0, errors.Wrap(e, "Update failed")
	}

	if n, e := r.RowsAffected(); e != nil {
		return 0, errors.Wrap(e, "Get rows affected failed")
	} else {
		return n, nil
	}
}
