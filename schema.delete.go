package mysql

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
)

func (sc *Schema) Delete(ctx context.Context, v any, where string, args ...interface{}) (int64, error) {
	if sc.dataInfo == nil {
		return 0, ErrNoDataInfo
	}

	if sc.dbWrite == nil {
		return 0, ErrNotReady
	}

	rv := reflect.ValueOf(v)
	elem := followPointer(rv)

	if elem.Kind() != reflect.Struct {
		return 0, ErrInvalidData
	}

	if where == "" {
		where = " "
		for _, field := range sc.dataInfo.PKFields {
			where += "`" + field.ColumnName + "` = ? AND "
			args = append(args, field.Serialize(elem))
		}
		where = where[:len(where)-5] // Remove last " AND "
	}

	sql := "DELETE FROM `" + sc.Name + "` WHERE " + where

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
