package mysql

import (
	"context"

	"github.com/pkg/errors"
)

// Get record count from the database
func (sc *Schema[T]) Count(ctx context.Context, where string, args ...any) (int64, error) {
	if sc.dbRead == nil {
		return 0, ErrNotReady
	}

	s := "SELECT COUNT(*) FROM `" + sc.Name + "`"
	if where != "" {
		s += " WHERE " + where
	}

	var count int64
	if e := sc.dbRead.Ctx.QueryRowContext(ctx, s, args...).Scan(&count); e != nil {
		return 0, errors.Wrap(e, "SelectOne failed")
	}

	return count, nil
}
