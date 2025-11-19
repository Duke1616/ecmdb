package service

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/bootstrap/internal/executor"
	"github.com/Duke1616/ecmdb/internal/bootstrap/structure"
	"github.com/Duke1616/ecmdb/internal/model"
	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
)

// Loader 配置加载器，负责加载和执行配置文件
type Loader interface {
	// LoadFromFile 从文件加载并执行配置
	LoadFromFile(ctx context.Context, filePath string) error

	// LoadFromConfig 从配置对象加载并执行
	LoadFromConfig(ctx context.Context, cfg *structure.Config) error
}

type loader struct {
	parser         Parser
	modelExec      executor.ModelExecutor
	modelGroupExec executor.ModelGroupExecutor
	attributeExec  executor.AttributeExecutor
	relationExec   executor.RelationExecutor
	logger         *elog.Component
}

// NewLoader 创建加载器
func NewLoader(
	modelSvc model.Service,
	modelGroupSvc model.MGService,
	attributeSvc attribute.Service,
	relationRTSvc relation.RTSvc,
	relationRMSvc relation.RMSvc,
) Loader {
	return &loader{
		parser:         NewParser(),
		modelExec:      executor.NewModelExecutor(modelSvc),
		modelGroupExec: executor.NewModelGroupExecutor(modelGroupSvc),
		attributeExec:  executor.NewAttributeExecutor(attributeSvc),
		relationExec:   executor.NewRelationExecutor(relationRTSvc, relationRMSvc),
		logger:         elog.DefaultLogger,
	}
}

// LoadFromFile 从文件加载并执行配置
func (l *loader) LoadFromFile(ctx context.Context, filePath string) error {
	l.logger.Info("开始从文件加载配置", elog.String("文件路径", filePath))

	// 解析配置文件
	cfg, err := l.parser.ParseFile(filePath)
	if err != nil {
		l.logger.Error("解析配置文件失败", elog.FieldErr(err))
		return err
	}

	// 执行配置
	return l.LoadFromConfig(ctx, cfg)
}

// LoadFromConfig 从配置对象加载并执行
func (l *loader) LoadFromConfig(ctx context.Context, cfg *structure.Config) error {
	l.logger.Info("开始加载配置",
		elog.Int("模型分组数量", len(cfg.ModelGroups)),
		elog.Int("模型数量", len(cfg.Models)),
		elog.Int("关联类型数量", len(cfg.RelationTypes)),
		elog.Int("模型关联数量", len(cfg.ModelRelations)))

	// 1. 创建模型分组（如果有）
	if len(cfg.ModelGroups) > 0 {
		if _, err := l.loadModelGroups(ctx, cfg.ModelGroups); err != nil {
			return err
		}
	}

	// 2. 创建关联类型（需要在模型关联之前创建）
	if err := l.loadRelationTypes(ctx, cfg.RelationTypes); err != nil {
		return err
	}

	// 3. 创建模型及其属性
	for _, modelCfg := range cfg.Models {
		if err := l.loadModel(ctx, modelCfg); err != nil {
			l.logger.Error("加载模型失败",
				elog.String("模型UID", modelCfg.UID),
				elog.FieldErr(err))
			return err
		}
	}

	// 4. 创建模型关联关系
	if err := l.loadModelRelations(ctx, cfg.ModelRelations); err != nil {
		return err
	}

	l.logger.Info("配置加载完成")
	return nil
}

// loadModel 加载单个模型及其属性
func (l *loader) loadModel(ctx context.Context, modelCfg structure.ModelConfig) error {
	l.logger.Info("开始加载模型", elog.String("模型UID", modelCfg.UID), elog.String("模型组", modelCfg.GroupName))
	if modelCfg.GroupName == "" {
		return fmt.Errorf("模型组传递不能为空")
	}

	// 1. 获取模型分组 ID
	group, err := l.modelGroupExec.Get(ctx, modelCfg.GroupName)
	if err != nil {
		return err
	}

	// 2. 创建模型 - 转换为 domain.Model
	if _, err = l.modelExec.Execute(ctx, model.Model{
		UID:     modelCfg.UID,
		Name:    modelCfg.Name,
		Icon:    modelCfg.Icon,
		GroupId: group.ID,
		Builtin: modelCfg.Builtin,
	}); err != nil {
		return err
	}

	// 3. 创建属性分组和字段
	if err = l.loadAttributes(ctx, modelCfg); err != nil {
		return err
	}

	l.logger.Info("模型加载成功", elog.String("模型UID", modelCfg.UID))
	return nil
}

