package user_test

import (
	"context"
	"testing"
	"time"

	"github.com/Bunny3th/easy-workflow/workflow/model"
	notificationv1 "github.com/Duke1616/ecmdb/api/proto/gen/ealert/notification/v1"
	notificationmocks "github.com/Duke1616/ecmdb/api/proto/gen/ealert/notification/v1/mocks"
	strategymocks "github.com/Duke1616/ecmdb/internal/event/mocks"
	"github.com/Duke1616/ecmdb/internal/event/service/strategy"
	"github.com/Duke1616/ecmdb/internal/event/service/strategy/user"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/pkg/notification"
	sendermocks "github.com/Duke1616/ecmdb/internal/pkg/notification/mocks"
	internaluser "github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/gotomicro/ego/core/elog"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
)

type UserTestSuite struct {
	suite.Suite
	ctrl        *gomock.Controller
	mockBase    *strategymocks.MockService
	mockSender  *sendermocks.MockNotificationSender
	mockNotiSvc *notificationmocks.MockNotificationServiceClient
	strategy    strategy.SendStrategy
}

func (s *UserTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockBase = strategymocks.NewMockService(s.ctrl)
	s.mockSender = sendermocks.NewMockNotificationSender(s.ctrl)
	s.mockNotiSvc = notificationmocks.NewMockNotificationServiceClient(s.ctrl)

	s.strategy = user.NewNotification(s.mockBase, nil, s.mockSender, s.mockNotiSvc)
	s.mockBase.EXPECT().Logger().Return(elog.DefaultLogger).AnyTimes()
}

func (s *UserTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *UserTestSuite) TestSend_StandardApproval() {
	ctx := context.Background()
	info := strategy.Info{
		FlowContext: strategy.FlowContext{
			InstID: 101,
			CurrentNode: &model.Node{
				NodeID: "audit_node",
			},
			Order: order.Order{
				Id:      1001,
				Provide: order.SystemProvide,
			},
		},
	}

	s.mockBase.EXPECT().GetNodeProperty(gomock.Any(), "audit_node").Return([]easyflow.Node{}, map[string]interface{}{
		"fields": []interface{}{
			map[string]interface{}{
				"name":     "审批意见",
				"key":      "remark",
				"type":     "string",
				"required": true,
			},
		},
	}, nil)

	s.mockBase.EXPECT().ResolveAssignees(gomock.Any(), gomock.Any(), gomock.Any()).Return([]internaluser.User{
		{Username: "approver1", FeishuInfo: internaluser.FeishuInfo{UserId: "fs_888"}},
	}, nil)

	s.mockBase.EXPECT().FetchRequiredData(gomock.Any(), gomock.Any(), gomock.Any()).Return(&strategy.NotificationData{
		TName: "权限申请",
	}, nil)

	s.mockBase.EXPECT().SafeGo(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, timeout time.Duration, f func(context.Context)) {
		f(ctx)
	})

	s.mockBase.EXPECT().FetchTasksWithRetry(gomock.Any(), gomock.Any()).Return([]model.Task{
		{TaskID: 50001, UserID: "approver1"},
	}, nil)

	s.mockBase.EXPECT().IsGlobalNotify(gomock.Any()).Return(true)

	s.mockBase.EXPECT().PrepareCommonFields(gomock.Any(), gomock.Any()).Return([]notification.Field{
		{Tag: "lark_md", Content: "Order: 1001"},
	})

	s.mockSender.EXPECT().BatchSend(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, ns []notification.Notification) (notification.NotificationResponse, error) {
		s.Len(ns, 1)
		noti := ns[0]
		hasField := false
		for _, f := range noti.Template.InputFields {
			if f.Key == "remark" && f.Name == "审批意见" {
				hasField = true
				break
			}
		}
		s.True(hasField, "Should have remark field")

		hasTaskID := false
		for _, v := range noti.Template.Values {
			if v.Key == "task_id" {
				if val, ok := v.Value.(int); ok && val == 50001 {
					hasTaskID = true
				} else if val, ok := v.Value.(int64); ok && val == 50001 {
					hasTaskID = true
				}
			}
		}
		s.True(hasTaskID, "Should have task_id value")
		s.Equal("fs_888", noti.Receiver)
		return notification.NotificationResponse{}, nil
	})

	resp, err := s.strategy.Send(ctx, info)
	s.NoError(err)
	s.Equal("success", resp.Status)
}

func (s *UserTestSuite) TestSend_AlertOrder() {
	ctx := context.Background()
	info := strategy.Info{
		FlowContext: strategy.FlowContext{
			InstID: 102,
			CurrentNode: &model.Node{
				NodeID: "alert_node",
			},
			Order: order.Order{
				Id:      1002,
				Provide: order.AlertProvide,
				NotificationConf: order.NotificationConf{
					TemplateID: 12345,
					Channel:    order.ChannelLarkCard,
					TemplateParams: map[string]interface{}{
						"alert_name": "CPU Usage High",
					},
				},
			},
		},
	}

	s.mockBase.EXPECT().GetNodeProperty(gomock.Any(), "alert_node").Return([]easyflow.Node{}, map[string]interface{}{}, nil)
	s.mockBase.EXPECT().ResolveAssignees(gomock.Any(), gomock.Any(), gomock.Any()).Return([]internaluser.User{
		{Username: "op1", FeishuInfo: internaluser.FeishuInfo{UserId: "fs_op1"}},
	}, nil)
	s.mockBase.EXPECT().FetchRequiredData(gomock.Any(), gomock.Any(), gomock.Any()).Return(&strategy.NotificationData{}, nil)
	s.mockBase.EXPECT().SafeGo(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, timeout time.Duration, f func(context.Context)) {
		f(ctx)
	})
	s.mockBase.EXPECT().FetchTasksWithRetry(gomock.Any(), gomock.Any()).Return([]model.Task{
		{TaskID: 60001, UserID: "op1"},
	}, nil)

	s.mockBase.EXPECT().IsGlobalNotify(gomock.Any()).Return(true)

	s.mockNotiSvc.EXPECT().SendNotification(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *notificationv1.SendNotificationRequest, opts ...grpc.CallOption) (*notificationv1.SendNotificationResponse, error) {
		s.Equal(int64(12345), req.Notification.TemplateId)
		s.Len(req.Notification.Receivers, 1)
		s.Equal("fs_op1", req.Notification.Receivers[0])
		return &notificationv1.SendNotificationResponse{}, nil
	})

	resp, err := s.strategy.Send(ctx, info)
	s.NoError(err)
	s.Equal("success", resp.Status)
}

func TestUserSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}
