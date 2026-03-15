package integration

import (
	"context"
	"testing"
	"time"

	"github.com/Bunny3th/easy-workflow/workflow/model"
	notificationv1 "github.com/Duke1616/ecmdb/api/proto/gen/ealert/notification/v1"
	teamv1 "github.com/Duke1616/ecmdb/api/proto/gen/ealert/team"
	teammocks "github.com/Duke1616/ecmdb/api/proto/gen/ealert/team/mocks"
	"github.com/Duke1616/ecmdb/internal/department"
	departmentmocks "github.com/Duke1616/ecmdb/internal/department/mocks"
	"github.com/Duke1616/ecmdb/internal/engine"
	enginemocks "github.com/Duke1616/ecmdb/internal/engine/mocks"
	"github.com/Duke1616/ecmdb/internal/event/integration/startup"
	"github.com/Duke1616/ecmdb/internal/event/service/easyflow"
	"github.com/Duke1616/ecmdb/internal/order"
	ordermocks "github.com/Duke1616/ecmdb/internal/order/mocks"
	"github.com/Duke1616/ecmdb/internal/pkg/notification"
	sendermocks "github.com/Duke1616/ecmdb/internal/pkg/notification/mocks"
	"github.com/Duke1616/ecmdb/internal/rota"
	"github.com/Duke1616/ecmdb/internal/task"
	taskmocks "github.com/Duke1616/ecmdb/internal/task/mocks"
	"github.com/Duke1616/ecmdb/internal/template"
	templatemocks "github.com/Duke1616/ecmdb/internal/template/mocks"
	"github.com/Duke1616/ecmdb/internal/user"
	usermocks "github.com/Duke1616/ecmdb/internal/user/mocks"
	"github.com/Duke1616/ecmdb/internal/workflow"
	workflowmocks "github.com/Duke1616/ecmdb/internal/workflow/mocks"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type NotificationIntegrationTestSuite struct {
	suite.Suite
	ctrl *gomock.Controller
	app  *startup.TestApp

	// Service Mocks
	mockUser       *usermocks.MockService
	mockOrder      *ordermocks.MockService
	mockTask       *taskmocks.MockService
	mockTemplate   *templatemocks.MockService
	mockEngine     *enginemocks.MockService
	mockWorkflow   *workflowmocks.MockService
	mockDepartment *departmentmocks.MockService
	mockTeam       *teammocks.MockTeamServiceClient
	mockSender     *sendermocks.MockNotificationSender

	eventHandler *easyflow.ProcessEvent
}

func (s *NotificationIntegrationTestSuite) SetupSuite() {
	// 配置已通过 internal/test/ioc 包的 init() 函数自动加载
}

func (s *NotificationIntegrationTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())

	// 1. 初始化所有 Mock
	s.mockUser = usermocks.NewMockService(s.ctrl)
	s.mockOrder = ordermocks.NewMockService(s.ctrl)
	s.mockTask = taskmocks.NewMockService(s.ctrl)
	s.mockTemplate = templatemocks.NewMockService(s.ctrl)
	s.mockEngine = enginemocks.NewMockService(s.ctrl)
	s.mockWorkflow = workflowmocks.NewMockService(s.ctrl)
	s.mockDepartment = departmentmocks.NewMockService(s.ctrl)
	s.mockSender = sendermocks.NewMockNotificationSender(s.ctrl)
	s.mockTeam = teammocks.NewMockTeamServiceClient(s.ctrl)

	// 2. 构造测试所需的模块 (Module-level Mocking)
	userModule := &user.Module{Svc: s.mockUser}
	orderModule := &order.Module{Svc: s.mockOrder}
	taskModule := &task.Module{Svc: s.mockTask}
	templateModule := &template.Module{Svc: s.mockTemplate}
	engineModule := &engine.Module{Svc: s.mockEngine}
	workflowModule := &workflow.Module{Svc: s.mockWorkflow}
	departmentModule := &department.Module{Svc: s.mockDepartment}
	rotaModule := &rota.Module{} // 对于当前用例，暂时不需要 rota

	// 3. 使用 Wire 初始化的 App 并注入 Mocked Modules
	var err error
	s.app, err = startup.InitApp(
		s.mockSender,
		s.mockTeam,
		nil,
		userModule,
		orderModule,
		taskModule,
		templateModule,
		engineModule,
		workflowModule,
		departmentModule,
		rotaModule)
	require.NoError(s.T(), err)

	s.eventHandler = s.app.EventHandler
}

