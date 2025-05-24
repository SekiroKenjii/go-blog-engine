package sqlc

import (
	"database/sql"
	"sync"

	"github.com/SekiroKenjii/go-blog-engine/internal/db"
	generated "github.com/SekiroKenjii/go-blog-engine/internal/db/sqlc/gen"
)

var (
	instance *Store
	once     sync.Once
)

type Store struct {
	*generated.Queries
	DB *sql.DB
}

func Instance() *Store {
	once.Do(func() {
		instance = newStore()
	})

	return instance
}

func newStore() *Store {
	db := db.Instance().Postgres

	return &Store{
		DB:      db,
		Queries: generated.New(db),
	}
}
