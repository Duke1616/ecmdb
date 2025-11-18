package v193

import (
	"context"

	"github.com/Duke1616/ecmdb/cmd/initial/backup"
	"github.com/Duke1616/ecmdb/cmd/initial/incr"
	"github.com/Duke1616/ecmdb/cmd/initial/ioc"
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/model"
	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/gotomicro/ego/core/elog"
)

type incrV193 struct {
	App    *ioc.App
	logger elog.Component
}

func NewIncrV193(app *ioc.App) incr.InitialIncr {
	return &incrV193{
		App:    app,
		logger: *elog.DefaultLogger,
	}
}

func (i *incrV193) Version() string {
	return "v1.9.3"
}

func (i *incrV193) Commit(ctx context.Context) error {
	i.logger.Info("开始执行 Commit", elog.String("版本", i.Version()))

	// 1. 创建主机模型
	if err := i.createHostModel(ctx); err != nil {
		i.logger.Error("创建主机模型失败", elog.FieldErr(err))
		return err
	}

	// 2. 创建登陆网关模型
	if err := i.createAuthGatewayModel(ctx); err != nil {
		i.logger.Error("创建登陆网关模型失败", elog.FieldErr(err))
		return err
	}

	// 3. 创建关联类型（如果不存在）
	if err := i.createRelationType(ctx); err != nil {
		i.logger.Error("创建关联类型失败", elog.FieldErr(err))
		return err
	}

	// 4. 创建模型关联关系
	if err := i.createModelRelation(ctx); err != nil {
		i.logger.Error("创建模型关联关系失败", elog.FieldErr(err))
		return err
	}

	i.logger.Info("Commit 执行完成", elog.String("版本", i.Version()))
	return nil
}

func (i *incrV193) Rollback(ctx context.Context) error {
	i.logger.Info("开始执行 Rollback", elog.String("版本", i.Version()))

	// TODO: 实现版本回滚逻辑
	// 例如：
	// - 恢复数据库结构
	// - 恢复数据
	// - 恢复配置
	// - 恢复 API 等

	i.logger.Info("Rollback 执行完成", elog.String("版本", i.Version()))
	return nil
}

func (i *incrV193) Before(ctx context.Context) error {
	i.logger.Info("开始执行 Before，备份数据", elog.String("版本", i.Version()))

	// 创建备份管理器
	backupManager := backup.NewBackupManager(i.App)

	// 备份选项
	opts := backup.Options{
		Version:     i.Version(),
		Description: "vv1.9.3 版本更新前备份",
		Tags: map[string]string{
			"type":   "version_upgrade",
			"module": "your_module_name",
		},
	}

	// 使用 backupManager 和 opts 进行备份
	_ = backupManager
	_ = opts

	// TODO: 根据实际需要备份相关数据
	// 例如：
	// _, err := backupManager.BackupMongoCollection(ctx, "your_collection", opts)
	// if err != nil {
	//     return err
	// }
	//
	// _, err = backupManager.BackupMySQLTable(ctx, "your_table", opts)
	// if err != nil {
	//     return err
	// }

	i.logger.Info("Before 执行完成，数据备份完成")
	return nil
}

func (i *incrV193) After(ctx context.Context) error {
	i.logger.Info("开始执行 After，更新版本信息", elog.String("版本", i.Version()))
	if err := i.App.VerSvc.CreateOrUpdateVersion(ctx, i.Version()); err != nil {
		i.logger.Error("更新版本信息失败", elog.FieldErr(err))
		return err
	}
	i.logger.Info("After 执行完成，版本信息已更新")
	return nil
}

