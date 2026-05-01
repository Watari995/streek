package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewDB(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}
