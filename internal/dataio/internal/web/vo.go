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
