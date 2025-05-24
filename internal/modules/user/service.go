package user

import (
	"github.com/SekiroKenjii/go-blog-engine/internal/db/sqlc"
)

type Service struct {
	store *sqlc.Store
}

func NewService(store *sqlc.Store) *Service {
	return &Service{store: store}
}
