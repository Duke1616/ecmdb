package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/user/internal/domain"
	"github.com/Duke1616/ecmdb/internal/user/internal/repository/cache"
	"github.com/Duke1616/ecmdb/internal/user/ldapx"
)

type LdapService interface {
	Login(ctx context.Context, req domain.User) (domain.Profile, error)
	SearchUserWithPager(ctx context.Context, keywords string, offset, limit int) ([]domain.Profile, int, error)
	RefreshCacheUserWithPager(ctx context.Context) error
}

type ldapService struct {
	ldap  ldapx.LdapProvider
	cache cache.RedisearchLdapUserCache
}

func NewLdapService(conf ldapx.Config, cache cache.RedisearchLdapUserCache) LdapService {
	return &ldapService{
		ldap:  ldapx.NewLdap(conf),
		cache: cache,
	}
}

func (l *ldapService) SearchUserWithPager(ctx context.Context, keywords string,
	offset, limit int) ([]domain.Profile, int, error) {
	return l.cache.Query(ctx, keywords, offset, limit)
}

func (l *ldapService) RefreshCacheUserWithPager(ctx context.Context) error {
	ldapUsers, err := l.ldap.SearchUserWithPaging()
	if err != nil {
		return err
	}

	return l.cache.Document(ctx, ldapUsers)
}

// Login LDAP 登录
func (l *ldapService) Login(ctx context.Context, req domain.User) (domain.Profile, error) {
	profile, err := l.ldap.VerifyUserCredentials(req.Username, req.Password)
	if err != nil {
		return domain.Profile{}, err
	}

	return profile, nil
}
