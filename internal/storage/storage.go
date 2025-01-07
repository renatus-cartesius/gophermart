package storage

import (
	"database/sql"
)

type PGStorage struct {
	db *sql.DB
}

func NewPGStorage(db *sql.DB) *PGStorage {
	return &PGStorage{
		db: db,
	}
}
