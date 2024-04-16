package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/user/internal/domain"
)

type LdapService interface {
	Login(ctx context.Context, req domain.User) (int64, error)
}

type ldapService struct {
}

func NewLdapService() LdapService {
	return &ldapService{}
}

func (l ldapService) Login(ctx context.Context, req domain.User) (int64, error) {
	//TODO implement me
	panic("implement me")
}
