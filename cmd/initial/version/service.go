package version

import "context"

type Service interface {
	CreateOrUpdateVersion(ctx context.Context, version string) error
	GetVersion(ctx context.Context) (string, error)
}

type service struct {
	dao Dao
}

func (s service) GetVersion(ctx context.Context) (string, error) {
	return s.dao.GetVersion(ctx)
}

func (s service) CreateOrUpdateVersion(ctx context.Context, version string) error {
	return s.dao.CreateOrUpdateVersion(ctx, version)
}

func NewService(dao Dao) Service {
	return service{
		dao: dao,
	}
}
