package mysql

import (
	"context"
	"database/sql"
	"reflect"

	"github.com/pkg/errors"
)

// Get a certain record from the database by primary key(s).
func SelectOne[T interface{}](ctx context.Context, sc *Schema, where string, args ...any) (*T, error) {
	var v T
	if e := sc.SelectOne(ctx, &v, where, args...); e != nil {
		return nil, e
	}
	return &v, nil
}

// Get a certain record from the database by primary key(s).
func (sc *Schema) SelectOne(ctx context.Context, v any, where string, args ...any) error {
	if sc.dataInfo == nil {
		return ErrNoDataInfo
	}
	if sc.dbRead == nil {
		return ErrNotReady
	}

	rv := reflect.ValueOf(v)
	elem := followPointer(rv)

	if elem.Kind() != reflect.Struct {
		return ErrInvalidData
	}

	s := "SELECT " + sc.sqlFieldList + " FROM `" + sc.Name + "`"
	if where != "" {
		s += " WHERE " + where
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

	if e := sc.dbRead.Ctx.QueryRowContext(ctx, s, args...).Scan(fields...); e != nil {
		if e == sql.ErrNoRows {
			return ErrNotFound
		}
		return errors.Wrap(e, "SelectOne failed")
	}

	for _, f := range sc.dataInfo.Serializers {
		f.Deserialize(elem, tempStrings[f.SerializerIndex])
	}

	return nil
}
