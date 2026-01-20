package service

import "context"

// IExchangeService 数据交换服务接口
// NOTE: 提供基于 Model-Attribute-Resource 架构的数据导入导出功能,支持 Excel 格式
type IExchangeService interface {
	// ImportData 批量导入资源实例 (Resource)
	// modelUID: 模型唯一标识 (对应 Model.UID)
	// fileData: Excel 文件的字节数据
	// validate: 是否根据 Attribute 定义进行严格校验(必填字段、下拉选项、数据类型等)
	// NOTE: 导入过程会解析 Excel 数据并创建 Resource 实例,失败时返回详细的错误信息
	ImportData(ctx context.Context, modelUID string, fileData []byte, validate bool) (importedCount int, err error)

	// ExportData 导出资源实例数据 (Resource)
	// modelUID: 模型唯一标识 (对应 Model.UID)
	// filter: 资源过滤条件,用于筛选需要导出的 Resource
	// NOTE: 导出的 Excel 包含基于 Attribute 定义的字段结构和实际 Resource 数据,带格式和下拉校验
	ExportData(ctx context.Context, modelUID string, filter interface{}) ([]byte, error)

	// ExportTemplate 导出空白导入模板
	// modelUID: 模型唯一标识 (对应 Model.UID)
	// NOTE: 基于 Attribute 定义生成带字段格式和下拉校验的空白表格,用于用户填写后导入
	ExportTemplate(ctx context.Context, modelUID string) ([]byte, error)
}
