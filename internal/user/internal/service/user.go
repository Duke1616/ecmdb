package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/policy"
	"github.com/Duke1616/ecmdb/internal/user/internal/domain"
	"github.com/Duke1616/ecmdb/internal/user/internal/repository"
	"github.com/Duke1616/ecmdb/pkg/cryptox"
	"github.com/gotomicro/ego/core/elog"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	// FindOrCreateByLdap 查找或创建来自LDAP的用户
	FindOrCreateByLdap(ctx context.Context, req domain.User) (domain.User, error)

	// SyncCreateLdapUser 同步 LDAP 用户
	SyncCreateLdapUser(ctx context.Context, req domain.User) (int64, error)

	// FindOrCreateBySystem 查找或创建来自本系统的用户
	FindOrCreateBySystem(ctx context.Context, user domain.User) (domain.User, error)

	// ListUser 获取用户列表
	ListUser(ctx context.Context, offset, limit int64) ([]domain.User, int64, error)

	// UpdateUser 更新用户
	UpdateUser(ctx context.Context, req domain.User) (int64, error)

	// Login 登陆
	Login(ctx context.Context, username, password string) (domain.User, error)

	// AddRoleBind 绑定角色
	AddRoleBind(ctx context.Context, id int64, roleCodes []string) (int64, error)

	// FindById 通过 ID 检索用户
	FindById(ctx context.Context, id int64) (domain.User, error)

	// FindByIds 通过IDS检索用户
	FindByIds(ctx context.Context, ids []int64) ([]domain.User, error)

	// FindByKeywords 根据 用户名称、用户名 关键字 检索用户列表
	FindByKeywords(ctx context.Context, offset, limit int64, keyword string) ([]domain.User, int64, error)

	// FindByDepartmentId 根据部门ID 查询用户
	FindByDepartmentId(ctx context.Context, offset, limit int64, departmentId int64) ([]domain.User, int64, error)

	// FindByDepartmentIds 根据部门IDs 查询用户
	FindByDepartmentIds(ctx context.Context, departmentIds []int64) ([]domain.User, error)

	// FindByUsername 根据用户名获取用户
	FindByUsername(ctx context.Context, username string) (domain.User, error)

	// FindByUsernames 根据用户名查询用户列表
	FindByUsernames(ctx context.Context, uns []string) ([]domain.User, error)

	// PipelineDepartmentId 根据部门聚合查询用户
	PipelineDepartmentId(ctx context.Context) ([]domain.UserCombination, error)

	// FindByWechatUser 根据 企业微信ID 查询用户
	FindByWechatUser(ctx context.Context, wechatUserId string) (domain.User, error)

	// FindByFeishuUserId 根据 飞书用户ID 查询用户
	FindByFeishuUserId(ctx context.Context, feishuUserId string) (domain.User, error)
}

type service struct {
	repo      repository.UserRepository
	policySvc policy.Service
	logger    *elog.Component
	crypto    cryptox.Crypto
}

func NewService(repo repository.UserRepository, policySvc policy.Service, crypto cryptox.Crypto) Service {
	return &service{
		repo:      repo,
		policySvc: policySvc,
		logger:    elog.DefaultLogger,
		crypto:    crypto,
	}
}

func (s *service) FindByFeishuUserId(ctx context.Context, feishuUserId string) (domain.User, error) {
	return s.repo.FindByFeishuUserId(ctx, feishuUserId)
}

func (s *service) FindByIds(ctx context.Context, ids []int64) ([]domain.User, error) {
	return s.repo.FindByIds(ctx, ids)
}

func (s *service) SyncCreateLdapUser(ctx context.Context, req domain.User) (int64, error) {
	return s.repo.CreatUser(ctx, req)
}

func (s *service) FindByWechatUser(ctx context.Context, wechatUserId string) (domain.User, error) {
	return s.repo.FindByWechatUser(ctx, wechatUserId)
}

func (s *service) FindByUsername(ctx context.Context, username string) (domain.User, error) {
	return s.repo.FindByUsername(ctx, username)
}

func (s *service) PipelineDepartmentId(ctx context.Context) ([]domain.UserCombination, error) {
	return s.repo.PipelineDepartmentId(ctx)
}

func (s *service) UpdateUser(ctx context.Context, req domain.User) (int64, error) {
	return s.repo.UpdateUser(ctx, req)
}

func (s *service) FindByDepartmentId(ctx context.Context, offset, limit int64, departmentId int64) ([]domain.User, int64, error) {
	var (
		eg    errgroup.Group
		us    []domain.User
		total int64
	)
	eg.Go(func() error {
		var err error
		us, err = s.repo.FindByDepartmentId(ctx, offset, limit, departmentId)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.TotalByDepartmentId(ctx, departmentId)
		return err
	})
	if err := eg.Wait(); err != nil {
		return us, total, err
	}
	return us, total, nil
}

