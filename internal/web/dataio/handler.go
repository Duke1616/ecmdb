package web

import (
	"github.com/Duke1616/ecmdb/internal/domain"
	service "github.com/Duke1616/ecmdb/internal/service/dataio"
	"github.com/Duke1616/ecmdb/pkg/storage"
	"github.com/Duke1616/eiam/pkg/web/capability"
	"github.com/ecodeclub/ginx"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

var _ ginx.Handler = &Handler{}

type Handler struct {
	svc     service.IDataIOService
	storage *storage.S3Storage
	capability.IRegistry
}

func NewHandler(svc service.IDataIOService, storage *storage.S3Storage) *Handler {
	return &Handler{
		svc:       svc,
		storage:   storage,
		IRegistry: capability.NewRegistry("cmdb", "dataio", "资产仓库/导入导出"),
	}
}

func (h *Handler) PublicRoutes(_ *gin.Engine) {}

func (h *Handler) IdentifyRoutes(_ *gin.Engine) {}

func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/dataio")
	// 导出模板
	g.GET("/template/export/:model_uid", h.Define("模板导出", "export_template").
		Bind(ginx.W(h.ExportTemplate)),
	)
	// 导入数据 (S3 模式)
	g.POST("/import", h.Define("数据导入", "import").
		Bind(ginx.B[ImportReq](h.Import)),
	)
	// 导出数据
	g.POST("/export", h.Define("数据导出", "export").
		Bind(ginx.B[ExportReq](h.Export)),
	)
}

// Export 导出数据
func (h *Handler) Export(ctx *ginx.Context, req ExportReq) (ginx.Result, error) {
	// 转换 FilterGroups
	groups := lo.Map(req.FilterGroups, func(src ExportFilterGroup, _ int) domain.FilterGroup {
		return domain.FilterGroup{
			Filters: lo.Map(src.Filters, func(f ExportFilterCondition, _ int) domain.FilterCondition {
				return domain.FilterCondition{
					FieldUID: f.FieldUID,
					Operator: domain.Operator(f.Operator),
					Value:    f.Value,
				}
			}),
		}
	})


	params := service.ExportParams{
		ModelUID:     req.ModelUID,
		Scope:        req.Scope.String(),
		ResourceIDs:  req.ResourceIDs,
		FilterGroups: groups,
		Fields:       req.Fields,
		FileName:     req.FileName,
	}

	// 调用 Service 导出数据
	excelData, err := h.svc.Export(ctx.Context, params)
	if err != nil {
		return systemErrorResult, err
	}

	fileName := req.FileName
	if fileName == "" {
		fileName = req.ModelUID + "_export.xlsx"
	}
	// 确保后缀
	if len(fileName) < 5 || fileName[len(fileName)-5:] != ".xlsx" {
		fileName += ".xlsx"
	}

	// 设置 HTTP 响应头
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", "attachment; filename="+fileName)
	ctx.Header("Content-Transfer-Encoding", "binary")

	// 直接写入 Excel 数据
	ctx.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", excelData)

	return ginx.Result{}, nil
}

// ExportTemplate 导出空白导入模板
func (h *Handler) ExportTemplate(ctx *ginx.Context) (ginx.Result, error) {
	// 根据请求获取模型UID
	modelUid, err := ctx.Param("model_uid").AsString()
	if err != nil {
		return systemErrorResult, err
	}

	// 调用 Service 生成 Excel 模板
	excelData, err := h.svc.ExportTemplate(ctx.Context, modelUid)
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
func (h *Handler) Import(ctx *ginx.Context, req ImportReq) (ginx.Result, error) {
	// 1. 从 S3 下载文件
	fileData, err := h.storage.GetFile(ctx.Context, "ecmdb", req.FileKey)
	if err != nil {
		return systemErrorResult, err
	}

	// 2. 调用 Service 导入数据
	importedCount, err := h.svc.Import(ctx.Context, req.ModelUID, fileData)
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

