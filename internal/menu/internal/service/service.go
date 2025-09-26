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
	"go.mongodb.org/mongo-driver/mongo"
)

type Service interface {
	// CreateMenu 创建菜单
	CreateMenu(ctx context.Context, req domain.Menu) (int64, error)

	// UpdateMenu 更新菜单
	UpdateMenu(ctx context.Context, req domain.Menu) (int64, error)

	// ListMenu 获取所有菜单
	ListMenu(ctx context.Context) ([]domain.Menu, error)

	// ListByPlatform 根据平台获取菜单列表
	ListByPlatform(ctx context.Context, platform string) ([]domain.Menu, error)

	// GetAllMenu 获取所有菜单
	GetAllMenu(ctx context.Context) ([]domain.Menu, error)

	// FindById 根据 ID 获取详情
	FindById(ctx context.Context, id int64) (domain.Menu, error)

	// FindByIds 根据 IDS 获取菜单组
	FindByIds(ctx context.Context, ids []int64) ([]domain.Menu, error)

	// DeleteMenu 删除指定 ID 菜单
	DeleteMenu(ctx context.Context, id int64) (int64, error)

	// InjectMenu 注入菜单数据
	InjectMenu(ctx context.Context, ms []domain.Menu) (*mongo.BulkWriteResult, error)

	// ChangeMenuEndpoints 变更菜单后端 API 接口
	ChangeMenuEndpoints(ctx context.Context, id int64, action domain.Action, endpoints []domain.Endpoint) (int64, error)
}

type service struct {
	producer event.MenuChangeEventProducer
	repo     repository.MenuRepository
	logger   *elog.Component
}

func (s *service) ChangeMenuEndpoints(ctx context.Context, id int64, action domain.Action, endpoints []domain.Endpoint) (int64, error) {
	// 获取现有菜单信息
	menu, err := s.repo.FindById(ctx, id)
	if err != nil {
		return 0, err
	}

	// 根据 action 类型处理权限变更
	switch action {
	case domain.WRITE:
		// 写入权限：添加新权限（不重复）
		menu.Endpoints = slice.UnionSet(menu.Endpoints, endpoints)
	case domain.DELETE:
		// 删除权限：移除指定权限
		menu.Endpoints = slice.DiffSet(menu.Endpoints, endpoints)
	default:
		return 0, fmt.Errorf("不支持的操作类型: %d", action)
	}

	// 更新菜单权限
	count, err := s.repo.UpdateMenuEndpoints(ctx, id, menu.Endpoints)
	if err != nil {
		return count, err
	}

	// 发送权限变更事件
	s.sendMenuEvent(event.Action(action), id, menu)

	return count, nil
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
	return s.repo.UpdateMenu(ctx, req)
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

func (s *service) InjectMenu(ctx context.Context, ms []domain.Menu) (*mongo.BulkWriteResult, error) {
	return s.repo.InjectMenu(ctx, ms)
}

func NewService(repo repository.MenuRepository, producer event.MenuChangeEventProducer) Service {
	return &service{
		repo:     repo,
		producer: producer,
		logger:   elog.DefaultLogger,
	}
}
