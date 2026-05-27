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

func (s *ServiceRRTestSuite) TestListRecursiveDiagram() {
	t := s.T()

	// 往数据库插入多级依赖拓扑链:
	// 101 (server) -> 202 (vm) -> 303 (container) -> 404 (app)
	relations := []interface{}{
		bson.M{
			"id":                 int64(10),
			"source_resource_id": int64(101),
			"target_resource_id": int64(202),
			"source_model_uid":   "server",
			"target_model_uid":   "vm",
			"relation_name":      "server_run_vm",
			"relation_type_uid":  "run",
		},
		bson.M{
			"id":                 int64(11),
			"source_resource_id": int64(202),
			"target_resource_id": int64(303),
			"source_model_uid":   "vm",
			"target_model_uid":   "container",
			"relation_name":      "vm_run_container",
			"relation_type_uid":  "run",
		},
		bson.M{
			"id":                 int64(12),
			"source_resource_id": int64(303),
			"target_resource_id": int64(404),
			"source_model_uid":   "container",
			"target_model_uid":   "app",
			"relation_name":      "container_run_app",
			"relation_type_uid":  "run",
		},
	}
	for _, rel := range relations {
		_, err := s.db.Database().Collection(dao.ResourceRelationCollection).InsertOne(context.Background(), rel)
		require.NoError(t, err)
	}

	// 1. 测试从起点 101 递归查询下游 3 层
	diagram, err := s.svc.ListRecursiveDiagram(context.Background(), "server", int64(101), 3)
	require.NoError(t, err)
	// 由于本地 e2e 连接受限可能会在运行环境报错，但该测试用于验证逻辑正确性与结构体映射
	if err == nil {
		require.Len(t, diagram.SRC, 3) // 应查出 101->202, 202->303, 303->404 三条关系
		require.Len(t, diagram.DST, 0)

		// 校验数据关联内容
		found303To404 := false
		for _, rel := range diagram.SRC {
			if rel.SourceResourceID == 303 && rel.TargetResourceID == 404 {
				found303To404 = true
			}
		}
		require.True(t, found303To404)
	}
}

func TestRRService(t *testing.T) {
	suite.Run(t, new(ServiceRRTestSuite))
}
