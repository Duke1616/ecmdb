package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/Duke1616/ecmdb/internal/domain"
	attribute "github.com/Duke1616/ecmdb/internal/service/attribute"
	model "github.com/Duke1616/ecmdb/internal/service/model"
	resource "github.com/Duke1616/ecmdb/internal/service/resource"
	"github.com/Duke1616/ecmdb/pkg/excelx"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

type dataIOService struct {
	attrSvc  attribute.Service
	resSvc   resource.EncryptedSvc
	modelSvc model.Service
}

// NewService 创建数据交换服务实例
func NewService(
	attrSvc attribute.Service,
	resSvc resource.EncryptedSvc,
	modelSvc model.Service,
) IDataIOService {
	return &dataIOService{
		attrSvc:  attrSvc,
		resSvc:   resSvc,
		modelSvc: modelSvc,
	}
}

// Import 批量导入资源实例 (Resource)
func (s *dataIOService) Import(ctx context.Context, modelUID string, fileData []byte) (importedCount int, err error) {
	// 1. 获取当前模型的字段元数据定义
	attrs, _, err := s.attrSvc.ListAttributes(ctx, modelUID)
	if err != nil {
		return 0, fmt.Errorf("获取模型字段定义失败: %w", err)
	}
	if len(attrs) == 0 {
		return 0, fmt.Errorf("模型 %s 未定义任何属性字段", modelUID)
	}

	// 2. 将元数据映射为声明式 Excel 表格 Schema
	columns := toExcelColumns(attrs)

	// 3. 由通用 excelx 引擎执行反序列化、表头匹配、空行过滤与强类型校验
	records, err := excelx.NewReader(columns).Read(fileData)
	if err != nil {
		return 0, err
	}

	// 4. 将反序列化后的纯数据记录组装为领域实体
	resources := lo.Map(records, func(rec map[string]interface{}, _ int) domain.Resource {
		return domain.Resource{
			ModelUID: modelUID,
			Data:     rec,
		}
	})

	// 5. 分批（Chunking）安全写入存储层，防御 MongoDB 单包 16MB 限制
	const chunkSize = 500
	chunks := lo.Chunk(resources, chunkSize)
	for _, chunk := range chunks {
		if err = s.resSvc.BatchCreateOrUpdate(ctx, chunk); err != nil {
			return 0, fmt.Errorf("批量写入资产失败: %w", err)
		}
	}

	return len(resources), nil
}

// Export 导出资源实例数据 (Resource)
func (s *dataIOService) Export(ctx context.Context, req ExportParams) ([]byte, error) {
	// 1. 获取模型与属性定义
	mdl, attrs, err := s.fetchModelAndAttributes(ctx, req.ModelUID)
	if err != nil {
		return nil, err
	}

	// 2. 字段过滤：若用户指定了目标字段集，则只导出对应字段
	if len(req.Fields) > 0 {
		attrs = lo.Filter(attrs, func(src domain.Attribute, _ int) bool {
			return lo.Contains(req.Fields, src.FieldUid)
		})
	}

	// 3. 依元数据自然排序字段 (name 置首，其余按 SortKey / Index 升序)
	sortedAttrs := sortAttributesByPriority(attrs)

	// 4. 提取目标字段 UID 列表用于资源批量检索
	dstFields := lo.Map(sortedAttrs, func(attr domain.Attribute, _ int) string {
		return attr.FieldUid
	})

	// 5. 分页循环检索符合条件的全量资产
	var allResources []domain.Resource
	offset := int64(0)
	limit := int64(100)

	for {
		resources, _, err1 := s.resSvc.ListResourcesWithFilters(ctx, dstFields, req.ModelUID, req.ResourceIDs, offset, limit, req.FilterGroups)
		if err1 != nil {
			return nil, fmt.Errorf("获取资源列表失败: %w", err1)
		}
		allResources = append(allResources, resources...)

		if len(resources) < int(limit) {
			break
		}
		offset += limit
	}

	// 6. 构造 Excel Schema 与数据记录，由 excelx 引擎统一输出
	columns := toExcelColumns(sortedAttrs)
	records := lo.Map(allResources, func(r domain.Resource, _ int) map[string]interface{} {
		return r.Data
	})

	return excelx.NewWriter(mdl.SheetName(), columns).WriteData(records)
}

// ExportTemplate 导出空白导入模板
func (s *dataIOService) ExportTemplate(ctx context.Context, modelUID string) ([]byte, error) {
	// 1. 获取模型与属性定义
	mdl, attrs, err := s.fetchModelAndAttributes(ctx, modelUID)
	if err != nil {
		return nil, err
	}

	// 2. 依元数据自然排序
	sortedAttrs := sortAttributesByPriority(attrs)

	// 3. 由 excelx 引擎渲染包含 3 行表头与下拉数据验证的空白模板
	columns := toExcelColumns(sortedAttrs)
	return excelx.NewWriter(mdl.SheetName(), columns).WriteTemplate()
}

// toExcelColumns 将 CMDB Attribute 领域模型转换为通用的 excelx 声明式列 Schema
func toExcelColumns(attrs []domain.Attribute) []excelx.Column {
	return lo.Map(attrs, func(a domain.Attribute, _ int) excelx.Column {
		return excelx.Column{
			Key:      a.FieldUid,
			Title:    a.FieldName,
			Type:     a.FieldType,
			Required: a.Required,
			Unique:   a.FieldUid == "name",
			Secure:   a.Secure,
			Options:  a.GetOptionStrings(),
			Hidden:   a.FieldType == "file",
		}
	})
}

// sortAttributesByPriority 按模型元数据自然排序字段
func sortAttributesByPriority(attrs []domain.Attribute) []domain.Attribute {
	sorted := make([]domain.Attribute, len(attrs))
	copy(sorted, attrs)

	sort.SliceStable(sorted, func(i, j int) bool {
		// name 资产名称优先排在首列
		if sorted[i].FieldUid == "name" {
			return true
		}
		if sorted[j].FieldUid == "name" {
			return false
		}
		// 依 SortKey 排序
		if sorted[i].SortKey != sorted[j].SortKey {
			return sorted[i].SortKey < sorted[j].SortKey
		}
		// 依 Index 排序
		return sorted[i].Index < sorted[j].Index
	})

	return sorted
}

func (s *dataIOService) fetchModelAndAttributes(ctx context.Context, modelUID string) (domain.Model, []domain.Attribute, error) {
	var (
		mdl   domain.Model
		attrs []domain.Attribute
		eg    errgroup.Group
	)

	// 并行获取 Model 信息和 Attribute 定义
	eg.Go(func() error {
		var err error
		mdl, err = s.modelSvc.GetByUid(ctx, modelUID)
		if err != nil {
			return fmt.Errorf("获取模型信息失败: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		var err error
		var total int64
		attrs, total, err = s.attrSvc.ListAttributes(ctx, modelUID)
		if err != nil {
			return fmt.Errorf("获取模型字段定义失败: %w", err)
		}
		if total == 0 {
			return fmt.Errorf("模型 %s 未定义任何属性字段", modelUID)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return domain.Model{}, nil, err
	}

	return mdl, attrs, nil
}
