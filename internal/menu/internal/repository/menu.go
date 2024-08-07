package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/menu/internal/domain"
	"github.com/Duke1616/ecmdb/internal/menu/internal/repository/dao"
)

type MenuRepository interface {
	CreateMenu(ctx context.Context, req domain.Menu) (int64, error)
}

type menuRepository struct {
	dao dao.MenuDAO
}

func NewMenuRepository(dao dao.MenuDAO) MenuRepository {
	return &menuRepository{
		dao: dao,
	}
}

func (repo *menuRepository) CreateMenu(ctx context.Context, req domain.Menu) (int64, error) {
	return repo.dao.CreateMenu(ctx, repo.toEntity(req))
}

func (repo *menuRepository) toEntity(req domain.Menu) dao.Menu {
	return dao.Menu{
		Pid:    req.Pid,
		Name:   req.Name,
		Path:   req.Path,
		Sort:   req.Sort,
		IsRoot: req.IsRoot,
		Type:   req.Type,
		Meta: dao.Meta{
			Title:       req.Meta.Title,
			IsHidden:    req.Meta.IsHidden,
			IsAffix:     req.Meta.IsAffix,
			IsKeepAlive: req.Meta.IsKeepAlive,
			Icon:        req.Meta.Icon,
		},
		EndpointIds: req.EndpointIds,
	}
}