// createHostModel 创建主机模型及其字段
func (i *incrV193) createHostModel(ctx context.Context) error {
	i.logger.Info("开始创建主机模型")

	// 检查模型是否已存在
	models, err := i.App.ModelSvc.GetByUids(ctx, []string{"host"})
	if err != nil {
		return err
	}
	if len(models) > 0 {
		i.logger.Info("主机模型已存在，跳过创建")
		return nil
	}

	// 创建主机模型
	_, err = i.App.ModelSvc.Create(ctx, model.Model{
		Name:    "主机",
		UID:     "host",
		Icon:    "icon-host",
		Builtin: true,
	})
	if err != nil {
		return err
	}
	i.logger.Info("主机模型创建成功")

	// 创建字段分组
	groupId, err := i.App.AttributeSvc.CreateAttributeGroup(ctx, attribute.AttributeGroup{
		Name:     "基础属性",
		ModelUid: "host",
		Index:    0,
	})
	if err != nil {
		return err
	}

	// 创建主机模型的字段
	hostFields := []attribute.Attribute{
		{
			ModelUid:  "host",
			FieldUid:  "name",
			FieldName: "名称",
			FieldType: "string",
			Required:  true,
			Display:   true,
			Secure:    false,
			Builtin:   true,
			GroupId:   groupId,
			Index:     0,
		},
		{
			ModelUid:  "host",
			FieldUid:  "ip",
			FieldName: "IP地址",
			FieldType: "string",
			Required:  true,
			Display:   true,
			Secure:    false,
			Builtin:   true,
			GroupId:   groupId,
			Index:     1,
		},
		{
			ModelUid:  "host",
			FieldUid:  "port",
			FieldName: "端口",
			FieldType: "number",
			Required:  false,
			Display:   true,
			Secure:    false,
			Builtin:   true,
			GroupId:   groupId,
			Index:     2,
		},
		{
			ModelUid:  "host",
			FieldUid:  "username",
			FieldName: "用户名",
			FieldType: "string",
			Required:  false,
			Display:   true,
			Secure:    false,
			Builtin:   true,
			GroupId:   groupId,
			Index:     3,
		},
		{
			ModelUid:  "host",
			FieldUid:  "password",
			FieldName: "密码",
			FieldType: "string",
			Required:  false,
			Display:   false,
			Secure:    true, // 加密字段
			Builtin:   true,
			GroupId:   groupId,
			Index:     4,
		},
		{
			ModelUid:  "host",
			FieldUid:  "private_key",
			FieldName: "私钥",
			FieldType: "text",
			Required:  false,
			Display:   false,
			Secure:    true, // 加密字段
			Builtin:   true,
			GroupId:   groupId,
			Index:     5,
		},
		{
			ModelUid:  "host",
			FieldUid:  "auth_type",
			FieldName: "认证类型",
			FieldType: "string",
			Required:  false,
			Display:   true,
			Secure:    false,
			Builtin:   true,
			GroupId:   groupId,
			Index:     6,
		},
	}

	// 批量创建字段
	err = i.App.AttributeSvc.BatchCreateAttribute(ctx, hostFields)
	if err != nil {
		return err
	}
	i.logger.Info("主机模型字段创建成功")

	return nil
}

