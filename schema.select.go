package mysql

import (
	"context"
)

func (sc *Schema[T]) SelectOne(ctx context.Context, where string, args ...any) (*T, error) {
	return sc.entity.SelectOne(ctx, where, args...)
}

func (sc *Schema[T]) Select(ctx context.Context, where string, args ...any) ([]*T, error) {
	return sc.entity.Select(ctx, where, args...)
}

// SelectPage selects a page of records from the database.
// page_idx shoud be 1-based.
// Return: records, current page index, page size, page count, total count, error
func (sc *Schema[T]) SelectPage(ctx context.Context, page_idx, page_size int64, where string, args ...any) ([]*T, int64, int64, int64, int64, error) {
	return sc.entity.SelectPage(ctx, page_idx, page_size, where, args...)
}
