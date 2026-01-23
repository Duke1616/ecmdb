package service

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/resource"
)

// IDataIOService 数据交换服务接口
// NOTE: 提供基于 Model-Attribute-Resource 架构的数据导入导出功能,支持 Excel 格式
type IDataIOService interface {
	// Import 批量导入资源实例 (Resource)
	// modelUID: 模型唯一标识 (对应 Model.UID)
	// fileData: Excel 文件的字节数据
	Import(ctx context.Context, modelUID string, fileData []byte) (importedCount int, err error)

	// Export 导出资源实例数据 (Resource)
	// req: 导出请求参数
	Export(ctx context.Context, req ExportParams) ([]byte, error)

	// ExportTemplate 导出模板
	// modelUID: 模型唯一标识 (对应 Model.UID)
	ExportTemplate(ctx context.Context, modelUID string) ([]byte, error)
}

type ExportParams struct {
	ModelUID     string
	Scope        string // "all", "current", "selected"
	ResourceIDs  []int64
	FilterGroups []resource.FilterGroup
	Fields       []string
	FileName     string
}
