package db

import (
	"database/sql"
	"time"
	"github.com/google/uuid"
)

type Store interface {
	Querier
	NewToken(id uuid.UUID, expiry time.Time, scope string) (*TokenLoc, error)
}

type SQLStore struct {
	db *sql.DB
	*Queries
}

func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}
