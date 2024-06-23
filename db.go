package mysql

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

type DB struct {
	Ctx *sql.DB
	cfg *Config
}

func NewDB(cfg *Config) (*DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true&loc=Local", cfg.Username, cfg.Password, cfg.Addr, cfg.DataBase)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to configure database")
	}

	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "Failed to connect to database")
	}

	return &DB{
		Ctx: db,
		cfg: cfg,
	}, nil
}

func (db *DB) Close() error {
	e := db.Ctx.Close()
	if e != nil {
		return errors.Wrap(e, "Failed to close database")
	}
	return nil
}

func (db *DB) Tx(ctx context.Context, f func(context.Context, *sql.Tx) error) error {
	tx, e := db.Ctx.BeginTx(ctx, nil)
	if e != nil {
		return errors.Wrap(e, "Failed to start transaction")
	}

	if e := f(ctx, tx); e != nil {
		tx.Rollback()
		return e
	}

	if e := tx.Commit(); e != nil {
		return errors.Wrap(e, "Failed to commit transaction")
	}

	return nil
}

func (db *DB) Exec(ctx context.Context, query string, args ...interface{}) (int64, error) {
	r, e := db.Ctx.ExecContext(ctx, query, args...)
	if e != nil {
		return 0, errors.Wrap(e, "Exec failed")
	}
	if n, e := r.RowsAffected(); e != nil {
		return 0, errors.Wrap(e, "Get rows affected failed")
	} else {
		return n, nil
	}
}
