package mysql

import (
	"context"
	"database/sql"

	"github.com/acsl-go/logger"
	"github.com/pkg/errors"
)

func InitSchema(dbr *DB, dbw *DB, name string, v any) *Schema {
	schema := &Schema{
		Name:    name,
		Engine:  "InnoDB",
		Collate: "utf8mb4_general_ci",
	}
	ok := schema.Reflect(v)
	if !ok {
		logger.Fatal("%+v", errors.New("Analyzing schema failed"))
	}
	if err := schema.UpdateSchema(context.Background(), dbw); err != nil {
		logger.Fatal("%+v", errors.Wrap(err, "UpdateSchema Failed"))
		panic(err)
	}
	if err := schema.Prepare(dbr, dbw); err != nil {
		logger.Fatal("%+v", errors.Wrap(err, "Prepare Failed"))
		panic(err)
	}
	return schema
}

func PrepareStmt(db *DB, query string) *sql.Stmt {
	stmt, e := db.Ctx.Prepare(query)
	if e != nil {
		logger.Fatal("%+v", errors.Wrap(e, "PrepareStmt Failed"))
		panic(e)
	}
	return stmt
}
