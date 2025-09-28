package repository

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/endpoint/internal/domain"
	"github.com/Duke1616/ecmdb/internal/endpoint/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type EndpointRepository interface {
	// CreateEndpoint 创建单个端点
	CreateEndpoint(ctx context.Context, req domain.Endpoint) (int64, error)
	
	// BatchRegisterByResource 按 Resource 批量注册端点
	// 支持智能同步：插入新端点、更新已存在端点、删除不再存在的端点
	BatchRegisterByResource(ctx context.Context, resource string, req []domain.Endpoint) (int64, error)
	
	// ListEndpoint 获取端点列表
	ListEndpoint(ctx context.Context, offset, limit int64, path string) ([]domain.Endpoint, error)
	
	// Total 获取端点总数
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

func (repo *endpointRepository) BatchRegisterByResource(ctx context.Context, resource string, req []domain.Endpoint) (int64, error) {
	if len(req) == 0 {
		return 0, nil
	}

	// 转换为 DAO 实体，并设置 Resource
	entities := make([]dao.Endpoint, len(req))
	for i, endpoint := range req {
		entities[i] = repo.toEntity(endpoint)
		entities[i].Resource = resource
	}

	// 调用 DAO 的按 Resource 批量同步方法
	return repo.dao.BatchCreateByResource(ctx, resource, entities)
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
