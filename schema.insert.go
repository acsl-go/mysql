package mysql

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
)

func (sc *Schema) Insert(ctx context.Context, v any) error {
	if sc.dataInfo == nil {
		return ErrNoDataInfo
	}

	if sc.stmtInsert == nil {
		return ErrNotReady
	}

	rv := reflect.ValueOf(v)
	elem := followPointer(rv)

	if elem.Kind() != reflect.Struct {
		return ErrInvalidData
	}

	args := make([]any, 0, sc.argsCountInsert)
	for i := 0; i < len(sc.dataInfo.Fields); i++ {
		field := sc.dataInfo.Fields[i]
		if field.IsAutoincrement {
			continue
		}
		args = append(args, field.Serialize(elem))
	}

	r, e := sc.stmtInsert.ExecContext(ctx, args...)
	if e != nil {
		return errors.Wrap(e, "Insert failed")
	}

	if sc.dataInfo.AIField != nil {
		id, e := r.LastInsertId()
		if e != nil {
			return errors.Wrap(e, "Get last insert id failed")
		}
		elem.Field(sc.dataInfo.AIField.FieldIndex).SetInt(id)
	}

	return nil
}
