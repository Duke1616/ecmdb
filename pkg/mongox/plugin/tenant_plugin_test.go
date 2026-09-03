package plugin

import (
	"context"
	"testing"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/Duke1616/eiam/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

type TestPrivateModel struct {
	ID       int64  `bson:"id" eiam:"private"`
	Name     string `bson:"name"`
	TenantID int64  `bson:"tenant_id"`
}

type TestSharedModel struct {
	ID       int64  `bson:"id" eiam:"shared:is_public=true"`
	Name     string `bson:"name"`
	TenantID int64  `bson:"tenant_id"`
}

type TestIgnoreModel struct {
	ID   int64  `bson:"id" eiam:"ignore"`
	Name string `bson:"name"`
}

func Test_TenantPlugin_ParseCondition(t *testing.T) {
	testCases := []struct {
		name     string
		condStr  string
		assertFn func(t *testing.T, cond bson.M)
	}{
		{
			name:    "JSON格式条件_正确解析属性与布尔值",
			condStr: `{"status":1,"is_shared":true}`,
			assertFn: func(t *testing.T, cond bson.M) {
				assert.Equal(t, float64(1), cond["status"])
				assert.Equal(t, true, cond["is_shared"])
			},
		},
		{
			name:    "逗号分隔键值对_正确解析数字与字符串",
			condStr: "is_public=true,count=10,name=system",
			assertFn: func(t *testing.T, cond bson.M) {
				assert.Equal(t, true, cond["is_public"])
				assert.Equal(t, int64(10), cond["count"])
				assert.Equal(t, "system", cond["name"])
			},
		},
		{
			name:    "空条件字符串_返回空map",
			condStr: "",
			assertFn: func(t *testing.T, cond bson.M) {
				assert.Empty(t, cond)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cond := parseCondition(tc.condStr)
			tc.assertFn(t, cond)
		})
	}
}

func Test_TenantPlugin_BuildFindFilter(t *testing.T) {
	p := NewTenantPlugin(WithTenantField("tenant_id"), WithSystemTenantID(0))
	ctx := context.Background()

	testCases := []struct {
		name     string
		conf     SharedConfig
		tid      int64
		assertFn func(t *testing.T, filter bson.M)
	}{
		{
			name: "私有模型_普通租户注入绝对隔离条件",
			conf: SharedConfig{IsPrivate: true},
			tid:  100,
			assertFn: func(t *testing.T, filter bson.M) {
				assert.Equal(t, bson.M{"tenant_id": int64(100)}, filter)
			},
		},
		{
			name: "共享模型_普通租户注入本租户与系统公共数据OR条件",
			conf: SharedConfig{
				IsShared:  true,
				Condition: bson.M{"is_public": true},
			},
			tid: 100,
			assertFn: func(t *testing.T, filter bson.M) {
				assert.Contains(t, filter, "$or")
				orBranches := filter["$or"].([]bson.M)
				assert.Len(t, orBranches, 2)
				assert.Equal(t, bson.M{"tenant_id": int64(100)}, orBranches[0])
				assert.Equal(t, bson.M{"tenant_id": int64(0), "is_public": true}, orBranches[1])
			},
		},
		{
			name: "共享模型_系统租户自身仅注入系统租户条件",
			conf: SharedConfig{
				IsShared:  true,
				Condition: bson.M{"is_public": true},
			},
			tid: 0,
			assertFn: func(t *testing.T, filter bson.M) {
				assert.Equal(t, bson.M{"tenant_id": int64(0)}, filter)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			filter := p.buildFindFilter(ctx, tc.conf, tc.tid)
			tc.assertFn(t, filter)
		})
	}
}

func Test_TenantPlugin_Lifecycle(t *testing.T) {
	p := NewTenantPlugin()

	testCases := []struct {
		name     string
		stmt     *mongox.Statement
		assertFn func(t *testing.T, stmt *mongox.Statement, err error)
	}{
		{
			name: "缺失租户上下文且未提权_触发FailClosed阻断报错",
			stmt: &mongox.Statement{
				Model:   &TestPrivateModel{},
				Filter:  bson.M{"name": "test"},
				Context: context.Background(),
			},
			assertFn: func(t *testing.T, stmt *mongox.Statement, err error) {
				assert.ErrorIs(t, err, mongox.ErrMissingTenantContext)
			},
		},
		{
			name: "ignore标记模型_无租户上下文依然全局放行",
			stmt: &mongox.Statement{
				Model:   &TestIgnoreModel{},
				Filter:  bson.M{"name": "test"},
				Context: context.Background(),
			},
			assertFn: func(t *testing.T, stmt *mongox.Statement, err error) {
				assert.NoError(t, err)
				assert.Equal(t, bson.M{"name": "test"}, stmt.Filter)
			},
		},
		{
			name: "携带提权标记上下文_无租户上下文直接放行",
			stmt: &mongox.Statement{
				Model:   &TestPrivateModel{},
				Filter:  bson.M{"name": "test"},
				Context: IgnoreTenantContext(context.Background()),
			},
			assertFn: func(t *testing.T, stmt *mongox.Statement, err error) {
				assert.NoError(t, err)
				assert.Equal(t, bson.M{"name": "test"}, stmt.Filter)
			},
		},
		{
			name: "具备正常租户上下文_私有模型成功注入租户条件",
			stmt: &mongox.Statement{
				Model:   &TestPrivateModel{},
				Filter:  bson.M{"name": "test"},
				Context: ctxutil.WithTenantID(context.Background(), 100),
			},
			assertFn: func(t *testing.T, stmt *mongox.Statement, err error) {
				assert.NoError(t, err)
				assert.Contains(t, stmt.Filter, "$and")
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := p.BeforeFind(tc.stmt)
			tc.assertFn(t, tc.stmt, err)
		})
	}
}
