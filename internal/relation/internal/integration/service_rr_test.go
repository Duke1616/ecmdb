package integration

import (
	"context"
	"testing"

	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/integration/startup"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/relation/internal/service"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
)

type ServiceRRTestSuite struct {
	suite.Suite

	dao dao.RelationResourceDAO
	db  *mongox.DB
	svc service.RelationResourceService
}

func (s *ServiceRRTestSuite) SetupTest() {
	s.svc = startup.InitRRSvc()
	db := startup.InitMongoDB()
	s.dao = dao.NewRelationResourceDAO(db)
	s.db = db
}

func (s *ServiceRRTestSuite) TearDownSuite() {
	_, err := s.db.Database().Collection(dao.ResourceRelationCollection).DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
	_, err = s.db.Database().Collection("c_relation_model").DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
	_, err = s.db.Database().Collection("c_id_generator").DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
}

func (s *ServiceRRTestSuite) TearDownTest() {
	_, err := s.db.Database().Collection(dao.ResourceRelationCollection).DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
	_, err = s.db.Database().Collection("c_relation_model").DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
	_, err = s.db.Database().Collection("c_id_generator").DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
}

func (s *ServiceRRTestSuite) TestCreate() {
	testCases := []struct {
		name      string
		before    func(t *testing.T)
		req       domain.ResourceRelation
		wantErr   string
		wantName  string
		wantModel string
	}{
		{
			name: "未注册模型关联-创建失败",
			before: func(t *testing.T) {
				// 什么都不干，此时元数据库为空
			},
			req: domain.ResourceRelation{
				SourceModelUID:   "server",
				TargetModelUID:   "ip",
				SourceResourceID: 101,
				TargetResourceID: 202,
				RelationTypeUID:  "bind",
			},
			wantErr: "拓扑关联校验失败：模型关系 server -> ip (关系类型: bind) 未注册定义",
		},
		{
			name: "已注册模型关联-创建成功并补齐字段",
			before: func(t *testing.T) {
				// 往元数据库插入 server -> ip 的 bind 关系定义
				_, err := s.db.Database().Collection("c_relation_model").InsertOne(context.Background(), bson.M{
					"id":                int64(999),
					"source_model_uid":  "server",
					"target_model_uid":  "ip",
					"relation_type_uid": "bind",
					"relation_name":     "server_bind_ip",
					"mapping":           "1",
				})
				require.NoError(t, err)
			},
			req: domain.ResourceRelation{
				SourceModelUID:   "server",
				TargetModelUID:   "ip",
				SourceResourceID: 101,
				TargetResourceID: 202,
				RelationTypeUID:  "bind",
			},
			wantName:  "server_bind_ip",
			wantModel: "server",
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			id, err := s.svc.CreateResourceRelation(context.Background(), tc.req)
			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Greater(t, id, int64(0))

			// 校验落库的实体，是否完全被补齐了 SourceModelUID, TargetModelUID, RelationTypeUID 字段（验证 toEntity 映射修复）
			var rr dao.ResourceRelation
			err = s.db.Database().Collection(dao.ResourceRelationCollection).FindOne(context.Background(), bson.M{"id": id}).Decode(&rr)
			require.NoError(t, err)
			require.Equal(t, tc.wantName, rr.RelationName)
			require.Equal(t, tc.wantModel, rr.SourceModelUID)
			require.Equal(t, "ip", rr.TargetModelUID)
			require.Equal(t, "bind", rr.RelationTypeUID)
		})
	}
}

func TestRRService(t *testing.T) {
	suite.Run(t, new(ServiceRRTestSuite))
}
