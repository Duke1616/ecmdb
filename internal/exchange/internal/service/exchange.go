package service

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/model"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/Duke1616/ecmdb/pkg/excel"
)

// NOTE: exchangeService 实现数据交换功能,依赖三个模块的 Service
type exchangeService struct {
	attrSvc  attribute.Service
	resSvc   resource.Service
	modelSvc model.Service
}

// NewExchangeService 创建数据交换服务实例
func NewExchangeService(
	attrSvc attribute.Service,
	resSvc resource.Service,
	modelSvc model.Service,

) IExchangeService {
	return &exchangeService{
		attrSvc:  attrSvc,
		resSvc:   resSvc,
		modelSvc: modelSvc,
	}
}

//func (s *exchangeService) RunTask(ctx context.Context, url string) error {
//	_, err := s.taskClient.CreateTask(ctx, &taskv1.CreateTaskRequest{
//		Name:     "",
//		CronExpr: "0 0 * * * ?",
//		Type:     taskv1.TaskType_ONE_TIME,
//		GrpcConfig: &taskv1.GrpcConfig{
//			ServiceName: "",
//			AuthToken:   "",
//			HandlerName: "",
//			Params:      nil,
//		},
//		ScheduleParams: map[string]string{
//			"123": "123",
//		},
//	})
//
//	return err
//}

// ImportData 批量导入资源实例 (Resource)
func (s *exchangeService) ImportData(ctx context.Context, modelUID string, fileData []byte, validate bool) (importedCount int, err error) {
	return 0, nil
}

// ExportData 导出资源实例数据 (Resource)
func (s *exchangeService) ExportData(ctx context.Context, modelUID string, filter interface{}) ([]byte, error) {
	return nil, nil
}

// ExportTemplate 导出空白导入模板
func (s *exchangeService) ExportTemplate(ctx context.Context, modelUID string) ([]byte, error) {
	// 1. 获取 Attribute 定义
	attrs, _, err := s.attrSvc.ListAttributes(ctx, modelUID)
	if err != nil {
		return nil, fmt.Errorf("获取模型字段定义失败: %w", err)
	}
	if len(attrs) == 0 {
		return nil, fmt.Errorf("模型 %s 没有定义字段", modelUID)
	}

	// 2. 构建 3 行表头数据
	row1 := make([]string, len(attrs)) // 字段约束
	row2 := make([]string, len(attrs)) // 字段 UID
	row3 := make([]string, len(attrs)) // 字段名称
	for i, attr := range attrs {
		row1[i] = attr.GetConstraintDescription()
		row2[i] = attr.FieldUid
		row3[i] = attr.FieldName
	}

	// 3. 构建 Excel
	builder := excel.NewBuilder("Sheet1").
		With3RowHeaders(row1, row2, row3)
	defer builder.Close()

	// 4. 添加数据验证(下拉列表)
	for colIdx, attr := range attrs {
		if attr.NeedsValidation() {
			builder.WithValidation(colIdx, attr.GetOptionStrings(), 4, 1000)
		}
	}

	// 5. 导出
	return builder.ToBytes()
}
