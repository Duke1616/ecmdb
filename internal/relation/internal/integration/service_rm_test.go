package integration

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/integration/startup"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/relation/internal/service"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

type ServiceRMTestSuite struct {
	suite.Suite

	dao dao.RelationModelDAO
	db  *mongox.Mongo
	svc service.RelationModelService
}

func (s *ServiceRMTestSuite) SetupTest() {
	s.svc = startup.InitRMSvc()
	db := startup.InitMongoDB()
	s.dao = dao.NewRelationModelDAO(db)
	s.db = db
}

func (s *ServiceRMTestSuite) TearDownSuite() {
	_, err := s.db.Collection(dao.ModelRelationCollection).DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
	_, err = s.db.Collection("c_id_generator").DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
}

func (s *ServiceRMTestSuite) TearDownTest() {
	_, err := s.db.Collection(dao.ModelRelationCollection).DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
	_, err = s.db.Collection("c_id_generator").DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
}

func (s *ServiceRMTestSuite) TestFindModelDiagramBySrcUids() {
	testCase := []struct {
		name string
		req  []string

		before   func(t *testing.T)
		after    func(t *testing.T)
		wantResp []domain.ModelDiagram
	}{
		{
			name: "查询模型关联关系绘制拓扑图",
			req:  []string{"mysql", "host"},
			before: func(t *testing.T) {
				_, err := s.dao.CreateModelRelation(context.Background(), dao.ModelRelation{
					SourceModelUid:  "mysql",
					TargetModelUid:  "mongo",
					RelationTypeUid: "run",
				})
				require.NoError(t, err)
				_, err = s.dao.CreateModelRelation(context.Background(), dao.ModelRelation{
					SourceModelUid:  "host",
					TargetModelUid:  "mongo",
					RelationTypeUid: "run",
				})
				require.NoError(t, err)
				_, err = s.dao.CreateModelRelation(context.Background(), dao.ModelRelation{
					SourceModelUid:  "mongo",
					TargetModelUid:  "mongo",
					RelationTypeUid: "run",
				})
				require.NoError(t, err)
			},
			after: func(t *testing.T) {

			},
			wantResp: []domain.ModelDiagram{
				{
					ID:              2,
					SourceModelUid:  "host",
					TargetModelUid:  "mongo",
					RelationTypeUid: "run",
				},
				{
					ID:              1,
					SourceModelUid:  "mysql",
					TargetModelUid:  "mongo",
					RelationTypeUid: "run",
				},
			},
		},
	}

	for _, tc := range testCase {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			mds, err := s.svc.FindModelDiagramBySrcUids(context.Background(), tc.req)
			if err != nil {
				return
			}

			tc.after(t)
			assert.Equal(t, mds, tc.wantResp)
		})
	}
}

func TestRMService(t *testing.T) {
	suite.Run(t, new(ServiceRMTestSuite))
}
