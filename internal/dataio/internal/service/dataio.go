package service

import (
	"bytes"
	"context"
	"fmt"
	"sort"

	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/dataio/internal/domain"
	"github.com/Duke1616/ecmdb/internal/model"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/xuri/excelize/v2"
	"golang.org/x/sync/errgroup"
)

// fieldPriority 字段优先级配置
// NOTE: 用于控制 Excel 导出时的列顺序,数字越小越靠前,未配置的字段默认为 999
var fieldPriority = map[string]int{
	"name": 1, // name 字段始终在第一列
	// 可以继续添加其他需要固定顺序的字段
	// "ip":   2,
	// "port": 3,
}

// sortAttributesByPriority 按优先级排序字段
func sortAttributesByPriority(attrs []attribute.Attribute) []attribute.Attribute {
	sorted := make([]attribute.Attribute, len(attrs))
	copy(sorted, attrs)

	sort.SliceStable(sorted, func(i, j int) bool {
		pi := fieldPriority[sorted[i].FieldUid]
		pj := fieldPriority[sorted[j].FieldUid]
		if pi == 0 {
			pi = 999 // 未配置的字段默认优先级
		}
		if pj == 0 {
			pj = 999
		}
		return pi < pj
	})

	return sorted
}

// NOTE: dataIOService 实现数据交换功能,依赖三个模块的 Service
type dataIOService struct {
	attrSvc  attribute.Service
	resSvc   resource.Service
	modelSvc model.Service
}

// NewDataIOService 创建数据交换服务实例
func NewDataIOService(
	attrSvc attribute.Service,
	resSvc resource.Service,
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
	// 1. 解析 Excel 文件
	f, err := excelize.OpenReader(bytes.NewReader(fileData))
	if err != nil {
		return 0, fmt.Errorf("解析 Excel 文件失败: %w", err)
	}
	defer f.Close()

	// 2. 获取 Attribute 定义
	attrs, _, err := s.attrSvc.ListAttributes(ctx, modelUID)
	if err != nil {
		return 0, fmt.Errorf("获取模型字段定义失败: %w", err)
	}
	if len(attrs) == 0 {
		return 0, fmt.Errorf("模型 %s 没有定义字段", modelUID)
	}

	// 3. 读取第一个 sheet 的数据
	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return 0, fmt.Errorf("读取 Excel 数据失败: %w", err)
	}
	if len(rows) < 4 {
		return 0, fmt.Errorf("Excel 文件格式错误,至少需要 3 行表头 + 1 行数据")
	}

	// 4. 解析第二行表头(FieldUid),建立列索引映射
	fieldUidRow := rows[1]              // 第二行是 FieldUid
	colIndexMap := make(map[int]string) // 列索引 → FieldUid

	for colIdx, fieldUid := range fieldUidRow {
		if fieldUid != "" {
			colIndexMap[colIdx] = fieldUid
		}
	}

	// 5. 逐行读取并构建 Resource (从第 4 行开始,跳过 3 行表头)
	resources := make([]resource.Resource, 0, len(rows)-3)
	for _, row := range rows[3:] {
		// 构建 Resource Data
		data := make(map[string]interface{})
		for colIdx, cellValue := range row {
			if fieldUid, ok := colIndexMap[colIdx]; ok {
				// 跳过空值
				if cellValue == "" {
					continue
				}
				data[fieldUid] = cellValue
			}
		}

		// 跳过空行
		if len(data) == 0 {
			continue
		}

		resources = append(resources, resource.Resource{
			ModelUID: modelUID,
			Data:     data,
		})
	}

	if len(resources) == 0 {
		return 0, fmt.Errorf("没有有效的数据行")
	}

	// 6. 批量创建或更新 Resource
	err = s.resSvc.BatchCreateOrUpdate(ctx, resources)
	if err != nil {
		return 0, fmt.Errorf("批量创建或更新资源失败: %w", err)
	}

	return len(resources), nil
}

// Export 导出资源实例数据 (Resource)
func (s *dataIOService) Export(ctx context.Context, modelUID string, filter interface{}) ([]byte, error) {
	return nil, nil
}

// ExportTemplate 导出空白导入模板
func (s *dataIOService) ExportTemplate(ctx context.Context, modelUID string) ([]byte, error) {
	// 1. 获取数据
	mdl, attrs, err := s.fetchModelAndAttributes(ctx, modelUID)
	if err != nil {
		return nil, err
	}

	// 2. 按优先级排序字段(name 字段会自动排在第一列)
	sortedAttrs := sortAttributesByPriority(attrs)

	// 3. 构建 3 行表头数据
	row1 := make([]string, len(sortedAttrs)) // 字段约束
	row2 := make([]string, len(sortedAttrs)) // 字段 UID
	row3 := make([]string, len(sortedAttrs)) // 字段名称

	for i, attr := range sortedAttrs {
		row1[i] = attr.GetConstraintDescription()
		row2[i] = attr.FieldUid
		row3[i] = attr.FieldName
	}

	// 5. 构建 Excel
	builder := domain.NewBuilder(mdl.SheetName()).
		With3RowHeaders(row1, row2, row3)
	defer builder.Close()

	// 6. 添加数据验证(下拉列表)
	for colIdx, attr := range sortedAttrs {
		if attr.NeedsValidation() {
			builder.WithValidation(colIdx, attr.GetOptionStrings(), 4, 1000)
		}
	}

	// 7. 导出
	return builder.ToBytes()
}

func (s *dataIOService) fetchModelAndAttributes(ctx context.Context, modelUID string) (model.Model, []attribute.Attribute, error) {
	var (
		mdl   model.Model
		attrs []attribute.Attribute
		eg    errgroup.Group
	)

	// 1. 并行获取 Model 信息和 Attribute 定义
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
			return fmt.Errorf("模型 %s 没有定义字段", modelUID)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return model.Model{}, nil, err
	}

	return mdl, attrs, nil
}
