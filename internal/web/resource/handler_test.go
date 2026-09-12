package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Duke1616/ecmdb/internal/domain"
	resourcemocks "github.com/Duke1616/ecmdb/internal/service/resource/mocks"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/ecodeclub/ginx"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHandler_AdminSearchStructure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name       string
		mock       func(ctrl *gomock.Controller) *resourcemocks.MockService
		req        AdminSearchStructureReq
		wantErrMsg string
		validate   func(t *testing.T, res ginx.Result)
	}{
		{
			name: "查询大盘结构服务异常",
			mock: func(ctrl *gomock.Controller) *resourcemocks.MockService {
				svc := resourcemocks.NewMockService(ctrl)
				svc.EXPECT().AdminSearchStructure(gomock.Any(), "err").
					Return(domain.AdminSearchStructureResult{}, errors.New("db error"))
				return svc
			},
			req:        AdminSearchStructureReq{Text: "err"},
			wantErrMsg: "db error",
		},
		{
			name: "获取大盘结构成功并正确转换VO",
			mock: func(ctrl *gomock.Controller) *resourcemocks.MockService {
				svc := resourcemocks.NewMockService(ctrl)
				svc.EXPECT().AdminSearchStructure(gomock.Any(), "prod").
					Return(domain.AdminSearchStructureResult{
						Total: 10,
						Organizations: []domain.AdminSearchStructureTenant{
							{
								TenantID:   1,
								TenantType: "system",
								Total:      8,
								Models: []domain.AdminSearchStructureModel{
									{ModelUID: "host", Total: 8},
								},
							},
						},
						Personals: []domain.AdminSearchStructureTenant{},
					}, nil)
				return svc
			},
			req: AdminSearchStructureReq{Text: "prod"},
			validate: func(t *testing.T, res ginx.Result) {
				data, ok := res.Data.(RetrieveAdminSearchStructure)
				assert.True(t, ok)
				assert.Equal(t, 10, data.Total)
				assert.Len(t, data.Organizations, 1)
				assert.Equal(t, int64(1), data.Organizations[0].TenantID)
				assert.Equal(t, "host", data.Organizations[0].Models[0].ModelUID)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := tc.mock(ctrl)
			h := &Handler{svc: svc}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req, _ := http.NewRequest(http.MethodPost, "/admin/search/structure", nil)
			c.Request = req

			res, err := h.AdminSearchStructure(&ginx.Context{Context: c}, tc.req)
			if tc.wantErrMsg != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErrMsg, err.Error())
			} else {
				assert.NoError(t, err)
				if tc.validate != nil {
					tc.validate(t, res)
				}
			}
		})
	}
}



func TestHandler_SearchStructure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name       string
		mock       func(ctrl *gomock.Controller) *resourcemocks.MockService
		req        SearchStructureReq
		wantErrMsg string
		validate   func(t *testing.T, res ginx.Result)
	}{
		{
			name: "单租户检索模型概览成功",
			mock: func(ctrl *gomock.Controller) *resourcemocks.MockService {
				svc := resourcemocks.NewMockService(ctrl)
				svc.EXPECT().SearchStructure(gomock.Any(), "prod").
					Return(domain.SearchStructureResult{
						Total: 5,
						Models: []domain.AdminSearchStructureModel{
							{ModelUID: "host", Total: 5},
						},
					}, nil)
				return svc
			},
			req: SearchStructureReq{Text: "prod"},
			validate: func(t *testing.T, res ginx.Result) {
				data, ok := res.Data.(RetrieveSearchStructure)
				assert.True(t, ok)
				assert.Equal(t, 5, data.Total)
				assert.Len(t, data.Models, 1)
				assert.Equal(t, "host", data.Models[0].ModelUID)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := tc.mock(ctrl)
			h := &Handler{svc: svc}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req, _ := http.NewRequest(http.MethodPost, "/search/structure", nil)
			c.Request = req

			res, err := h.SearchStructure(&ginx.Context{Context: c}, tc.req)
			if tc.wantErrMsg != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErrMsg, err.Error())
			} else {
				assert.NoError(t, err)
				if tc.validate != nil {
					tc.validate(t, res)
				}
			}
		})
	}
}

func TestHandler_SearchPagedResources(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name       string
		mock       func(ctrl *gomock.Controller) *resourcemocks.MockService
		req        SearchPagedResourcesReq
		wantErrMsg string
		validate   func(t *testing.T, res ginx.Result)
	}{
		{
			name: "单租户默认上下文分页检索明细成功",
			mock: func(ctrl *gomock.Controller) *resourcemocks.MockService {
				svc := resourcemocks.NewMockService(ctrl)
				svc.EXPECT().SearchPagedResources(gomock.Any(), int64(0), "host", "web", int64(0), int64(15)).
					Return([]domain.Resource{
						{ID: 101, Name: "web-01", ModelUID: "host", Data: mongox.MapStr{"ip": "10.0.0.1"}},
					}, int64(1), nil)
				return svc
			},
			req: SearchPagedResourcesReq{
				Page:     Page{Offset: 0, Limit: 15},
				TenantID: 0,
				ModelUID: "host",
				Text:     "web",
			},
			validate: func(t *testing.T, res ginx.Result) {
				data, ok := res.Data.(RetrieveResources)
				assert.True(t, ok)
				assert.Equal(t, int64(1), data.Total)
				assert.Len(t, data.Resources, 1)
				assert.Equal(t, "web-01", data.Resources[0].Name)
			},
		},
		{
			name: "跨租户大盘指定租户分页检索明细成功",
			mock: func(ctrl *gomock.Controller) *resourcemocks.MockService {
				svc := resourcemocks.NewMockService(ctrl)
				svc.EXPECT().SearchPagedResources(gomock.Any(), int64(2), "host", "web", int64(0), int64(15)).
					Return([]domain.Resource{
						{ID: 102, Name: "web-02", ModelUID: "host", Data: mongox.MapStr{"ip": "10.0.0.2"}},
					}, int64(1), nil)
				return svc
			},
			req: SearchPagedResourcesReq{
				Page:     Page{Offset: 0, Limit: 15},
				TenantID: 2,
				ModelUID: "host",
				Text:     "web",
			},
			validate: func(t *testing.T, res ginx.Result) {
				data, ok := res.Data.(RetrieveResources)
				assert.True(t, ok)
				assert.Equal(t, int64(1), data.Total)
				assert.Len(t, data.Resources, 1)
				assert.Equal(t, "web-02", data.Resources[0].Name)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := tc.mock(ctrl)
			h := &Handler{svc: svc}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req, _ := http.NewRequest(http.MethodPost, "/search/resources", nil)
			c.Request = req

			res, err := h.SearchPagedResources(&ginx.Context{Context: c}, tc.req)
			if tc.wantErrMsg != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErrMsg, err.Error())
			} else {
				assert.NoError(t, err)
				if tc.validate != nil {
					tc.validate(t, res)
				}
			}
		})
	}
}

