package service

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/template/internal/domain"
)

type Service interface {
	CreateTemplate(ctx context.Context, req domain.Template) error
}

type service struct {
}

func NewService() Service {
	return &service{}
}

func (s *service) CreateTemplate(ctx context.Context, req domain.Template) error {
	fmt.Println(req)

	return nil
}
