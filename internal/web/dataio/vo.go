package web

// GenerateUploadURLReq 生成上传 URL 请求
type GenerateUploadURLReq struct {
	FileName string `json:"file_name" binding:"required"`
}

// ExportTemplateReq 导出模板请求
type ExportTemplateReq struct {
	ModelUID string `json:"model_uid" binding:"required"`
}

// ImportReq 导入数据请求
type ImportReq struct {
	ModelUID string `json:"model_uid" binding:"required"` // 模型 UID
	FileKey  string `json:"file_key" binding:"required"`  // S3 文件 key
}

// ImportV2Req 导入数据请求 V2 (直接上传文件)
type ImportV2Req struct {
	ModelUID string `form:"model_uid" binding:"required"` // 模型 UID
}

// ExportScope 导出范围枚举
type ExportScope string

const (
	ExportScopeAll      ExportScope = "all"
	ExportScopeCurrent  ExportScope = "current"
	ExportScopeSelected ExportScope = "selected"
)

func (s ExportScope) String() string {
	return string(s)
}

// ExportFilterCondition 导出筛选条件
type ExportFilterCondition struct {
	FieldUID string      `json:"field_uid"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// ExportFilterGroup 导出筛选条件组 (组内 AND)
type ExportFilterGroup struct {
	Filters []ExportFilterCondition `json:"filters"`
}

// ExportReq 导出数据请求
type ExportReq struct {
	ModelUID     string              `json:"model_uid" binding:"required"`
	Scope        ExportScope         `json:"scope" binding:"required"`
	ResourceIDs  []int64             `json:"resource_ids"`  // string or number
	FilterGroups []ExportFilterGroup `json:"filter_groups"` // scope='all' 或 'current' 时可选
	Fields       []string            `json:"fields"`        // 导出字段列表 (可选)
	FileName     string              `json:"file_name"`     // 文件名 (可选)
}
