package mysql

import (
	"context"
	"reflect"

	drv "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

func (sc *Schema[T]) InsertEx(ctx context.Context, db IDBLike, data *T) error {
	val := reflect.ValueOf(data).Elem()
	args := make([]any, len(sc.insertArgFields))
	for i := 0; i < len(sc.insertArgFields); i++ {
		f := sc.insertArgFields[i]
		args[i] = SerializeField(f.SerializeMethod, val.Field(f.EntityIndex).Interface())
	}

	r, e := db.ExecContext(ctx, sc.insertCmd, args...)
	//r, e := sc.insertStmt.ExecContext(ctx, args...)
	if e != nil {
		mysqlErr, ok := e.(*drv.MySQLError)
		if ok {
			if mysqlErr.Number == 1062 {
				return ErrDuplicateKey
			}
		}
		return errors.Wrap(e, "Insert failed")
	}

	if sc.aiField != nil {
		id, e := r.LastInsertId()
		if e != nil {
			return errors.Wrap(e, "Get last insert id failed")
		}
		if sc.aiField.IsUnsigned {
			val.Field(sc.aiField.EntityIndex).SetUint(uint64(id))
		} else {
			val.Field(sc.aiField.EntityIndex).SetInt(id)
		}
	}

	return nil
}

func (sc *Schema[T]) Insert(ctx context.Context, data *T) error {
	return sc.InsertEx(ctx, sc.dbWrite.Ctx, data)
}
