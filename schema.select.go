package mysql

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
)

func Select[T interface{}](ctx context.Context, sc *Schema, where string, args ...any) ([]T, error) {
	if sc.dataInfo == nil {
		return nil, ErrNoDataInfo
	}
	if sc.dbRead == nil {
		return nil, ErrNotReady
	}

	xv := make([]T, 0)

	s := "SELECT " + sc.sqlFieldList + " FROM `" + sc.Name + "`"
	if where != "" {
		s += " WHERE " + where
	}

	rows, e := sc.dbRead.Ctx.QueryContext(ctx, s, args...)
	if e != nil {
		return nil, errors.Wrap(e, "Select failed")
	}
	defer rows.Close()

	tempStrings := make([]string, sc.dataInfo.SerializerCount)
	for rows.Next() {
		elem := reflect.New(sc.dataInfo.DataType).Elem()
		fields := make([]any, len(sc.dataInfo.Fields))
		for i, f := range sc.dataInfo.Fields {
			if f.SerializeMethod != NONE {
				fields[i] = &tempStrings[f.SerializerIndex]
			} else {
				fields[i] = elem.Field(f.FieldIndex).Addr().Interface()
			}
		}

		if e := rows.Scan(fields...); e != nil {
			return nil, errors.Wrap(e, "Select failed")
		}

		for _, f := range sc.dataInfo.Serializers {
			f.Deserialize(elem, tempStrings[f.SerializerIndex])
		}

		xv = append(xv, elem.Addr().Interface().(T))
	}

	return xv, nil
}

func (sc *Schema) Select(ctx context.Context, slice_ptr any, where string, args ...any) error {
	if sc.dataInfo == nil {
		return ErrNoDataInfo
	}
	if sc.dbRead == nil {
		return ErrNotReady
	}

	rv := reflect.ValueOf(slice_ptr)
	if rv.Kind() != reflect.Ptr {
		return ErrInvalidData
	}

	elem := followPointer(rv)
	if elem.Kind() != reflect.Slice {
		return ErrInvalidData
	}

	arr := reflect.SliceOf(reflect.PointerTo(sc.dataInfo.DataType))
	inst := reflect.New(arr).Elem()

	s := "SELECT " + sc.sqlFieldList + " FROM `" + sc.Name + "`"
	if where != "" {
		s += " WHERE " + where
	}

	rows, e := sc.dbRead.Ctx.QueryContext(ctx, s, args...)
	if e != nil {
		return errors.Wrap(e, "Select failed")
	}
	defer rows.Close()

	tempStrings := make([]string, sc.dataInfo.SerializerCount)
	for rows.Next() {
		elem := reflect.New(sc.dataInfo.DataType).Elem()
		fields := make([]any, len(sc.dataInfo.Fields))
		for i, f := range sc.dataInfo.Fields {
			if f.SerializeMethod != NONE {
				fields[i] = &tempStrings[f.SerializerIndex]
			} else {
				fields[i] = elem.Field(f.FieldIndex).Addr().Interface()
			}
		}

		if e := rows.Scan(fields...); e != nil {
			return errors.Wrap(e, "Select failed")
		}

		for _, f := range sc.dataInfo.Serializers {
			f.Deserialize(elem, tempStrings[f.SerializerIndex])
		}

		inst = reflect.Append(inst, elem.Addr())
	}

	rv.Elem().Set(inst)

	return nil
}
