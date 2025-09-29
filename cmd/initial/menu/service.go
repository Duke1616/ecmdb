package menu

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/cmd/initial/ioc"
	"github.com/gotomicro/ego/core/elog"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/sync/errgroup"
)

// RoleCode 角色相关常量
const (
	RoleCode = "admin"
)

type ChangeSync struct {
	App    *ioc.App
	logger elog.Component
}

func NewChange(app *ioc.App) *ChangeSync {
	return &ChangeSync{
		App:    app,
		logger: *elog.DefaultLogger,
	}
}

// UpdateMenu 智能菜单更新方法（带 MD5 检查）
func (m *ChangeSync) UpdateMenu(ctx context.Context) error {
	m.logger.Info("开始检查菜单是否需要更新")

	var (
		currentHash string
		storedHash  string
		eg          errgroup.Group
	)
	// 1. 计算当前菜单数据的 MD5 哈希值
	eg.Go(func() error {
		var err error
		hashCalculator := NewMenuHashCalculator()
		currentHash, err = hashCalculator.CalculateProjectMenuHash()
		return err
	})
	// 2. 获取存储的菜单哈希值
	eg.Go(func() error {
		var err error
		storedHash, err = m.App.VerSvc.GetMenuHash(ctx)
		return err
	})

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("计算菜单数据哈希失败: %w", err)
	}

	// 3. 比较哈希值
	if currentHash == storedHash {
		m.logger.Info("Hash校验，菜单数据未发生变化，跳过菜单更新")
		return nil
	}

	m.logger.Info("菜单数据已发生变化，开始更新菜单")
	// 4. 执行菜单更新
	result, err := m.App.MenuSvc.InjectMenu(ctx, GetInjectMenus())
	if err != nil {
		m.logger.Error("注入菜单失败", elog.FieldErr(err))
		return fmt.Errorf("注入菜单失败: %w", err)
	}

	// 5. 检查是否有变化
	if HasChanged(result) {
		m.logger.Info("菜单数据已更新",
			elog.Int64("修改数量 ——【UPDATE】", result.ModifiedCount),
			elog.Int64("插入数量 ——【INSERT】", result.InsertedCount),
			elog.Int64("删除数量 ——【DELETE】", result.DeletedCount),
			elog.Int64("变更数量 ——【UPSERT】", result.UpsertedCount))

		// 6、角色添加菜单
		_, err = m.App.RoleSvc.CreateOrUpdateRoleMenuIds(ctx, RoleCode, GetAllMenuIDs())
		if err != nil {
			return err
		}

		// 7. 角色同步权限
		if err = m.App.PermissionSvc.AddPermissionForRole(ctx, RoleCode, GetAllMenuIDs()); err != nil {
			m.logger.Error("权限初始化失败", elog.FieldErr(err))
			return fmt.Errorf("权限初始化失败: %w", err)
		}
	}

	// 7. 更新 Hash 数据
	if err = m.App.VerSvc.SetMenuHash(ctx, currentHash); err != nil {
		m.logger.Error("更新菜单哈希失败", elog.FieldErr(err))
	}

	m.logger.Info("菜单更新完成", elog.String("hash", currentHash))
	return nil
}

func HasChanged(res *mongo.BulkWriteResult) bool {
	if res == nil {
		return false
	}
	return res.ModifiedCount > 0 || res.UpsertedCount > 0 || res.InsertedCount > 0 || res.DeletedCount > 0
}
