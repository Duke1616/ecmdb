package chat_test

import (
	"context"
	"testing"
	"time"

	"github.com/Bunny3th/easy-workflow/workflow/model"
	notificationv1 "github.com/Duke1616/ecmdb/api/proto/gen/ealert/notification/v1"
	teamv1 "github.com/Duke1616/ecmdb/api/proto/gen/ealert/team"
	teammocks "github.com/Duke1616/ecmdb/api/proto/gen/ealert/team/mocks"
	strategymocks "github.com/Duke1616/ecmdb/internal/event/mocks"
	"github.com/Duke1616/ecmdb/internal/event/service/strategy"
	"github.com/Duke1616/ecmdb/internal/event/service/strategy/chat"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/pkg/notification"
	sendermocks "github.com/Duke1616/ecmdb/internal/pkg/notification/mocks"
	"github.com/Duke1616/ecmdb/internal/pkg/rule"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/gotomicro/ego/core/elog"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type ChatTestSuite struct {
	suite.Suite
	ctrl       *gomock.Controller
	mockBase   *strategymocks.MockService
	mockSender *sendermocks.MockNotificationSender
	mockTeam   *teammocks.MockTeamServiceClient
	strategy   strategy.SendStrategy
}

func (s *ChatTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockBase = strategymocks.NewMockService(s.ctrl)
	s.mockSender = sendermocks.NewMockNotificationSender(s.ctrl)
	s.mockTeam = teammocks.NewMockTeamServiceClient(s.ctrl)

	s.strategy = chat.NewNotification(s.mockBase, s.mockSender, nil, s.mockTeam)
	s.mockBase.EXPECT().Logger().Return(elog.DefaultLogger).AnyTimes()
}

func (s *ChatTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *ChatTestSuite) TestSend_ExistingGroup() {
	ctx := context.Background()
	info := strategy.Info{
		FlowContext: strategy.FlowContext{
			InstID: 501,
			CurrentNode: &model.Node{
				NodeID: "chat_node_1",
			},
			Order: order.Order{
				Id:      2001,
				Provide: order.SystemProvide,
				Data: map[string]interface{}{
					"app_name": "TestApp",
					"project":  "ECEDB",
				},
			},
		},
	}

	s.mockBase.EXPECT().GetNodeProperty(gomock.Any(), "chat_node_1").Return([]easyflow.Node{}, map[string]interface{}{
		"mode":           "existing",
		"chat_group_ids": []int64{888},
		"title":          "工单#{{ticket_id}} - {{field.app_name}} 结果通知",
		"is_auto":        []interface{}{"ticket_data", "user_input"},
	}, nil)

	s.mockBase.EXPECT().FetchRequiredData(gomock.Any(), gomock.Any(), gomock.Any()).Return(&strategy.NotificationData{
		TName:     "应用发版",
		StartUser: user.User{DisplayName: "陆安"},
		Rules: []rule.Rule{
			{Title: "项目名称", Field: "project", Type: "input"},
		},
	}, nil)

	s.mockBase.EXPECT().ResolveAssignees(gomock.Any(), gomock.Any(), gomock.Any()).Return([]user.User{}, nil).AnyTimes()

	s.mockBase.EXPECT().FindTaskForms(gomock.Any(), int64(2001)).Return([]order.FormValue{
		{Name: "版本号", Value: "v1.0.0"},
	}, nil)

	s.mockBase.EXPECT().SafeGo(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, timeout time.Duration, f func(context.Context)) {
		f(ctx)
	})

	s.mockBase.EXPECT().FetchTasksWithRetry(gomock.Any(), gomock.Any()).Return([]model.Task{{TaskID: 70001}}, nil)
	s.mockBase.EXPECT().IsGlobalNotify(gomock.Any()).Return(true)
	s.mockTeam.EXPECT().GetChatGroupByIds(gomock.Any(), gomock.Any()).Return(&teamv1.GetChatGroupByIdsResponse{
		Groups: []*teamv1.ChatGroup{
			{ChatId: "oc_chat123", Name: "交付群", Channel: notificationv1.Channel_LARK_CARD},
		},
	}, nil)
	s.mockBase.EXPECT().PassTask(gomock.Any(), 70001, gomock.Any()).Return(nil)

	s.mockSender.EXPECT().BatchSend(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, ns []notification.Notification) (notification.NotificationResponse, error) {
		s.Len(ns, 1)
		noti := ns[0]
		s.Equal("工单#2001 - TestApp 结果通知", noti.Template.Title)

		var dividerCount int
		for _, f := range noti.Template.Fields {
			if f.IsDivider {
				dividerCount++
				if dividerCount == 1 {
					s.Equal("**📋 工单信息**", f.Content)
				} else if dividerCount == 2 {
					s.Equal("**✍️ 用户提交**", f.Content)
				}
			}
		}
		s.Equal(2, dividerCount)

		var foundProject, foundVersion bool
		for _, f := range noti.Template.Fields {
			if f.Content == "**项目名称:**\\nECEDB" {
				foundProject = true
			}
			if f.Content == "**版本号:**\nv1.0.0" {
				foundVersion = true
			}
		}
		s.True(foundProject, "Project info not found")
		s.True(foundVersion, "Version info not found")

		return notification.NotificationResponse{}, nil
	})

	resp, err := s.strategy.Send(ctx, info)
	s.NoError(err)
	s.NotNil(resp)
}

func TestChatSuite(t *testing.T) {
	suite.Run(t, new(ChatTestSuite))
}
