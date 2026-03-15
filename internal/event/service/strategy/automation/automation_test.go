package automation_test

import (
	"context"
	"testing"

	"github.com/Bunny3th/easy-workflow/workflow/model"
	strategymocks "github.com/Duke1616/ecmdb/internal/event/mocks"
	"github.com/Duke1616/ecmdb/internal/event/service/strategy"
	"github.com/Duke1616/ecmdb/internal/event/service/strategy/automation"
	"github.com/Duke1616/ecmdb/internal/pkg/notification"
	sendermocks "github.com/Duke1616/ecmdb/internal/pkg/notification/mocks"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type AutomationTestSuite struct {
	suite.Suite
	ctrl       *gomock.Controller
	mockBase   *strategymocks.MockService
	mockSender *sendermocks.MockNotificationSender
	strategy   strategy.SendStrategy
}

func (s *AutomationTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockBase = strategymocks.NewMockService(s.ctrl)
	s.mockSender = sendermocks.NewMockNotificationSender(s.ctrl)

	s.strategy = automation.NewNotification(s.mockBase, s.mockSender)
}

func (s *AutomationTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *AutomationTestSuite) TestSend() {
	ctx := context.Background()
	info := strategy.Info{
		FlowContext: strategy.FlowContext{
			InstID: 200,
			CurrentNode: &model.Node{
				NodeID: "auto1",
			},
		},
	}

	// 1. Mock GetNodeProperty
	s.mockBase.EXPECT().GetNodeProperty(info, "auto1").Return([]easyflow.Node{}, map[string]interface{}{
		"is_notify":     true,
		"notify_method": []int64{2}, // ProcessNowSend
	}, nil).Times(2)

	// 2. Mock IsGlobalNotify
	s.mockBase.EXPECT().IsGlobalNotify(gomock.Any()).Return(true)

	// 3. Mock FetchRequiredData
	s.mockBase.EXPECT().FetchRequiredData(gomock.Any(), gomock.Any(), gomock.Any()).Return(&strategy.NotificationData{
		TName: "测试模板",
		StartUser: user.User{
			FeishuInfo: user.FeishuInfo{
				UserId: "user123",
			},
		},
	}, nil)

	// 4. Mock Sender
	s.mockSender.EXPECT().Send(gomock.Any(), gomock.Any()).Return(notification.NewSuccessResponse(0, "success"), nil)

	resp, err := s.strategy.Send(ctx, info)
	s.NoError(err)
	s.NotNil(resp)
	s.Equal("success", resp.Status)
}

func TestAutomationSuite(t *testing.T) {
	suite.Run(t, new(AutomationTestSuite))
}
