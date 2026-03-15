package integration

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Bunny3th/easy-workflow/workflow/model"
	ioc "github.com/Duke1616/ecmdb/internal/test/ioc"
	"github.com/Duke1616/ecmdb/internal/workflow/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type ConvertIntegrationTestSuite struct {
	suite.Suite
	db          *gorm.DB
	workflowDAO dao.WorkflowDAO
	converter   easyflow.Converter
}

func (s *ConvertIntegrationTestSuite) SetupSuite() {
	s.db = ioc.InitMySQLDB()
	s.workflowDAO = dao.NewWorkflowDAO(ioc.InitMongoDB())
	s.converter = easyflow.NewLogicFlowToEngineConvert()
}

func TestConvertIntegration(t *testing.T) {
	suite.Run(t, new(ConvertIntegrationTestSuite))
}

// TestConvert_AllWorkflows 从 MongoDB 加载所有 workflow，转换后与 MySQL 中存储的 process definition 比对
func (s *ConvertIntegrationTestSuite) TestConvert_AllWorkflows() {
	t := s.T()
	ctx := context.Background()

	workflows, err := s.workflowDAO.List(ctx, 0, 1000)
	require.NoError(t, err)
	require.NotEmpty(t, workflows, "MongoDB 中没有 workflow 数据")

	t.Logf("共加载 %d 个 workflow", len(workflows))

	for _, wf := range workflows {
		t.Run(wf.Name, func(t *testing.T) {
			if wf.ProcessId == 0 {
				t.Skipf("未发布，跳过")
			}

			// List 接口排除了 flow_data 大字段，重新 Find 获取完整数据
			full, err := s.workflowDAO.Find(ctx, wf.Id)
			require.NoError(t, err)

			converted, err := s.converter.Convert(easyflow.Workflow{
				Id:    full.Id,
				Name:  full.Name,
				Owner: full.Owner,
				FlowData: easyflow.LogicFlow{
					Nodes: full.FlowData.Nodes,
					Edges: full.FlowData.Edges,
				},
			})
			require.NoError(t, err)

			stored, err := s.getLatestProcessDef(full.ProcessId)
			require.NoError(t, err)

			s.assertNodesMatch(t, stored.Nodes, converted.Nodes)
		})
	}
}

// assertNodesMatch 双向比对节点集合，确保无多出也无缺失
func (s *ConvertIntegrationTestSuite) assertNodesMatch(t *testing.T, stored []processNode, converted []model.Node) {
	t.Helper()

	assert.Equal(t, len(stored), len(converted), "节点数量不一致")

	storedSet := make(map[string]struct{}, len(stored))
	for _, n := range stored {
		storedSet[n.NodeID] = struct{}{}
	}
	convertedSet := make(map[string]struct{}, len(converted))
	for _, n := range converted {
		convertedSet[n.NodeID] = struct{}{}
	}

	for _, n := range converted {
		assert.Contains(t, storedSet, n.NodeID, "转换后多出节点: %s (%s)", n.NodeID, n.NodeName)
	}
	for _, n := range stored {
		assert.Contains(t, convertedSet, n.NodeID, "转换后缺少节点: %s (%s)", n.NodeID, n.NodeName)
	}
}

// getLatestProcessDef 从 MySQL 读取最新版本的 process definition
func (s *ConvertIntegrationTestSuite) getLatestProcessDef(processID int) (processDef, error) {
	var resource string
	err := s.db.Table("(?) as t",
		s.db.Raw("? UNION ALL ?",
			s.db.Table("proc_def").Select("resource, version").Where("id = ?", processID),
			s.db.Table("hist_proc_def").Select("resource, version").Where("proc_id = ?", processID),
		),
	).Select("resource").Order("version DESC").Limit(1).Scan(&resource).Error
	if err != nil {
		return processDef{}, err
	}

	var result processDef
	if err = json.Unmarshal([]byte(resource), &result); err != nil {
		return processDef{}, err
	}
	return result, nil
}

type processDef struct {
	Nodes []processNode `json:"Nodes"`
}

type processNode struct {
	NodeID   string `json:"NodeID"`
	NodeName string `json:"NodeName"`
	NodeType int    `json:"NodeType"`
}
