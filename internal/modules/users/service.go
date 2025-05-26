package users

import (
	"github.com/SekiroKenjii/go-blog-engine/internal/db"
)

type Service struct {
	repo *db.Repository
}

func NewService() *Service {
	return &Service{repo: db.RepositoryInstance()}
}
