package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/user/internal/domain"
	"github.com/Duke1616/ecmdb/internal/user/internal/repository"
	"github.com/Duke1616/ecmdb/pkg/cryptox"
	"github.com/gotomicro/ego/core/elog"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	FindOrCreateByLdap(ctx context.Context, req domain.User) (domain.User, error)
	FindOrCreateBySystem(ctx context.Context, username, password, displayName string) (domain.User, error)
	ListUser(ctx context.Context, offset, limit int64) ([]domain.User, int64, error)
	Login(ctx context.Context, username, password string) (domain.User, error)
	AddRoleBind(ctx context.Context, id int64, roleCodes []string) (int64, error)
	FindById(ctx context.Context, id int64) (domain.User, error)
}

type service struct {
	repo   repository.UserRepository
	logger *elog.Component
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
	aesKey := viper.Get("crypto_aes_key").(string)
	pwd, err := cryptox.DecryptAES[string](aesKey, u.Password)
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
	return s.repo.AddRoleBind(ctx, id, roleCodes)
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

func NewService(repo repository.UserRepository) Service {
	return &service{
		repo:   repo,
		logger: elog.DefaultLogger,
	}
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

func (s *service) FindOrCreateBySystem(ctx context.Context, username, password, displayName string) (domain.User, error) {
	// 设置用户ID
	var id int64

	// 查询数据
	u, err := s.repo.FindByUsername(ctx, username)
	id = u.Id
	// 函数完成，注入密码
	defer func() {
		if u.Password == "" {
			pwd, er := encryptAES(password)
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
		Username:    username,
		DisplayName: displayName,
		Status:      domain.ENABLED,
		CreateType:  domain.SYSTEM,
	}

	// 创建用户
	id, err = s.repo.CreatUser(ctx, user)
	if err != nil {
		return domain.User{}, err
	}

	user.Id = id
	return user, nil
}

func encryptAES(passwork string) (string, error) {
	aesKey := viper.Get("crypto_aes_key").(string)
	return cryptox.EncryptAES(aesKey, passwork)
}
