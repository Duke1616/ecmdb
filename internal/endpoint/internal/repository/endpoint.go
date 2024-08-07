package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/endpoint/internal/domain"
	"github.com/Duke1616/ecmdb/internal/endpoint/internal/repository/dao"
)

type EndpointRepository interface {
	CreateEndpoint(ctx context.Context, req domain.Endpoint) (int64, error)
}

type endpointRepository struct {
	dao dao.EndpointDAO
}

func NewEndpointRepository(dao dao.EndpointDAO) EndpointRepository {
	return &endpointRepository{
		dao: dao,
	}
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
