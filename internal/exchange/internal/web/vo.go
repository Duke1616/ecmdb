package web

// GenerateUploadURLReq 生成上传 URL 请求
type GenerateUploadURLReq struct {
	FileName string `json:"file_name" binding:"required"`
}

// ExportTemplateReq 导出模板请求
type ExportTemplateReq struct {
	ModelUID string `json:"model_uid" binding:"required"`
}
