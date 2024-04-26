//go:build e2e

package integration

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/model/internal/integration/startup"
	"github.com/Duke1616/ecmdb/internal/model/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/model/internal/service"
	"github.com/Duke1616/ecmdb/internal/model/internal/web"
	"github.com/Duke1616/ecmdb/pkg/ginx/test"
	"github.com/ecodeclub/ekit/iox"
	"github.com/stretchr/testify/assert"
	"net/http"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

type HandlerTestSuite struct {
	suite.Suite

	dao    dao.ModelDAO
	db     *mongox.Mongo
	server *gin.Engine
	svc    service.Service
}

func (s *HandlerTestSuite) TearDownSuite() {
	_, err := s.db.Collection(dao.ModelCollection).DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
	_, err = s.db.Collection("c_id_generator").DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
}

func (s *HandlerTestSuite) TearDownTest() {
	_, err := s.db.Collection(dao.ModelCollection).DeleteMany(context.Background(), bson.M{})
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
	s.dao = dao.NewModelDAO(s.db)
	s.server = server
}

func (s *HandlerTestSuite) TestCreate() {
	testCase := []struct {
		name   string
		req    web.CreateModelReq
		before func(t *testing.T)

		wantCode int
		wantResp test.Result[int64]
	}{
		{
			name: "创建模型",
			req: web.CreateModelReq{
				Name:    "host",
				GroupId: 1,
				UID:     "host",
				Icon:    "icon-host",
			},
			wantCode: 200,
			wantResp: test.Result[int64]{
				Data: 1,
				Msg:  "添加模型成功",
			},
		},
	}

	for _, tc := range testCase {
		s.T().Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost,
				"/model/create", iox.NewJSONReader(tc.req))
			req.Header.Set("content-type", "application/json")
			require.NoError(t, err)
			recorder := test.NewJSONResponseRecorder[int64]()
			s.server.ServeHTTP(recorder, req)
			require.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantResp, recorder.MustScan())
		})
	}
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