func (s *NotificationIntegrationTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *NotificationIntegrationTestSuite) TestEventNotify_UserApprovalFlow() {
	t := s.T()

	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		instID   int
		node     *model.Node
		prevNode model.Node
		wantErr  error
	}{
		{
			name: "审批节点通知_成功模拟用户与工单数据并校验发送",
			before: func(t *testing.T) {
				// 0. Mock Engine: 加载上下文所需元数据 ( LoadContext 调用 )
				s.mockEngine.EXPECT().GetOrderIdByVariable(gomock.Any(), 888).Return("888", nil).AnyTimes()
				s.mockEngine.EXPECT().GetInstanceByID(gomock.Any(), 888).Return(engine.Instance{
					ProcID:      1,
					ProcVersion: 1,
				}, nil).AnyTimes()

				// 0.1 Mock Engine: 获取节点任务信息 ( FetchTasksWithRetry 调用 )
				s.mockEngine.EXPECT().GetTasksByCurrentNodeId(gomock.Any(), 888, "node_user_1").Return([]model.Task{
					{TaskID: 1001, UserID: "approver_tester"},
				}, nil).AnyTimes()

				// 0.2 Mock Workflow: 加载流程定义快照 ( LoadContext 调用 )
				s.mockWorkflow.EXPECT().FindInstanceFlow(gomock.Any(), gomock.Any(), 1, 1).Return(workflow.Workflow{
					Id:       1,
					IsNotify: true,
					FlowData: workflow.LogicFlow{
						Nodes: []map[string]interface{}{
							{
								"id": "node_user_1",
								"properties": map[string]interface{}{
									"name":     "审批节点",
									"type":     "appoint",
									"approved": []string{"approver_tester"},
								},
							},
						},
					},
				}, nil).AnyTimes()

				// 0.3 Mock Template: 获取规则 ( FetchRequiredData 调用 )
				s.mockTemplate.EXPECT().DetailTemplate(gomock.Any(), gomock.Any()).Return(template.Template{
					Id:   1,
					Name: "测试模版888",
				}, nil).AnyTimes()

				// 1. Mock User: 获取审批人信息 ( ResolveAssignees 调用 )
				s.mockUser.EXPECT().FindByUsernames(gomock.Any(), []string{"approver_tester"}).Return([]user.User{
					{
						Id:          200,
						Username:    "approver_tester",
						DisplayName: "审批人",
						FeishuInfo:  user.FeishuInfo{UserId: "ou_56789"},
					},
				}, nil).AnyTimes()

				// 1.1 Mock User: 获取发起人信息 ( FetchRequiredData 调用 )
				s.mockUser.EXPECT().FindByUsername(gomock.Any(), gomock.Any()).Return(user.User{
					DisplayName: "发起人",
				}, nil).AnyTimes()

				// 2. Mock Order: 获取工单元数据 ( LoadContext 调用 )
				s.mockOrder.EXPECT().Detail(gomock.Any(), int64(888)).Return(order.Order{
					Id: 888,
					Data: map[string]interface{}{
						"app_name": "E-CMDB-Integration-Test",
					},
					WorkflowId: 1,
					CreateBy:   "start_user",
				}, nil).AnyTimes()

				done := make(chan struct{}, 1)
				t.Cleanup(func() {
					select {
					case <-done:
					case <-time.After(time.Second * 5):
						// t.Errorf("timeout waiting for notification")
					}
				})

				// 3. Mock Sender: 校验最终发送的通知 ( asyncSendNotification 调用 )
				s.mockSender.EXPECT().BatchSend(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, ns []notification.Notification) (notification.NotificationResponse, error) {
					defer func() {
						select {
						case done <- struct{}{}:
						default:
						}
					}()
					require.Len(t, ns, 1)
					noti := ns[0]
					require.Equal(t, "ou_56789", noti.Receiver)
					require.Contains(t, noti.Template.Title, "888")
					return notification.NotificationResponse{}, nil
				}).AnyTimes()
			},
			after:  func(t *testing.T) {},
			instID: 888,
			node: &model.Node{
				NodeID:   "node_user_1",
				NodeName: "审批节点",
				NodeType: model.TaskNode,
				UserIDs:  []string{"approver_tester"},
			},
			prevNode: model.Node{
				NodeID: "root",
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			err := s.eventHandler.EventNotify(tc.instID, tc.node, tc.prevNode)
			require.Equal(t, tc.wantErr, err)
			tc.after(t)
		})
	}
}

