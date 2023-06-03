package main

import "fmt"

type dbType string

var (
	postresDb   dbType = "postgres"
	sqliteDb    dbType = "sqlite"
	sqlServerDb dbType = "sqlserver"
	mysqlDb     dbType = "mysql"
)

func (db *dbType) Set(v string) error {
	switch dbType(v) {
	case postresDb:
		*db = dbType(v)
		return nil
	case sqliteDb, sqlServerDb, mysqlDb:
		return fmt.Errorf("db (%s) not yet supported", v)
	default:
		return fmt.Errorf("invalid db type: %s", v)
	}
}

func (db *dbType) String() string {
	return string(*db)
}
