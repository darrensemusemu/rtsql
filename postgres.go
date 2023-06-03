package main

import "github.com/jmoiron/sqlx"

type postgresDB struct {
	db *sqlx.DB
}

var _ Repository = (*postgresDB)(nil)

func NewPostgresDB(db *sqlx.DB) *postgresDB {
	return &postgresDB{
		db: db,
	}
}

func (p *postgresDB) Ready() bool {
	return p.db.Ping() != nil
}

func (p *postgresDB) OnUpdate() error {
	return nil
}