func (s *service) FindByDepartmentIds(ctx context.Context, departmentIds []int64) ([]domain.User, error) {
	return s.repo.FindByDepartmentIds(ctx, departmentIds)
}

func (s *service) FindByUsernames(ctx context.Context, uns []string) ([]domain.User, error) {
	// TODO 不能返回错误，会导致部分地方逻辑处理异常
	if len(uns) == 0 {
		s.logger.Warn("用户传入参数为空")
		return []domain.User{}, nil
	}

	return s.repo.FindByUsernames(ctx, uns)
}

func (s *service) FindByKeywords(ctx context.Context, offset, limit int64, keyword string) ([]domain.User, int64, error) {
	var (
		eg    errgroup.Group
		us    []domain.User
		total int64
	)
	eg.Go(func() error {
		var err error
		us, err = s.repo.FindByKeywords(ctx, offset, limit, keyword)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.TotalByKeywords(ctx, keyword)
		return err
	})
	if err := eg.Wait(); err != nil {
		return us, total, err
	}
	return us, total, nil
}

func (s *service) FindById(ctx context.Context, id int64) (domain.User, error) {
	return s.repo.FindById(ctx, id)
}

func (s *service) Login(ctx context.Context, username, password string) (domain.User, error) {
	// 查看用户是否存在
	u, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		return domain.User{}, fmt.Errorf("查询用户失败，%w", err)
	}

	// 判断密码是否正确
	pwd, err := s.crypto.Decrypt(u.Password)
	if err != nil {
		return domain.User{}, fmt.Errorf("用户：%s, 解密错误", username)
	}

	// 密码不正确
	if pwd != password {
		return domain.User{}, fmt.Errorf("用户：%s, 密码错误", username)
	}

	return u, nil
}

func (s *service) AddRoleBind(ctx context.Context, id int64, roleCodes []string) (int64, error) {
	// 添加绑定
	_, err := s.repo.AddRoleBind(ctx, id, roleCodes)
	if err != nil {
		return 0, err
	}

	// 更新策略
	ok, err := s.policySvc.UpdateFilteredGrouping(ctx, id, roleCodes)
	if err != nil && !ok {
		s.logger.Warn("更新策略失败", elog.FieldErr(err))
		return 0, err
	}
	return id, nil
}

func (s *service) ListUser(ctx context.Context, offset, limit int64) ([]domain.User, int64, error) {
	var (
		eg    errgroup.Group
		us    []domain.User
		total int64
	)
	eg.Go(func() error {
		var err error
		us, err = s.repo.ListUser(ctx, offset, limit)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.Total(ctx)
		return err
	})
	if err := eg.Wait(); err != nil {
		return us, total, err
	}
	return us, total, nil
}

func (s *service) FindOrCreateByLdap(ctx context.Context, req domain.User) (domain.User, error) {
	// 查询数据
	u, err := s.repo.FindByUsername(ctx, req.Username)
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return u, err
	}

	// 创建用户
	id, err := s.repo.CreatUser(ctx, req)
	if err != nil {
		return domain.User{}, err
	}

	req.Id = id
	return req, nil
}

func (s *service) FindOrCreateBySystem(ctx context.Context, req domain.User) (domain.User, error) {
	// 设置用户ID
	var id int64

	// 查询数据
	u, err := s.repo.FindByUsername(ctx, req.Username)
	id = u.Id

	// 函数完成，注入密码
	defer func() {
		if u.Password == "" {
			pwd, er := s.crypto.Encrypt(req.Password)
			if er != nil {
				return
			}

			er = s.repo.UpdatePassword(ctx, id, pwd)
			if er != nil {
				s.logger.Error("修改密码错误", elog.Any("err: ", er))
			}
		}
	}()

	if !errors.Is(err, mongo.ErrNoDocuments) {
		return u, err
	}

	// 生成结构
	user := domain.User{
		Username:     req.Username,
		DisplayName:  req.DisplayName,
		DepartmentId: req.DepartmentId,
		Title:        req.Title,
		Email:        req.Email,
		WechatInfo: domain.WechatInfo{
			UserId: req.WechatInfo.UserId,
		},
		FeishuInfo: domain.FeishuInfo{
			UserId: req.FeishuInfo.UserId,
		},
		Status:     domain.ENABLED,
		CreateType: domain.SYSTEM,
	}

	// 创建用户
	id, err = s.repo.CreatUser(ctx, user)
	if err != nil {
		return domain.User{}, err
	}

	user.Id = id
	return user, nil
}
