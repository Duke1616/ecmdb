package mongox

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func Test_ToFilterMap_Isolation(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		modify   func(cloned bson.M)
		assertFn func(t *testing.T, orig interface{})
	}{
		{
			name:  "克隆隔离_修改新map不污染原始bson.M",
			input: bson.M{"status": "active", "type": 1},
			modify: func(cloned bson.M) {
				cloned["status"] = "inactive"
				cloned["new_field"] = "new_value"
			},
			assertFn: func(t *testing.T, orig interface{}) {
				m := orig.(bson.M)
				assert.Equal(t, "active", m["status"])
				_, exists := m["new_field"]
				assert.False(t, exists)
			},
		},
		{
			name:  "边界场景_入参为nil返回空map且安全隔离",
			input: nil,
			modify: func(cloned bson.M) {
				cloned["key"] = "val"
			},
			assertFn: func(t *testing.T, orig interface{}) {
				assert.Nil(t, orig)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cloned := toFilterMap(tc.input)
			tc.modify(cloned)
			tc.assertFn(t, tc.input)
		})
	}
}

func Test_TryMergeTenantID(t *testing.T) {
	col := &Collection[any]{}

	testCases := []struct {
		name       string
		pipeline   mongo.Pipeline
		tenantID   int64
		wantMerged bool
		assertFn   func(t *testing.T, p mongo.Pipeline)
	}{
		{
			name:       "空管道_合并失败",
			pipeline:   mongo.Pipeline{},
			tenantID:   100,
			wantMerged: false,
		},
		{
			name: "首阶段为match且值为bson.M_成功合并tenant_id",
			pipeline: mongo.Pipeline{
				bson.D{{Key: "$match", Value: bson.M{"status": 1}}},
			},
			tenantID:   100,
			wantMerged: true,
			assertFn: func(t *testing.T, p mongo.Pipeline) {
				m := p[0][0].Value.(bson.M)
				assert.Equal(t, int64(100), m["tenant_id"])
				assert.Equal(t, 1, m["status"])
			},
		},
		{
			name: "首阶段为match且值为bson.D_成功追加tenant_id",
			pipeline: mongo.Pipeline{
				bson.D{{Key: "$match", Value: bson.D{{Key: "status", Value: 1}}}},
			},
			tenantID:   100,
			wantMerged: true,
			assertFn: func(t *testing.T, p mongo.Pipeline) {
				d := p[0][0].Value.(bson.D)
				assert.Equal(t, "tenant_id", d[1].Key)
				assert.Equal(t, int64(100), d[1].Value)
			},
		},
		{
			name: "首阶段非match阶段_返回false由上层prepend",
			pipeline: mongo.Pipeline{
				bson.D{{Key: "$project", Value: bson.M{"_id": 1}}},
			},
			tenantID:   100,
			wantMerged: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			merged := col.tryMergeTenantID(tc.pipeline, tc.tenantID)
			assert.Equal(t, tc.wantMerged, merged)
			if tc.assertFn != nil {
				tc.assertFn(t, tc.pipeline)
			}
		})
	}
}

func Test_GenerateIndexName(t *testing.T) {
	testCases := []struct {
		name string
		keys interface{}
		want string
	}{
		{
			name: "多字段复合索引键_拼接生成标准索引名",
			keys: bson.D{{Key: "name", Value: 1}, {Key: "age", Value: -1}},
			want: "name_1_age_-1",
		},
		{
			name: "空索引键_返回空字符串",
			keys: nil,
			want: "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, generateIndexName(tc.keys))
		})
	}
}

func Test_Aggregate_FailClosed(t *testing.T) {
	col := &Collection[any]{}

	testCases := []struct {
		name     string
		ctx      context.Context
		pipeline mongo.Pipeline
		wantErr  error
	}{
		{
			name:     "无租户上下文且未提权_直接返回ErrMissingTenantContext",
			ctx:      context.Background(),
			pipeline: mongo.Pipeline{},
			wantErr:  ErrMissingTenantContext,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := col.Aggregate(tc.ctx, tc.pipeline)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
