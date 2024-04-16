package service

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/user/internal/domain"
	"github.com/Duke1616/ecmdb/internal/user/ldapx"
)

type LdapService interface {
	Login(ctx context.Context, req domain.User) (string, error)
}

type ldapService struct {
	ldap ldapx.LdapProvider
}

func NewLdapService(conf ldapx.Config) LdapService {
	return &ldapService{
		ldap: ldapx.NewLdap(conf),
	}
}

func (l ldapService) Login(ctx context.Context, req domain.User) (string, error) {
	// 判断用户是否存在此系统

	// 验证用户
	profile, err := l.ldap.CheckUserPassword(req.User, req.Password)
	if err != nil {
		return "", err
	}

	fmt.Println(profile)

	// 判断账号是否创建, 没有则创建
	return profile.Username, nil
}
