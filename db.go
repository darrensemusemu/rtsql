package main

import (
	"context"
	"fmt"
)

type Repository interface {
	Ready(context.Context) bool
	AddTrigger(ctx context.Context, cfg ConfigModelTable) error
	RemoveTrigger(ctx context.Context) error
	OnUpdate() error
	Listen(ctx context.Context, channel string) error
	Close() error
}

type dbType string

const (
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

type dbOnType string

const (
	dbOnAfterInsert dbOnType = "after_insert"
)
