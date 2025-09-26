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

func NewService(dao Dao) Service {
	return service{
		dao: dao,
	}
}
