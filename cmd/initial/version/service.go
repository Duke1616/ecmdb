package version

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
)

type Service interface {
	// CreateOrUpdateVersion 创建或更新版本信息
	CreateOrUpdateVersion(ctx context.Context, version string) error

	// GetVersion 获取当前版本
	GetVersion(ctx context.Context) (string, error)

	// 菜单相关方法
	SetMenuHash(ctx context.Context, hash string) error
	GetMenuHash(ctx context.Context) (string, error)
}

type service struct {
	dao Dao
}

func (s service) GetVersion(ctx context.Context) (string, error) {
	ver, err := s.dao.GetVersion(ctx)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return "", nil
	}

	if err != nil {
		return "", err
	}

	return ver, nil
}

func (s service) CreateOrUpdateVersion(ctx context.Context, version string) error {
	return s.dao.CreateOrUpdateVersion(ctx, version)
}

func (s service) SetMenuHash(ctx context.Context, hash string) error {
	return s.dao.SetMenuHash(ctx, hash)
}

func (s service) GetMenuHash(ctx context.Context) (string, error) {
	return s.dao.GetMenuHash(ctx)
}

func NewService(dao Dao) Service {
	return service{
		dao: dao,
	}
}
