package mysql

import (
	"context"
	drv "database/sql"
)

type IDBLike interface {
	QueryContext(ctx context.Context, query string, args ...any) (*drv.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *drv.Row
	ExecContext(ctx context.Context, query string, args ...any) (drv.Result, error)
}
