package mysql

import (
	"context"

	"github.com/pkg/errors"
)

// Get a certain record from the database by primary key(s).
func (sc *Schema) Count(ctx context.Context, where string, args ...any) (int64, error) {
	if sc.dataInfo == nil {
		return 0, ErrNoDataInfo
	}
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
