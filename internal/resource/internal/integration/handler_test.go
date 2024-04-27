//go:build e2e

package integration

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/attribute"
	attributemocks "github.com/Duke1616/ecmdb/internal/attribute/mocks"
	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"github.com/Duke1616/ecmdb/internal/resource/internal/integration/startup"
	"github.com/Duke1616/ecmdb/internal/resource/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/resource/internal/web"
	"github.com/Duke1616/ecmdb/pkg/ginx/test"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/ecodeclub/ekit/iox"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/mock/gomock"
	"net/http"
	"testing"
	"time"
)

type HandlerTestSuite struct {
	suite.Suite

	dao    dao.ResourceDAO
	db     *mongox.Mongo
	server *gin.Engine
	ctrl   *gomock.Controller
}

func (s *HandlerTestSuite) TearDownSuite() {
	_, err := s.db.Collection("c_resources").DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
	_, err = s.db.Collection("c_id_generator").DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
}

func (s *HandlerTestSuite) TearDownTest() {
	_, err := s.db.Collection("c_resources").DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
	_, err = s.db.Collection("c_id_generator").DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
}

func (s *HandlerTestSuite) SetupSuite() {
	s.ctrl = gomock.NewController(s.T())
	attrSvc := attributemocks.NewMockService(s.ctrl)

	project := []string{"hostname"}
	attrSvc.EXPECT().SearchAttributeFieldsByModelUid(gomock.Any(), "mysql").AnyTimes().Return(project, nil)
	handler, err := startup.InitHandler(&attribute.Module{
		Svc: attrSvc,
	})

	require.NoError(s.T(), err)

	server := gin.Default()

	handler.RegisterRoutes(server)

	s.db = startup.InitMongoDB()
	s.dao = dao.NewResourceDAO(s.db)
	s.server = server
}

func (s *HandlerTestSuite) TestCreate() {
	testCases := []struct {
		name   string
		req    web.CreateResourceReq
		before func(t *testing.T)
		after  func(t *testing.T)

		wantCode int
		wantResp test.Result[int64]
	}{
		{
			name: "创建资源",
			after: func(t *testing.T) {

			},
			before: func(t *testing.T) {

			},
			req: web.CreateResourceReq{
				Name:     "Instance01",
				ModelUid: "mysql",
				Data:     nil,
			},
			wantResp: test.Result[int64]{
				Data: 1,
				Msg:  "创建资源成功",
			},
			wantCode: 200,
		},
		{
			name: "联合唯一索引冲突",
			after: func(t *testing.T) {

			},
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				_, err := s.dao.CreateResource(ctx, dao.Resource{
					Name:     "Instance01",
					ModelUID: "mysql",
				})

				assert.NoError(t, err)
			},
			req: web.CreateResourceReq{
				Name:     "Instance01",
				ModelUid: "mysql",
				Data:     nil,
			},
			wantResp: test.Result[int64]{Code: 503001, Msg: "系统错误"},
			wantCode: 500,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			req, err := http.NewRequest(http.MethodPost,
				"/resource/create", iox.NewJSONReader(tc.req))
			req.Header.Set("content-type", "application/json")
			recorder := test.NewJSONResponseRecorder[int64]()
			s.server.ServeHTTP(recorder, req)
			require.NoError(t, err)
			require.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantResp, recorder.MustScan())

			_, err = s.db.Collection("c_resources").DeleteMany(context.Background(), bson.M{})
			require.NoError(t, err)
			_, err = s.db.Collection("c_id_generator").DeleteMany(context.Background(), bson.M{})
			require.NoError(t, err)
		})
	}
}

func (s *HandlerTestSuite) TestListManyByIds() {
	total := 3
	for idx := 1; idx < total; idx++ {
		rs := dao.Resource{
			Name:     fmt.Sprintf("instance-%d", idx),
			ModelUID: "mysql",
			Data: mongox.MapStr{
				"hostname": fmt.Sprintf("张三-%d", idx),
			},
		}

		_, err := s.dao.CreateResource(context.Background(), rs)
		require.NoError(s.T(), err)
	}

	testCases := []struct {
		name string
		req  web.ListResourceReq

		wantCode int
		wantResp test.Result[[]domain.Resource]
	}{
		{
			name: "查询资产列表",
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost,
				"/resource/list", iox.NewJSONReader(tc.req))
			req.Header.Set("content-type", "application/json")
			recorder := test.NewJSONResponseRecorder[[]domain.Resource]()
			s.server.ServeHTTP(recorder, req)
			require.NoError(t, err)
			require.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantResp.Data, recorder.MustScan().Data)
		})
	}
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
