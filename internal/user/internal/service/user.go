package service

import (
	"github.com/Duke1616/ecmdb/internal/user/internal/repostory"
)

type Service interface {
}

type service struct {
	repo repostory.UserRepository
}

func NewService(repo repostory.UserRepository) Service {
	return &service{
		repo: repo,
	}
}
