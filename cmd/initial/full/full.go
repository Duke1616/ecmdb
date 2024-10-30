package full

import (
	"context"
	"errors"
	"github.com/Duke1616/ecmdb/internal/role"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

func (i *fullInitial) InitUser() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	// 查找用户是否存在
	_, err := i.App.UserSvc.FindById(ctx, 1)
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return nil
	}

	if err != nil {
		return err
	}

	// 创建用户
	_, err = i.App.UserSvc.FindOrCreateBySystem(ctx, UserName, Password, DisPlayName)
	if err != nil {
		return err
	}

	return nil
}

func (i *fullInitial) InitRole() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	// 查找用户是否存在
	_, err := i.App.RoleSvc.FindByRoleCode(ctx, RoleCode)
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return nil
	}

	if err != nil {
		return err
	}

	// 创建用户
	_, err = i.App.RoleSvc.CreateRole(ctx, role.Role{
		Name:   "超级管理员",
		Code:   RoleCode,
		Status: true,
	})
	if err != nil {
		return err
	}

	return nil
}

func (i *fullInitial) InitMenu() error {
	_, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	return nil
}
