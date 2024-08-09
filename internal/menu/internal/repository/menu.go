package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/menu/internal/domain"
	"github.com/Duke1616/ecmdb/internal/menu/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type MenuRepository interface {
	CreateMenu(ctx context.Context, req domain.Menu) (int64, error)
	ListMenu(ctx context.Context) ([]domain.Menu, error)
	UpdateMenu(ctx context.Context, req domain.Menu) (int64, error)
}

type menuRepository struct {
	dao dao.MenuDAO
}

func (repo *menuRepository) UpdateMenu(ctx context.Context, req domain.Menu) (int64, error) {
	return repo.dao.UpdateMenu(ctx, repo.toEntity(req))
}

func NewMenuRepository(dao dao.MenuDAO) MenuRepository {
	return &menuRepository{
		dao: dao,
	}
}

func (repo *menuRepository) ListMenu(ctx context.Context) ([]domain.Menu, error) {
	menu, err := repo.dao.ListMenu(ctx)
	return slice.Map(menu, func(idx int, src dao.Menu) domain.Menu {
		return repo.toDomain(src)
	}), err
}

func (repo *menuRepository) CreateMenu(ctx context.Context, req domain.Menu) (int64, error) {
	return repo.dao.CreateMenu(ctx, repo.toEntity(req))
}

func (repo *menuRepository) toEntity(req domain.Menu) dao.Menu {
	return dao.Menu{
		Id:            req.Id,
		Pid:           req.Pid,
		Path:          req.Path,
		Sort:          req.Sort,
		Redirect:      req.Redirect,
		Name:          req.Name,
		Component:     req.Component,
		ComponentPath: req.ComponentPath,
		Status:        req.Status.ToUint8(),
		Type:          req.Type.ToUint8(),
		Meta: dao.Meta{
			Title:       req.Meta.Title,
			IsHidden:    req.Meta.IsHidden,
			IsAffix:     req.Meta.IsAffix,
			IsKeepAlive: req.Meta.IsKeepAlive,
			Icon:        req.Meta.Icon,
		},
		Endpoints: slice.Map(req.Endpoints, func(idx int, src domain.Endpoint) dao.Endpoint {
			return dao.Endpoint{
				Id:     src.Id,
				Path:   src.Path,
				Method: src.Method,
				Desc:   src.Desc,
			}
		}),
	}
}

func (repo *menuRepository) toDomain(req dao.Menu) domain.Menu {
	return domain.Menu{
		Id:            req.Id,
		Pid:           req.Pid,
		Path:          req.Path,
		Sort:          req.Sort,
		Name:          req.Name,
		Component:     req.Component,
		ComponentPath: req.ComponentPath,
		Redirect:      req.Redirect,
		Status:        domain.Status(req.Status),
		Type:          domain.Type(req.Type),
		Meta: domain.Meta{
			Title:       req.Meta.Title,
			IsHidden:    req.Meta.IsHidden,
			IsAffix:     req.Meta.IsAffix,
			IsKeepAlive: req.Meta.IsKeepAlive,
			Icon:        req.Meta.Icon,
		},
		Endpoints: slice.Map(req.Endpoints, func(idx int, src dao.Endpoint) domain.Endpoint {
			return domain.Endpoint{
				Id:     src.Id,
				Path:   src.Path,
				Method: src.Method,
				Desc:   src.Desc,
			}
		}),
	}
}