func (s *NotificationIntegrationTestSuite) TestChatNotification_WithMockedData() {
	t := s.T()

	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		instID   int
		node     *model.Node
		prevNode model.Node
		wantErr  error
	}{
		{
			name: "群聊通知节点_模拟外部依赖并校验卡片发送",
			before: func(t *testing.T) {
				// 0. Mock Engine: 加载上下文所需元数据
				s.mockEngine.EXPECT().GetOrderIdByVariable(gomock.Any(), 999).Return("999", nil).AnyTimes()
				s.mockEngine.EXPECT().GetInstanceByID(gomock.Any(), 999).Return(engine.Instance{
					ProcID:      1,
					ProcVersion: 1,
				}, nil).AnyTimes()

				s.mockEngine.EXPECT().GetTasksByCurrentNodeId(gomock.Any(), 999, "node_chat_1").Return([]model.Task{
					{TaskID: 2001, UserID: "system"},
				}, nil).AnyTimes()

				s.mockEngine.EXPECT().Pass(gomock.Any(), 2001, "ChatGroup Auto Pass").Return(nil).AnyTimes()

				// 0.1 Mock Workflow: 加载流程定义快照
				s.mockWorkflow.EXPECT().FindInstanceFlow(gomock.Any(), gomock.Any(), 1, 1).Return(workflow.Workflow{
					Id:       2,
					IsNotify: true,
					FlowData: workflow.LogicFlow{
						Nodes: []map[string]interface{}{
							{
								"id": "node_chat_1",
							},
						},
					},
				}, nil).AnyTimes()

				// 0.2 Mock Template: 获取规则
				s.mockTemplate.EXPECT().DetailTemplate(gomock.Any(), gomock.Any()).Return(template.Template{
					Id:   2,
					Name: "测试群谈模版999",
				}, nil).AnyTimes()

				// 1. Mock User: 获取发起人信息
				s.mockUser.EXPECT().FindByUsername(gomock.Any(), gomock.Any()).Return(user.User{
					DisplayName: "发起人",
				}, nil).AnyTimes()

				// 2. Mock Order
				s.mockOrder.EXPECT().Detail(gomock.Any(), int64(999)).Return(order.Order{
					Id: 999,
					Data: map[string]interface{}{
						"app_name": "E-CMDB-System-Mock",
					},
					WorkflowId: 2,
					CreateBy:   "start_user",
				}, nil).AnyTimes()

				// 3. Mock Team Service (群聊信息)
				s.mockTeam.EXPECT().GetChatGroupByIds(gomock.Any(), gomock.Any()).Return(&teamv1.GetChatGroupByIdsResponse{
					Groups: []*teamv1.ChatGroup{
						{ChatId: "oc_lark_group_1", Name: "交付群", Channel: notificationv1.Channel_LARK_CARD},
					},
				}, nil).AnyTimes()

				done := make(chan struct{}, 1)
				t.Cleanup(func() {
					select {
					case <-done:
					case <-time.After(time.Second * 5):
						// t.Errorf("timeout waiting for notification")
					}
				})

				// 4. Mock Sender
				s.mockSender.EXPECT().BatchSend(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, ns []notification.Notification) (notification.NotificationResponse, error) {
					defer func() {
						select {
						case done <- struct{}{}:
						default:
						}
					}()
					require.Len(t, ns, 1)
					noti := ns[0]
					require.Contains(t, noti.Template.Title, "999")
					require.Equal(t, "oc_lark_group_1", noti.Receiver)
					return notification.NotificationResponse{}, nil
				}).AnyTimes()
			},
			after:  func(t *testing.T) {},
			instID: 999,
			node: &model.Node{
				NodeID:   "node_chat_1",
				NodeName: "通知群聊",
				NodeType: model.TaskNode,
			},
			prevNode: model.Node{
				NodeID: "node_user_1",
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			err := s.eventHandler.EventChatGroup(tc.instID, tc.node, tc.prevNode)
			require.Equal(t, tc.wantErr, err)
			tc.after(t)
		})
	}
}

func TestNotificationIntegration(t *testing.T) {
	suite.Run(t, new(NotificationIntegrationTestSuite))
}
