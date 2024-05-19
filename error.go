package mysql

import "github.com/pkg/errors"

var (
	ErrInvalidData    = errors.New("invalid data")
	ErrNotFound       = errors.New("not found")
	ErrNoDataInfo     = errors.New("dataInfo is missing, call Reflect on Schema first")
	ErrNotReady       = errors.New("stmtInsert is not ready, call Prepare on Schema first")
	ErrNoPrimaryKey   = errors.New("no primary key")
	ErrNoRowsAffected = errors.New("no rows affected")
)