// loadAttributes 加载模型的属性分组和字段
func (l *loader) loadAttributes(ctx context.Context, modelCfg structure.ModelConfig) error {
	if len(modelCfg.Attributes.Groups) == 0 {
		l.logger.Info("模型无属性分组，跳过", elog.String("模型UID", modelCfg.UID))
		return nil
	}

	// 组合组数据
	groups := slice.Map(modelCfg.Attributes.Groups, func(idx int, src structure.AttributeGroupConfig) attribute.AttributeGroup {
		return attribute.AttributeGroup{
			Name:     src.Name,
			ModelUid: modelCfg.UID,
			Index:    src.Index,
		}
	})

	// 返回组数据
	attrGroups, err := l.attributeExec.ExecuteGroups(ctx, modelCfg.UID, groups)
	if err != nil {
		return err
	}

	// 构建分组名称到ID的映射
	groupNameToID := slice.ToMap(attrGroups, func(element attribute.AttributeGroup) string {
		return element.Name
	})

	// 2. 为每个分组创建字段
	var attrs []attribute.Attribute
	for _, groupCfg := range modelCfg.Attributes.Groups {
		groupInfo, ok := groupNameToID[groupCfg.Name]
		if !ok {
			l.logger.Warn("分组不存在",
				elog.String("分组名称", groupCfg.Name),
				elog.String("模型UID", modelCfg.UID))
			continue
		}

		attr := slice.Map(groupCfg.Fields, func(idx int, src structure.FieldConfig) attribute.Attribute {
			return attribute.Attribute{
				ModelUid:  modelCfg.UID,
				FieldUid:  src.UID,
				FieldName: src.Name,
				FieldType: src.Type,
				Option:    src.Option,
				Required:  src.Required,
				Display:   src.Display,
				Secure:    src.Secure,
				Builtin:   src.Builtin,
				GroupId:   groupInfo.ID,
				Index:     src.Index,
			}
		})

		attrs = append(attrs, attr...)
	}

	return l.attributeExec.ExecuteFields(ctx, modelCfg.UID, attrs)
}

// loadRelationTypes 加载关联类型
func (l *loader) loadRelationTypes(ctx context.Context, relationTypes []structure.RelationTypeConfig) error {
	if len(relationTypes) == 0 {
		return nil
	}

	rts := slice.Map(relationTypes, func(idx int, src structure.RelationTypeConfig) relation.RelationType {
		return relation.RelationType{
			UID:            src.UID,
			Name:           src.Name,
			SourceDescribe: src.SourceDescribe,
			TargetDescribe: src.TargetDescribe,
		}
	})

	return l.relationExec.ExecuteRelationTypes(ctx, rts)
}

// loadModelRelations 加载模型关联关系
func (l *loader) loadModelRelations(ctx context.Context, modelRelations []structure.ModelRelationConfig) error {
	if len(modelRelations) == 0 {
		return nil
	}

	mrs := slice.Map(modelRelations, func(idx int, src structure.ModelRelationConfig) relation.ModelRelation {
		return relation.ModelRelation{
			SourceModelUID:  src.SourceModelUID,
			TargetModelUID:  src.TargetModelUID,
			RelationTypeUID: src.RelationTypeUID,
			RelationName:    src.RelationName,
			Mapping:         src.Mapping,
		}
	})

	return l.relationExec.ExecuteModelRelations(ctx, mrs)
}

// loadModelGroups 加载模型分组
func (l *loader) loadModelGroups(ctx context.Context, modelGroups []structure.ModelGroupConfig) (map[string]int64, error) {
	if len(modelGroups) == 0 {
		return make(map[string]int64), nil
	}

	// 转换为 domain.ModelGroup
	groups := slice.Map(modelGroups, func(idx int, src structure.ModelGroupConfig) model.ModelGroup {
		return model.ModelGroup{
			Name: src.Name,
		}
	})

	// 创建或查询模型分组
	groupsResp, err := l.modelGroupExec.Execute(ctx, groups)
	if err != nil {
		return nil, err
	}

	// 模型组 key value 返回
	return slice.ToMapV(groupsResp, func(element model.ModelGroup) (string, int64) {
		return element.Name, element.ID
	}), nil
}
