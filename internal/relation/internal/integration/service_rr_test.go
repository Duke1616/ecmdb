package integration

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/relation/internal/integration/startup"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/relation/internal/service"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

type ServiceRRTestSuite struct {
	suite.Suite

	dao dao.RelationResourceDAO
	db  *mongox.Mongo
	svc service.RelationResourceService
}

func (s *ServiceRRTestSuite) SetupTest() {
	s.svc = startup.InitRRSvc()
	db := startup.InitMongoDB()
	s.dao = dao.NewRelationResourceDAO(db)
	s.db = db
}

func (s *ServiceRRTestSuite) TearDownSuite() {
	_, err := s.db.Collection(dao.ResourceRelationCollection).DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
	_, err = s.db.Collection("c_id_generator").DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
}

func (s *ServiceRRTestSuite) TearDownTest() {
	_, err := s.db.Collection(dao.ResourceRelationCollection).DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
	_, err = s.db.Collection("c_id_generator").DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
}

func (s *ServiceRRTestSuite) TestCreate() {
	testCase := []struct {
		name string
	}{
		{
			name: "创建成功",
		},
	}

	for _, tc := range testCase {
		s.T().Run(tc.name, func(t *testing.T) {

		})
	}
}

func TestRRService(t *testing.T) {
	suite.Run(t, new(ServiceRRTestSuite))
}
