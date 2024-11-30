package mysql

import (
	"context"
	drv "database/sql"

	"github.com/pkg/errors"
)

func (ent *Entity[T]) SelectOne(ctx context.Context, where string, args ...any) (*T, error) {
	sql := "SELECT " + ent.columnNamesStr + " FROM `" + ent.tableNameStr + "`"
	if where != "" {
		sql += " WHERE " + where
	}
	row := ent.dbRead.Ctx.QueryRowContext(ctx, sql, args...)
	v := new(T)
	if e := ent.scan(row, v); e != nil {
		if errors.Is(e, drv.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, e
	}
	return v, nil
}

func (ent *Entity[T]) Select(ctx context.Context, where string, args ...any) ([]*T, error) {
	sql := "SELECT " + ent.columnNamesStr + " FROM `" + ent.tableNameStr + "`"
	if where != "" {
		sql += " WHERE " + where
	}
	rows, e := ent.dbRead.Ctx.QueryContext(ctx, sql, args...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	result := make([]*T, 0)
	for rows.Next() {
		v := new(T)
		if e := ent.scan(rows, v); e != nil {
			return nil, e
		}
		result = append(result, v)
	}
	return result, nil
}

// SelectPage selects a page of records from the database.
// page_idx shoud be 1-based.
// Return: records, current page index, page size, page_count, total count, error
func (ent *Entity[T]) SelectPage(ctx context.Context, page_idx, page_size int64, where string, args ...any) ([]*T, int64, int64, int64, int64, error) {
	sql := "SELECT count(*) FROM `" + ent.tableNameStr + "`"
	if where != "" {
		sql += " WHERE " + where
	}
	var cnt int64
	vargs := make([]interface{}, 0, len(args)+2)
	vargs = append(vargs, args...)
	if e := ent.dbRead.Ctx.QueryRowContext(ctx, sql, args...).Scan(&cnt); e != nil {
		return nil, 0, 0, 0, 0, errors.Wrap(e, "SelectPage failed")
	}

	if page_size < 1 {
		page_size = 1
	}

	if cnt == 0 {
		return make([]*T, 0), 1, page_size, 0, 0, nil
	}

	page_count := cnt / page_size
	if cnt%page_size > 0 {
		page_count++
	}

	if page_idx < 1 {
		page_idx = 1
	}

	/* Allow page_idx > page_count
	if page_idx > page_count {
		page_idx = page_count
	}
	*/

	offset := (page_idx - 1) * page_size

	vargs = append(vargs, offset, page_size)
	sql = "SELECT " + ent.columnNamesStr + " FROM `" + ent.tableNameStr + "` "
	if where != "" {
		sql += "WHERE " + where
	}
	sql += " LIMIT ?, ?"
	rows, e := ent.dbRead.Ctx.QueryContext(ctx, sql, vargs...)
	if e != nil {
		return nil, 0, 0, 0, 0, errors.Wrap(e, "SelectPage failed")
	}
	defer rows.Close()
	result := make([]*T, 0)
	for rows.Next() {
		v := new(T)
		if e := ent.scan(rows, v); e != nil {
			return nil, 0, 0, 0, 0, errors.Wrap(e, "SelectPage failed")
		}
		result = append(result, v)
	}
	return result, page_idx, page_size, page_count, cnt, nil
}
