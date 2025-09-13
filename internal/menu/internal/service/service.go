package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/internal/menu/internal/domain"
	"github.com/Duke1616/ecmdb/internal/menu/internal/event"
	"github.com/Duke1616/ecmdb/internal/menu/internal/repository"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
)

type Service interface {
	CreateMenu(ctx context.Context, req domain.Menu) (int64, error)
	UpdateMenu(ctx context.Context, req domain.Menu) (int64, error)
	ListMenu(ctx context.Context) ([]domain.Menu, error)
	// ListByPlatform 根据平台获取菜单列表
	ListByPlatform(ctx context.Context, platform string) ([]domain.Menu, error)
	GetAllMenu(ctx context.Context) ([]domain.Menu, error)
	FindById(ctx context.Context, id int64) (domain.Menu, error)
	FindByIds(ctx context.Context, ids []int64) ([]domain.Menu, error)
	DeleteMenu(ctx context.Context, id int64) (int64, error)

	// InjectMenu 注入菜单数据
	InjectMenu(ctx context.Context, ms []domain.Menu) error
}

type service struct {
	producer event.MenuChangeEventProducer
	repo     repository.MenuRepository
	logger   *elog.Component
}

func (s *service) ListByPlatform(ctx context.Context, platform string) ([]domain.Menu, error) {
	return s.repo.ListByPlatform(ctx, platform)
}

func (s *service) FindByIds(ctx context.Context, ids []int64) ([]domain.Menu, error) {
	return s.repo.FindByIds(ctx, ids)
}

func (s *service) FindById(ctx context.Context, id int64) (domain.Menu, error) {
	return s.repo.FindById(ctx, id)
}
func (s *service) ListMenu(ctx context.Context) ([]domain.Menu, error) {
	return s.repo.ListMenu(ctx)
}

func (s *service) GetAllMenu(ctx context.Context) ([]domain.Menu, error) {
	return s.repo.ListMenu(ctx)
}

func (s *service) CreateMenu(ctx context.Context, req domain.Menu) (int64, error) {
	id, err := s.repo.CreateMenu(ctx, req)
	if err != nil {
		return id, err
	}

	// 判断 id 不为空，以及有新增接口权限
	if id != 0 {
		go s.sendMenuEvent(event.CREATE, id, req)
	}

	return id, nil
}

func (s *service) DeleteMenu(ctx context.Context, id int64) (int64, error) {
	count, err := s.repo.DeleteMenu(ctx, id)
	if err != nil {
		return id, err
	}

	// TODO 菜单删除，是验证有角色绑定了菜单不允许删除，还是考虑清除与菜单相关的角色、casbin数据
	//if id != 0 {
	//	go s.sendMenuEvent(event.DELETE, id, domain.Menu{})
	//}

	return count, nil
}

func (s *service) UpdateMenu(ctx context.Context, req domain.Menu) (int64, error) {
	// 查看现有数据
	oldMenu, err := s.repo.FindById(ctx, req.Id)
	if err != nil {
		return 0, err
	}

	// 针对老数据生成 map
	pMap := slice.ToMap(oldMenu.Endpoints, func(element domain.Endpoint) string {
		return fmt.Sprintf("%s:%s", element.Path, element.Method)
	})

	// 修改数据
	count, err := s.repo.UpdateMenu(ctx, req)
	if err != nil {
		return count, err
	}

	// 计算新数据是否包含所有老数据
	oldNum := 0
	for _, e := range req.Endpoints {
		key := fmt.Sprintf("%s:%s", e.Path, e.Method)
		if _, ok := pMap[key]; ok {
			oldNum++
		}
	}

	// 判定修改是否有变更API接口，如果没有则直接退出
	if oldNum == len(oldMenu.Endpoints) &&
		len(req.Endpoints) != 0 &&
		len(oldMenu.Endpoints) != 0 &&
		len(oldMenu.Endpoints) == len(req.Endpoints) {
		return count, nil
	}

	// 判断新增数据拥有所有的老数据，并且新增是要大于老数据，那么就新增角色
	// 否则重新生成角色菜单对应API权限
	if oldNum == len(oldMenu.Endpoints) && len(req.Endpoints) > len(oldMenu.Endpoints) {
		go s.sendMenuEvent(event.WRITE, req.Id, req)
	} else {
		go s.sendMenuEvent(event.REWRITE, req.Id, req)
	}

	return count, nil
}

func (s *service) sendMenuEvent(action event.Action, id int64, menu domain.Menu) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	evt := event.MenuEvent{
		Action: action,
		Menu: event.Menu{
			Id: id,
			Endpoints: slice.Map(menu.Endpoints, func(idx int, src domain.Endpoint) event.Endpoint {
				return event.Endpoint{
					Path:   src.Path,
					Method: src.Method,
				}
			}),
		},
	}

	err := s.producer.Produce(ctx, evt)
	if err != nil {
		s.logger.Error("发送菜单变更事件失败",
			elog.FieldErr(err),
			elog.Any("evt", evt),
		)
	}
}

func (s *service) InjectMenu(ctx context.Context, ms []domain.Menu) error {
	return s.repo.InjectMenu(ctx, ms)
}

func NewService(repo repository.MenuRepository, producer event.MenuChangeEventProducer) Service {
	return &service{
		repo:     repo,
		producer: producer,
		logger:   elog.DefaultLogger,
	}
}
