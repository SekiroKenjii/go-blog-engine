package db

import (
	"sync"

	dbCtx "github.com/SekiroKenjii/go-blog-engine/internal/db/sqlc/gen"
)

var (
	repoInstance *Repository
	repoOnce     sync.Once
)

type Repository struct {
	*dbCtx.Queries
}

func RepositoryInstance() *Repository {
	repoOnce.Do(func() {
		repoInstance = newRepository()
	})

	return repoInstance
}

func newRepository() *Repository {
	pg := DatabaseInstance().Postgres()

	return &Repository{Queries: dbCtx.New(pg)}
}
