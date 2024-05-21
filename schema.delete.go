package mysql

import (
	"context"

	"github.com/pkg/errors"
)

func (sc *Schema[T]) Delete(ctx context.Context, where string, args ...any) (int64, error) {
	if sc.dbRead == nil {
		return 0, ErrNotReady
	}

	s := "DELETE FROM `" + sc.Name + "`"
	if where != "" {
		s += " WHERE " + where
	}

	if r, e := sc.dbRead.Ctx.ExecContext(ctx, s, args...); e != nil {
		return 0, errors.Wrap(e, "SelectOne failed")
	} else {
		c, e := r.RowsAffected()
		if e != nil {
			return 0, errors.Wrap(e, "Get rows affected failed")
		}
		return c, nil
	}
}
