package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/user/internal/domain"
	"github.com/Duke1616/ecmdb/internal/user/internal/repository/cache"
	"github.com/Duke1616/ecmdb/internal/user/ldapx"
	"golang.org/x/sync/errgroup"
)

type LdapService interface {
	Login(ctx context.Context, req domain.User) (domain.Profile, error)
	SearchUserWithPager(ctx context.Context, offset, limit int64) ([]domain.Profile, int64, error)
	RefreshCacheUserWithPager(ctx context.Context) error
}

type ldapService struct {
	ldap  ldapx.LdapProvider
	cache cache.LdapUserCache
}

func NewLdapService(conf ldapx.Config, cache cache.LdapUserCache) LdapService {
	return &ldapService{
		ldap:  ldapx.NewLdap(conf),
		cache: cache,
	}
}

func (l *ldapService) SearchUserWithPager(ctx context.Context, offset, limit int64) ([]domain.Profile, int64, error) {
	var (
		eg       errgroup.Group
		profiles []domain.Profile
		total    int64
	)
	eg.Go(func() error {
		var err error
		profiles, err = l.cache.Lrange(ctx, offset, limit)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = l.cache.Count(ctx)
		return err
	})
	if err := eg.Wait(); err != nil {
		return profiles, total, err
	}
	return profiles, total, nil

}

func (l *ldapService) RefreshCacheUserWithPager(ctx context.Context) error {
	ldapUsers, err := l.ldap.SearchUserWithPaging()
	if err != nil {
		return err
	}

	return l.cache.LPush(ctx, ldapUsers)
}

// Login LDAP 登录
func (l *ldapService) Login(ctx context.Context, req domain.User) (domain.Profile, error) {
	profile, err := l.ldap.VerifyUserCredentials(req.Username, req.Password)
	if err != nil {
		return domain.Profile{}, err
	}

	return profile, nil
}
