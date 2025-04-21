package mysql

import (
	"context"
	"database/sql"
	"fmt"

	driver "github.com/go-sql-driver/mysql"
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
		return nil, errors.Wrap(err, "Failed to configure database connection")
	}

	if err := db.Ping(); err != nil {
		if sqlerr, ok := err.(*driver.MySQLError); ok {
			if sqlerr.Number == 1049 {
				// Database does not exist, try to create it
				createDBDsn := fmt.Sprintf("%s:%s@tcp(%s)/?charset=utf8mb4&parseTime=true&loc=Local", cfg.Username, cfg.Password, cfg.Addr)
				createDBConn, createErr := sql.Open("mysql", createDBDsn)
				if createErr != nil {
					return nil, errors.Wrap(createErr, "Failed to configure database connection for creating")
				}
				defer createDBConn.Close()

				// Try to ping the connection to ensure it's valid
				if err := createDBConn.Ping(); err != nil {
					return nil, errors.Wrap(err, "Failed to connect to database for creating")
				}

				// Try to create the database
				_, createErr = createDBConn.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", cfg.DataBase))
				if createErr != nil {
					return nil, errors.Wrap(createErr, "Failed to create database")
				}

				// Successfully created the database, now try to connect again
				return NewDB(cfg)
			}
		}
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
