package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/user/internal/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

type LdapUserCache interface {
	LPush(ctx context.Context, profiles []domain.Profile) error
	Lrange(ctx context.Context, offset, limit int64) ([]domain.Profile, error)
	Count(ctx context.Context) (int64, error)
}

type ldapUserCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

func NewLdapUserCache(client redis.Cmdable, expiration time.Duration) LdapUserCache {
	return &ldapUserCache{
		client:     client,
		expiration: expiration,
	}
}

func (cache *ldapUserCache) LPush(ctx context.Context, profiles []domain.Profile) error {
	err := cache.client.Del(ctx, cache.key()).Err()
	if err != nil {
		return err
	}

	pipe := cache.client.Pipeline()
	for _, user := range profiles {
		uJson, er := json.Marshal(user)
		if er != nil {
			return er
		}
		cache.client.LPush(ctx, cache.key(), uJson)
	}

	_, err = pipe.Exec(ctx)
	return err
}

func (cache *ldapUserCache) Lrange(ctx context.Context, offset, limit int64) ([]domain.Profile, error) {
	// 设置分页参数
	start := offset
	end := start + limit - 1

	// 使用 start 和 end 进行 Redis LRange 查询
	result, err := cache.client.LRange(ctx, cache.key(), start, end).Result()
	if err != nil {
		return nil, err
	}

	// 将结果转换为 []domain.Profile 类型
	profiles := make([]domain.Profile, len(result))
	for i, item := range result {
		p := domain.Profile{}
		err = json.Unmarshal([]byte(item), &p)
		if err != nil {
			return nil, err
		}
		profiles[i] = p
	}

	return profiles, nil
}

func (cache *ldapUserCache) Count(ctx context.Context) (int64, error) {
	return cache.client.LLen(ctx, cache.key()).Result()
}

func (cache *ldapUserCache) key() string {
	return fmt.Sprintf("ecmdb:user:ldap")
}
