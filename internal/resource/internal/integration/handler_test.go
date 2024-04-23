//go:build e2e

package integration

import (
	"github.com/Duke1616/ecmdb/internal/resource/internal/integration/startup"
	"github.com/Duke1616/ecmdb/internal/resource/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/resource/internal/web"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/ecodeclub/ekit/iox"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

type HandlerTestSuite struct {
	suite.Suite

	dao    dao.ResourceDAO
	db     *mongox.Mongo
	server *gin.Engine
	ctrl   *gomock.Controller
}

func (s *HandlerTestSuite) SetupSuite() {
	s.ctrl = gomock.NewController(s.T())
	handler, err := startup.InitHandler()
	require.NoError(s.T(), err)

	server := gin.Default()

	handler.RegisterRoutes(server)

	s.db = startup.InitMongoDB()
	s.dao = dao.NewResourceDAO(s.db)
	s.server = server
}

func (s *HandlerTestSuite) TestCreate() {
	testCases := []struct {
		name string
		req  web.CreateResourceReq

		wantCode int
	}{
		{
			name:     "创建资源",
			wantCode: 200,
			req: web.CreateResourceReq{
				Name:     "Instance01",
				ModelUid: "mysql",
				Data:     nil,
			},
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost,
				"/resource/create", iox.NewJSONReader(tc.req))
			req.Header.Set("content-type", "application/json")
			recorder := httptest.NewRecorder()
			s.server.ServeHTTP(recorder, req)
			require.NoError(t, err)
		})
	}
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
