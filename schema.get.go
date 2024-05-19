package mysql

import (
	"context"
	"database/sql"
	"reflect"

	"github.com/pkg/errors"
)

// Get a certain record from the database by primary key(s).
func (sc *Schema) Get(ctx context.Context, v any) error {
	if sc.dataInfo == nil {
		return ErrNoDataInfo
	}
	if sc.stmtGet == nil {
		return ErrNotReady
	}

	rv := reflect.ValueOf(v)
	elem := followPointer(rv)

	if elem.Kind() != reflect.Struct {
		return ErrInvalidData
	}

	args := make([]any, len(sc.dataInfo.PKFields))
	for i, field := range sc.dataInfo.PKFields {
		args[i] = field.Serialize(elem)
	}

	tempStrings := make([]string, sc.dataInfo.SerializerCount)
	fields := make([]any, len(sc.dataInfo.Fields))
	for i, f := range sc.dataInfo.Fields {
		if f.SerializeMethod != NONE {
			fields[i] = &tempStrings[f.SerializerIndex]
		} else {
			fields[i] = elem.Field(f.FieldIndex).Addr().Interface()
		}
	}

	if e := sc.stmtGet.QueryRowContext(ctx, args...).Scan(fields...); e != nil {
		if e == sql.ErrNoRows {
			return ErrNotFound
		}
		return errors.Wrap(e, "Get failed")
	}

	for _, f := range sc.dataInfo.Serializers {
		f.Deserialize(elem, tempStrings[f.SerializerIndex])
	}

	return nil
}
