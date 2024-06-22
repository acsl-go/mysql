package mysql

import (
	"context"
	"reflect"

	drv "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

func (sc *Schema[T]) Insert(ctx context.Context, data *T) error {
	if sc.insertStmt == nil {
		return ErrNotReady
	}

	val := reflect.ValueOf(data).Elem()
	args := make([]any, len(sc.insertArgFields))
	for i := 0; i < len(sc.insertArgFields); i++ {
		f := sc.insertArgFields[i]
		args[i] = SerializeField(f.SerializeMethod, val.Field(f.EntityIndex).Interface())
	}

	r, e := sc.insertStmt.ExecContext(ctx, args...)
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
		val.Field(sc.aiField.EntityIndex).SetInt(id)
	}

	return nil
}