// createAuthGatewayModel 创建登陆网关模型及其字段
func (i *incrV193) createAuthGatewayModel(ctx context.Context) error {
	i.logger.Info("开始创建登陆网关模型")

	// 检查模型是否已存在
	models, err := i.App.ModelSvc.GetByUids(ctx, []string{"AuthGateway"})
	if err != nil {
		return err
	}
	if len(models) > 0 {
		i.logger.Info("登陆网关模型已存在，跳过创建")
		return nil
	}

	// 创建登陆网关模型
	_, err = i.App.ModelSvc.Create(ctx, model.Model{
		Name:    "登陆网关",
		UID:     "AuthGateway",
		Icon:    "icon-gateway",
		Builtin: true,
	})
	if err != nil {
		return err
	}
	i.logger.Info("登陆网关模型创建成功")

	// 创建字段分组
	groupId, err := i.App.AttributeSvc.CreateAttributeGroup(ctx, attribute.AttributeGroup{
		Name:     "基础属性",
		ModelUid: "AuthGateway",
		Index:    0,
	})
	if err != nil {
		return err
	}

	// 创建登陆网关模型的字段
	gatewayFields := []attribute.Attribute{
		{
			ModelUid:  "AuthGateway",
			FieldUid:  "name",
			FieldName: "名称",
			FieldType: "string",
			Required:  true,
			Display:   true,
			Secure:    false,
			Builtin:   true,
			GroupId:   groupId,
			Index:     0,
		},
		{
			ModelUid:  "AuthGateway",
			FieldUid:  "host",
			FieldName: "主机地址",
			FieldType: "string",
			Required:  true,
			Display:   true,
			Secure:    false,
			Builtin:   true,
			GroupId:   groupId,
			Index:     1,
		},
		{
			ModelUid:  "AuthGateway",
			FieldUid:  "port",
			FieldName: "端口",
			FieldType: "number",
			Required:  false,
			Display:   true,
			Secure:    false,
			Builtin:   true,
			GroupId:   groupId,
			Index:     2,
		},
		{
			ModelUid:  "AuthGateway",
			FieldUid:  "username",
			FieldName: "用户名",
			FieldType: "string",
			Required:  false,
			Display:   true,
			Secure:    false,
			Builtin:   true,
			GroupId:   groupId,
			Index:     3,
		},
		{
			ModelUid:  "AuthGateway",
			FieldUid:  "password",
			FieldName: "密码",
			FieldType: "string",
			Required:  false,
			Display:   false,
			Secure:    true, // 加密字段
			Builtin:   true,
			GroupId:   groupId,
			Index:     4,
		},
		{
			ModelUid:  "AuthGateway",
			FieldUid:  "private_key",
			FieldName: "私钥",
			FieldType: "text",
			Required:  false,
			Display:   false,
			Secure:    true, // 加密字段
			Builtin:   true,
			GroupId:   groupId,
			Index:     5,
		},
		{
			ModelUid:  "AuthGateway",
			FieldUid:  "auth_type",
			FieldName: "认证类型",
			FieldType: "string",
			Required:  false,
			Display:   true,
			Secure:    false,
			Builtin:   true,
			GroupId:   groupId,
			Index:     6,
		},
		{
			ModelUid:  "AuthGateway",
			FieldUid:  "sort",
			FieldName: "排序",
			FieldType: "number",
			Required:  false,
			Display:   true,
			Secure:    false,
			Builtin:   true,
			GroupId:   groupId,
			Index:     7,
		},
	}

	// 批量创建字段
	err = i.App.AttributeSvc.BatchCreateAttribute(ctx, gatewayFields)
	if err != nil {
		return err
	}
	i.logger.Info("登陆网关模型字段创建成功")

	return nil
}

// createRelationType 创建关联类型
func (i *incrV193) createRelationType(ctx context.Context) error {
	i.logger.Info("开始创建关联类型")

	// 创建 default 关联类型
	_, err := i.App.RelationRTSvc.Create(ctx, relation.RelationType{
		Name:           "默认",
		UID:            "default",
		SourceDescribe: "源",
		TargetDescribe: "目标",
	})
	if err != nil {
		// 如果已存在则忽略错误
		i.logger.Warn("关联类型可能已存在", elog.FieldErr(err))
	} else {
		i.logger.Info("关联类型创建成功")
	}

	return nil
}

// createModelRelation 创建模型关联关系
func (i *incrV193) createModelRelation(ctx context.Context) error {
	i.logger.Info("开始创建模型关联关系")

	// 创建 AuthGateway -> host 的关联关系
	_, err := i.App.RelationRMSvc.CreateModelRelation(ctx, relation.ModelRelation{
		SourceModelUID:  "AuthGateway",
		TargetModelUID:  "host",
		RelationTypeUID: "default",
		RelationName:    "AuthGateway_default_host",
		Mapping:         "one_to_many",
	})
	if err != nil {
		i.logger.Warn("模型关联关系可能已存在", elog.FieldErr(err))
	} else {
		i.logger.Info("模型关联关系创建成功")
	}

	return nil
}
