package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/endpoint/internal/domain"
	"github.com/Duke1616/ecmdb/internal/endpoint/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type EndpointRepository interface {
	CreateEndpoint(ctx context.Context, req domain.Endpoint) (int64, error)
	CreateMutilEndpoint(ctx context.Context, req []domain.Endpoint) (int64, error)
	ListEndpoint(ctx context.Context, offset, limit int64, path string) ([]domain.Endpoint, error)
	Total(ctx context.Context, path string) (int64, error)
}

type endpointRepository struct {
	dao dao.EndpointDAO
}

func NewEndpointRepository(dao dao.EndpointDAO) EndpointRepository {
	return &endpointRepository{
		dao: dao,
	}
}

func (repo *endpointRepository) CreateMutilEndpoint(ctx context.Context, req []domain.Endpoint) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (repo *endpointRepository) ListEndpoint(ctx context.Context, offset, limit int64, path string) ([]domain.Endpoint, error) {
	ts, err := repo.dao.ListEndpoint(ctx, offset, limit, path)
	return slice.Map(ts, func(idx int, src dao.Endpoint) domain.Endpoint {
		return repo.toDomain(src)
	}), err
}

func (repo *endpointRepository) Total(ctx context.Context, path string) (int64, error) {
	return repo.dao.Count(ctx, path)
}

func (repo *endpointRepository) CreateEndpoint(ctx context.Context, req domain.Endpoint) (int64, error) {
	return repo.dao.CreateEndpoint(ctx, repo.toEntity(req))
}

func (repo *endpointRepository) toEntity(req domain.Endpoint) dao.Endpoint {
	return dao.Endpoint{
		Path:         req.Path,
		Method:       req.Method,
		Resource:     req.Resource,
		Desc:         req.Desc,
		IsAuth:       req.IsAuth,
		IsAudit:      req.IsAudit,
		IsPermission: req.IsPermission,
	}
}

func (repo *endpointRepository) toDomain(req dao.Endpoint) domain.Endpoint {
	return domain.Endpoint{
		Id:           req.Id,
		Path:         req.Path,
		Method:       req.Method,
		Resource:     req.Resource,
		Desc:         req.Desc,
		IsAuth:       req.IsAuth,
		IsAudit:      req.IsAudit,
		IsPermission: req.IsPermission,
	}
}
