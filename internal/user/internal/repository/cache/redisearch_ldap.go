package cache

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/user/internal/domain"
	"github.com/RediSearch/redisearch-go/v2/redisearch"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
	"time"
)

const KeyPrefix = "ecmdb:user:ldap:"

type RedisearchLdapUserCache interface {
	Document(ctx context.Context, profiles []domain.Profile) error
	Query(ctx context.Context, keywords string, offset, limit int) ([]domain.Profile, int, error)
}

type redisearchLdapUserCache struct {
	conn   *redisearch.Client
	logger *elog.Component
}

func NewRedisearchLdapUserCache(conn *redisearch.Client) RedisearchLdapUserCache {
	logger := elog.DefaultLogger
	sc := redisearch.NewSchema(redisearch.DefaultOptions).
		AddField(redisearch.NewTextField("username")).
		AddField(redisearch.NewTextField("display_name")).
		AddField(redisearch.NewTextField("title")).
		AddField(redisearch.NewTextField("email")).
		AddField(redisearch.NewTextField("when_created"))

	// 检查索引是否已经存在
	_, err := conn.Info()
	if err != nil {
		indexDefinition := redisearch.NewIndexDefinition().AddPrefix(KeyPrefix)
		if err = conn.CreateIndexWithIndexDefinition(sc, indexDefinition); err != nil {
			logger.Error("redisearch 创建索引失败, 将影响 LDAP 获取用户功能", elog.FieldErr(err))
		}
	}

	return &redisearchLdapUserCache{
		conn:   conn,
		logger: logger,
	}
}

func (cache *redisearchLdapUserCache) Document(ctx context.Context, profiles []domain.Profile) error {
	var docs []redisearch.Document
	existDocs := make(map[string]bool)
	for _, profile := range profiles {
		t, _ := time.Parse("20060102150405.0Z", profile.WhenCreated)
		doc := redisearch.NewDocument(cache.key(profile.Username), 1.0)

		doc.Set("username", profile.Username).
			Set("display_name", profile.DisplayName).
			Set("title", profile.Title).
			Set("email", profile.Email).
			Set("when_created", t.Unix())
		docs = append(docs, doc)

		// 缓存存在的数据
		existDocs[profile.Username] = true
	}

	if err := cache.conn.IndexOptions(redisearch.IndexingOptions{
		NoSave:           false,
		Replace:          true,
		Partial:          false,
		ReplaceCondition: "",
	}, docs...); err != nil {
		return err
	}

	return cache.dropDocument(existDocs)
}

func (cache *redisearchLdapUserCache) dropDocument(existDocs map[string]bool) error {
	query := redisearch.NewQuery("*").SetReturnFields()
	allDocs, _, err := cache.conn.Search(query)
	if err != nil {
		return err
	}

	docIds := slice.FilterMap(allDocs, func(idx int, src redisearch.Document) (string, bool) {
		if _, ok := existDocs[src.Id]; ok {
			return src.Id, true
		}

		return src.Id, false
	})

	if len(docIds) == len(allDocs) {
		return nil
	}

	for _, id := range docIds {
		err = cache.conn.DeleteDocument(id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cache *redisearchLdapUserCache) Query(ctx context.Context, keywords string,
	offset, limit int) ([]domain.Profile, int, error) {
	defer func() {
		if r := recover(); r != nil {
			cache.logger.Info("LDAP 查询数据可能为空，刷新缓存", elog.Any("recover", r))
		}
	}()

	// 判断传递关键字，如果为空查询所有
	raw := "*"
	if keywords != "" {
		// 进行模糊匹配
		raw = fmt.Sprintf("*%s*", keywords)
	}

	query := redisearch.NewQuery(raw).
		Limit(offset, limit).
		SetReturnFields("username", "display_name", "title", "email").
		SetSortBy("when_created", false)

	docs, total, err := cache.conn.Search(query)

	if err != nil {
		return nil, 0, err
	}

	var profiles []domain.Profile
	for _, doc := range docs {
		profile := domain.Profile{
			Username:    doc.Properties["username"].(string),
			DisplayName: doc.Properties["display_name"].(string),
			Title:       doc.Properties["title"].(string),
			Email:       doc.Properties["email"].(string),
		}
		profiles = append(profiles, profile)
	}

	return profiles, total, nil
}

func (cache *redisearchLdapUserCache) key(username string) string {
	return fmt.Sprintf("%s%s", KeyPrefix, username)
}
