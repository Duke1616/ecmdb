package web

import (
	"github.com/Duke1616/ecmdb/internal/exchange/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/Duke1616/ecmdb/pkg/storage"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc     service.IExchangeService
	storage *storage.S3Storage
}

func NewHandler(svc service.IExchangeService, storage *storage.S3Storage) *Handler {
	return &Handler{
		svc:     svc,
		storage: storage,
	}
}

func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/exchange")
	// 生成预签名上传 URL
	g.POST("/presigned_upload", ginx.WrapBody[GenerateUploadURLReq](h.GenerateUploadURL))
	// 导出模板
	g.GET("/template/export/:model_uid", ginx.Wrap(h.ExportTemplate))
	// 导入数据 (S3 模式)
	g.POST("/import", ginx.WrapBody[ImportReq](h.Import))
	// 导入数据 V2 (直接上传文件)
	g.POST("/import/v2", ginx.Wrap(h.ImportV2))
}

// GenerateUploadURL 生成预签名上传 URL
// NOTE: 前端获取上传 URL 后,使用 PUT 方法直接上传到 S3,不经过后端服务器
func (h *Handler) GenerateUploadURL(ctx *gin.Context, req GenerateUploadURLReq) (ginx.Result, error) {
	// 生成预签名上传 URL(默认 15 分钟有效期)
	fileKey, uploadURL, err := h.storage.GenerateUploadURL(ctx.Request.Context(), req.FileName, 900)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg: "生成上传链接成功",
		Data: map[string]interface{}{
			"file_key":   fileKey,
			"upload_url": uploadURL,
			"method":     "PUT",
			"expires_in": 900,
		},
	}, nil
}

// ExportTemplate 导出空白导入模板
func (h *Handler) ExportTemplate(ctx *gin.Context) (ginx.Result, error) {
	// 根据请求获取模型UID
	modelUid := ctx.Param("model_uid")

	// 调用 Service 生成 Excel 模板
	excelData, err := h.svc.ExportTemplate(ctx.Request.Context(), modelUid)
	if err != nil {
		return systemErrorResult, err
	}

	// 设置 HTTP 响应头,直接返回 Excel 文件
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", "attachment; filename="+modelUid+"_template.xlsx")
	ctx.Header("Content-Transfer-Encoding", "binary")

	// 直接写入 Excel 数据
	ctx.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", excelData)

	// NOTE: 返回空 Result,因为已经通过 ctx.Data 直接发送了响应
	return ginx.Result{}, nil
}

// Import 导入数据
// NOTE: 前端先通过 GenerateUploadURL 上传文件到 S3,然后调用此接口传入 file_key 进行导入
func (h *Handler) Import(ctx *gin.Context, req ImportReq) (ginx.Result, error) {
	// 1. 从 S3 下载文件
	fileData, err := h.storage.GetFile(ctx.Request.Context(), req.FileKey)
	if err != nil {
		return systemErrorResult, err
	}

	// 2. 调用 Service 导入数据
	importedCount, err := h.svc.Import(ctx.Request.Context(), req.ModelUID, fileData)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg: "导入成功",
		Data: map[string]interface{}{
			"imported_count": importedCount,
		},
	}, nil
}
