package mysql

import "github.com/pkg/errors"

func (sc *Schema) Prepare(dbRead *DB, dbWrite *DB) error {
	var e error

	if sc.dataInfo == nil {
		return ErrNoDataInfo
	}

	sc.dbRead = dbRead
	sc.dbWrite = dbWrite

	sc.argsCountInsert = 0
	sqla := "INSERT INTO `" + sc.Name + "` ("
	sqlb := " VALUES ("
	for _, field := range sc.dataInfo.Fields {
		if field.IsAutoincrement {
			continue
		}
		sc.argsCountInsert++
		sqla += "`" + field.ColumnName + "`,"
		sqlb += "?,"
	}
	sqla = sqla[:len(sqla)-1] + ")"
	sqlb = sqlb[:len(sqlb)-1] + ")"
	sc.stmtInsert, e = dbWrite.Ctx.Prepare(sqla + sqlb)
	if e != nil {
		return errors.Wrap(e, "Prepare insert failed")
	}

	sql := ""
	for _, field := range sc.dataInfo.Fields {
		sql += "`" + field.ColumnName + "`,"
	}
	sc.sqlFieldList = sql[:len(sql)-1]

	sql = "SELECT " + sc.sqlFieldList + " FROM `" + sc.Name + "` WHERE "
	for _, field := range sc.dataInfo.PKFields {
		sql += "`" + field.ColumnName + "` = ? AND "
	}
	sql = sql[:len(sql)-5] // Remove last " AND "
	sc.stmtGet, e = dbRead.Ctx.Prepare(sql)
	if e != nil {
		return errors.Wrap(e, "Prepare get failed")
	}

	return nil
}
