package repository

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/menu/internal/domain"
	"github.com/Duke1616/ecmdb/internal/menu/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
	"go.mongodb.org/mongo-driver/mongo"
)

type MenuRepository interface {
	CreateMenu(ctx context.Context, req domain.Menu) (int64, error)
	ListMenu(ctx context.Context) ([]domain.Menu, error)
	// ListByPlatform 根据平台获取菜单列表
	ListByPlatform(ctx context.Context, platform string) ([]domain.Menu, error)
	UpdateMenu(ctx context.Context, req domain.Menu) (int64, error)
	FindByIds(ctx context.Context, ids []int64) ([]domain.Menu, error)
	FindById(ctx context.Context, id int64) (domain.Menu, error)
	GetAllMenu(ctx context.Context) ([]domain.Menu, error)
	DeleteMenu(ctx context.Context, id int64) (int64, error)

	UpdateMenuEndpoints(ctx context.Context, id int64, endpoints []domain.Endpoint) (int64, error)

	// InjectMenu 注入菜单数据
	InjectMenu(ctx context.Context, ms []domain.Menu) (*mongo.BulkWriteResult, error)

	// GetMaxSortKeyByPid 获取当前 PID 下最大的排序值
	GetMaxSortKeyByPid(ctx context.Context, pid int64) (int64, error)
	// ListByPid 根据 PID 获取菜单列表
	ListByPid(ctx context.Context, pid int64) ([]domain.Menu, error)

	// UpdateSort 更新单个菜单的排序
	UpdateSort(ctx context.Context, id, pid, sortKey int64) error
	// BatchUpdateSortKey 批量更新排序
	BatchUpdateSortKey(ctx context.Context, items []domain.MenuSortItem) error
}

type menuRepository struct {
	dao dao.MenuDAO
}

func (repo *menuRepository) BatchUpdateSortKey(ctx context.Context, items []domain.MenuSortItem) error {
	return repo.dao.BatchUpdateSortKey(ctx, slice.Map(items, func(idx int, src domain.MenuSortItem) dao.Menu {
		return dao.Menu{
			Id:   src.ID,
			Pid:  src.Pid,
			Sort: src.SortKey,
		}
	}))
}

func (repo *menuRepository) UpdateSort(ctx context.Context, id, pid, sortKey int64) error {
	return repo.dao.UpdateSort(ctx, id, pid, sortKey)
}

func (repo *menuRepository) ListByPid(ctx context.Context, pid int64) ([]domain.Menu, error) {
	menus, err := repo.dao.ListByPid(ctx, pid)
	return slice.Map(menus, func(idx int, src dao.Menu) domain.Menu {
		return repo.toDomain(src)
	}), err
}

func (repo *menuRepository) GetMaxSortKeyByPid(ctx context.Context, pid int64) (int64, error) {
	return repo.dao.GetMaxSortKeyByPid(ctx, pid)
}

func (repo *menuRepository) UpdateMenuEndpoints(ctx context.Context, id int64, endpoints []domain.Endpoint) (int64, error) {
	return repo.dao.UpdateMenuEndpoints(ctx, id, slice.Map(endpoints, func(idx int, src domain.Endpoint) dao.Endpoint {
		return dao.Endpoint{
			Path:     src.Path,
			Method:   src.Method,
			Resource: src.Resource,
			Desc:     src.Desc,
		}
	}))
}

func (repo *menuRepository) ListByPlatform(ctx context.Context, platform string) ([]domain.Menu, error) {
	menus, err := repo.dao.ListByPlatform(ctx, platform)
	return slice.Map(menus, func(idx int, src dao.Menu) domain.Menu {
		return repo.toDomain(src)
	}), err
}

func (repo *menuRepository) DeleteMenu(ctx context.Context, id int64) (int64, error) {
	return repo.dao.DeleteMenu(ctx, id)
}

func (repo *menuRepository) FindById(ctx context.Context, id int64) (domain.Menu, error) {
	menu, err := repo.dao.FindById(ctx, id)
	return repo.toDomain(menu), err
}

func (repo *menuRepository) GetAllMenu(ctx context.Context) ([]domain.Menu, error) {
	menu, err := repo.dao.GetAllMenu(ctx)
	return slice.Map(menu, func(idx int, src dao.Menu) domain.Menu {
		return repo.toDomain(src)
	}), err
}

func (repo *menuRepository) FindByIds(ctx context.Context, ids []int64) ([]domain.Menu, error) {
	menu, err := repo.dao.FindByIds(ctx, ids)
	return slice.Map(menu, func(idx int, src dao.Menu) domain.Menu {
		return repo.toDomain(src)
	}), err
}

func (repo *menuRepository) UpdateMenu(ctx context.Context, req domain.Menu) (int64, error) {
	return repo.dao.UpdateMenu(ctx, repo.toEntity(req))
}

func (repo *menuRepository) InjectMenu(ctx context.Context, ms []domain.Menu) (*mongo.BulkWriteResult, error) {
	return repo.dao.InjectMenu(ctx, slice.Map(ms, func(idx int, src domain.Menu) dao.Menu {
		return repo.toEntity(src)
	}))
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
		Id:        req.Id,
		Pid:       req.Pid,
		Path:      req.Path,
		Sort:      req.Sort,
		Redirect:  req.Redirect,
		Name:      req.Name,
		Component: req.Component,
		Status:    req.Status.ToUint8(),
		Type:      req.Type.ToUint8(),
		Meta: dao.Meta{
			Title:       req.Meta.Title,
			IsHidden:    req.Meta.IsHidden,
			IsAffix:     req.Meta.IsAffix,
			IsKeepAlive: req.Meta.IsKeepAlive,
			Platforms:   req.Meta.Platforms,
			Icon:        req.Meta.Icon,
		},
		Endpoints: slice.Map(req.Endpoints, func(idx int, src domain.Endpoint) dao.Endpoint {
			return dao.Endpoint{
				Path:     src.Path,
				Method:   src.Method,
				Resource: src.Resource,
				Desc:     src.Desc,
			}
		}),
	}
}

func (repo *menuRepository) toDomain(req dao.Menu) domain.Menu {
	return domain.Menu{
		Id:        req.Id,
		Pid:       req.Pid,
		Path:      req.Path,
		Sort:      req.Sort,
		Name:      req.Name,
		Component: req.Component,
		Redirect:  req.Redirect,
		Status:    domain.Status(req.Status),
		Type:      domain.Type(req.Type),
		Meta: domain.Meta{
			Title:       req.Meta.Title,
			IsHidden:    req.Meta.IsHidden,
			IsAffix:     req.Meta.IsAffix,
			Platforms:   req.Meta.Platforms,
			IsKeepAlive: req.Meta.IsKeepAlive,
			Icon:        req.Meta.Icon,
		},
		Endpoints: slice.Map(req.Endpoints, func(idx int, src dao.Endpoint) domain.Endpoint {
			return domain.Endpoint{
				Path:     src.Path,
				Method:   src.Method,
				Resource: src.Resource,
				Desc:     src.Desc,
			}
		}),
	}
}
