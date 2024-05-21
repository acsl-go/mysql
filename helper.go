package mysql

import (
	"database/sql"

	"github.com/acsl-go/logger"
	"github.com/pkg/errors"
)

func PrepareStmt(db *DB, query string) *sql.Stmt {
	stmt, e := db.Ctx.Prepare(query)
	if e != nil {
		logger.Fatal("%+v", errors.Wrap(e, "PrepareStmt Failed"))
		panic(e)
	}
	return stmt
}
