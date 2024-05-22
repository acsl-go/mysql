package mysql

import (
	"context"
	"reflect"

	"github.com/acsl-go/logger"
	"github.com/pkg/errors"
)

func NewSchema[T interface{}](dbr *DB, dbw *DB, name string) *Schema[T] {
	schema := &Schema[T]{
		Name:    name,
		Engine:  "InnoDB",
		Collate: "utf8mb4_general_ci",
		dbRead:  dbr,
		dbWrite: dbw,
	}
	schema.fromType(reflect.TypeOf((*T)(nil)))
	if err := schema.updateSchema(context.Background()); err != nil {
		logger.Fatal("%+v", errors.Wrap(err, "UpdateSchema Failed"))
		panic(err)
	}

	if err := schema.init(); err != nil {
		logger.Fatal("%+v", errors.Wrap(err, "Init Schema Failed"))
		panic(err)
	}

	return schema
}

func (sc *Schema[T]) init() error {
	var e error
	sc.insertArgFields = make([]*Field, 0, len(sc.Fields))
	sqla := "INSERT INTO `" + sc.Name + "` ("
	sqlb := " VALUES ("
	for _, field := range sc.Fields {
		if field.IsAutoIncrement {
			continue
		}
		sc.insertArgFields = append(sc.insertArgFields, field)
		sqla += "`" + field.Name + "`,"
		sqlb += "?,"
	}
	sqla = sqla[:len(sqla)-1] + ")"
	sqlb = sqlb[:len(sqlb)-1] + ")"
	sc.insertStmt, e = sc.dbWrite.Ctx.Prepare(sqla + sqlb)
	if e != nil {
		return errors.Wrap(e, "Prepare insert failed")
	}

	sc.updateAllFields = make([]*Field, 0, len(sc.Fields))
	sqla = "UPDATE `" + sc.Name + "` SET "
	for _, field := range sc.Fields {
		if field.IsPrimaryKey {
			continue
		}
		sc.updateAllFields = append(sc.updateAllFields, field)
		sqla += "`" + field.Name + "` = ?,"
	}
	sqla = sqla[:len(sqla)-1] + " WHERE " + sc.primaryWhere
	sc.updateAllStmt, e = sc.dbWrite.Ctx.Prepare(sqla)
	if e != nil {
		return errors.Wrap(e, "Prepare updateAll failed")
	}
	return nil
}
