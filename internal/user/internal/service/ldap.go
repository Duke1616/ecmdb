package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/user/internal/domain"
	"github.com/Duke1616/ecmdb/internal/user/ldapx"
)

type LdapService interface {
	Login(ctx context.Context, req domain.User) (*ldapx.Profile, error)
}

type ldapService struct {
	ldap ldapx.LdapProvider
}

func NewLdapService(conf ldapx.Config) LdapService {
	return &ldapService{
		ldap: ldapx.NewLdap(conf),
	}
}

// Login LDAP 登录
func (l ldapService) Login(ctx context.Context, req domain.User) (*ldapx.Profile, error) {
	profile, err := l.ldap.VerifyUserCredentials(req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	return profile, nil
}
