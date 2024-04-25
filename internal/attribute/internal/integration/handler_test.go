//go:build e2e

package integration

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/integration/startup"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/service"
	"github.com/Duke1616/ecmdb/internal/attribute/internal/web"
	"github.com/Duke1616/ecmdb/pkg/ginx/test"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/ecodeclub/ekit/iox"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"testing"
	"time"
)

type HandlerTestSuite struct {
	suite.Suite

	dao    dao.AttributeDAO
	db     *mongox.Mongo
	server *gin.Engine
	svc    service.Service
}

func (s *HandlerTestSuite) TearDownSuite() {
	_, err := s.db.Collection("c_attribute").DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
	_, err = s.db.Collection("c_id_generator").DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
}

func (s *HandlerTestSuite) TearDownTest() {
	_, err := s.db.Collection("c_attribute").DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
	_, err = s.db.Collection("c_id_generator").DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
}

func (s *HandlerTestSuite) SetupSuite() {
	handler, err := startup.InitHandler()
	require.NoError(s.T(), err)
	server := gin.Default()
	handler.RegisterRoutes(server)

	s.db = startup.InitMongoDB()
	s.dao = dao.NewAttributeDAO(s.db)
	s.server = server
}

func (s *HandlerTestSuite) TestCreate() {
	testCases := []struct {
		name   string
		req    web.CreateAttributeReq
		before func(t *testing.T)

		wantCode int
		wantResp test.Result[int64]
	}{
		{
			name: "创建字段信息",
			req: web.CreateAttributeReq{
				Name:      "访问地址",
				ModelUid:  "host",
				FieldName: "IP",
				FieldType: "string",
				Required:  true,
			},
			wantCode: 200,
			wantResp: test.Result[int64]{
				Data: 1,
				Msg:  "添加模型属性成功",
			},
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost,
				"/attribute/create", iox.NewJSONReader(tc.req))
			req.Header.Set("content-type", "application/json")
			require.NoError(t, err)
			recorder := test.NewJSONResponseRecorder[int64]()
			s.server.ServeHTTP(recorder, req)
			require.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantResp, recorder.MustScan())
		})
	}
}

func (s *HandlerTestSuite) TestList() {
	for i := 1; i <= 2; i++ {
		name := fmt.Sprintf("host%d", i)
		fieldName := fmt.Sprintf("IP%d", i)
		_, err := s.dao.CreateAttribute(context.Background(), dao.Attribute{
			Name:      name,
			ModelUID:  "host",
			FieldName: fieldName,
			FieldType: "string",
			Required:  true,
			Ctime:     time.Unix(0, 0).UnixMilli(),
			Utime:     time.Unix(0, 0).UnixMilli(),
		})
		if err != nil {
			return
		}

		require.NoError(s.T(), err)
	}

	testCases := []struct {
		name   string
		req    web.ListAttributeReq
		before func(t *testing.T)

		wantCode int
		wantResp test.Result[web.RetrieveAttributeList]
	}{
		{
			name: "查询字段信息",
			req: web.ListAttributeReq{
				ModelUid: "host",
			},
			wantCode: 200,
			wantResp: test.Result[web.RetrieveAttributeList]{
				Data: web.RetrieveAttributeList{
					Attribute: []web.Attribute{
						{
							ID:        1,
							ModelUID:  "host",
							Name:      "host1",
							FieldName: "IP1",
							FieldType: "string",
							Required:  true,
						},
						{
							ID:        2,
							ModelUID:  "host",
							Name:      "host2",
							FieldName: "IP2",
							FieldType: "string",
							Required:  true,
						},
					},
					Total: 2,
				},
			},
		},
		{
			name: "查询不存在的模型",
			req: web.ListAttributeReq{
				ModelUid: "mysql",
			},
			wantCode: 200,
			wantResp: test.Result[web.RetrieveAttributeList]{
				Data: web.RetrieveAttributeList{
					Attribute: nil,
					Total:     0,
				},
			},
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost,
				"/attribute/list", iox.NewJSONReader(tc.req))
			req.Header.Set("content-type", "application/json")
			require.NoError(t, err)
			recorder := test.NewJSONResponseRecorder[web.RetrieveAttributeList]()
			s.server.ServeHTTP(recorder, req)
			require.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantResp, recorder.MustScan())
		})
	}
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
