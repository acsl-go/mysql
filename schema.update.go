package mysql

import (
	"context"
	"reflect"

	drv "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

func (sc *Schema[T]) UpdateEx(ctx context.Context, db IDBLike, data *T, columns ...string) (int64, error) {
	if sc.dbWrite == nil {
		return 0, ErrNotReady
	}

	val := reflect.ValueOf(data).Elem()
	args := make([]any, 0, len(sc.Fields))

	if len(columns) == 0 {
		// Update all fields except primary key
		for _, field := range sc.updateAllFields {
			args = append(args, SerializeField(field.SerializeMethod, val.Field(field.EntityIndex).Interface()))
		}
		for _, field := range sc.primaryFields {
			args = append(args, SerializeField(field.SerializeMethod, val.Field(field.EntityIndex).Interface()))
		}
		r, e := db.ExecContext(ctx, sc.updateAllCmd, args...)
		//r, e := sc.updateAllStmt.ExecContext(ctx, args...)
		if e != nil {
			mysqlErr, ok := e.(*drv.MySQLError)
			if ok {
				if mysqlErr.Number == 1062 {
					return 0, ErrDuplicateKey
				}
			}
			return 0, errors.Wrap(e, "Update failed")
		}
		if n, e := r.RowsAffected(); e != nil {
			return 0, errors.Wrap(e, "Get rows affected failed")
		} else {
			return n, nil
		}
	} else {
		s := "UPDATE `" + sc.Name + "` SET "
		for _, column := range columns {
			field, ok := sc.FieldsByColumn[column]
			if !ok {
				return 0, errors.New("Unknown column: " + column)
			}
			if field.IsPrimaryKey {
				return 0, errors.New("Cannot update primary key: " + column)
			}
			s += "`" + column + "` = ?,"
			args = append(args, SerializeField(field.SerializeMethod, val.Field(field.EntityIndex).Interface()))
		}
		s = s[:len(s)-1] + " WHERE " + sc.primaryWhere
		for _, field := range sc.primaryFields {
			args = append(args, SerializeField(field.SerializeMethod, val.Field(field.EntityIndex).Interface()))
		}
		r, e := db.ExecContext(ctx, s, args...)
		if e != nil {
			mysqlErr, ok := e.(*drv.MySQLError)
			if ok {
				if mysqlErr.Number == 1062 {
					return 0, ErrDuplicateKey
				}
			}
			return 0, errors.Wrap(e, "Update failed")
		}
		if n, e := r.RowsAffected(); e != nil {
			return 0, errors.Wrap(e, "Get rows affected failed")
		} else {
			return n, nil
		}
	}
}

func (sc *Schema[T]) Update(ctx context.Context, data *T, columns ...string) (int64, error) {
	return sc.UpdateEx(ctx, sc.dbWrite.Ctx, data, columns...)
}
