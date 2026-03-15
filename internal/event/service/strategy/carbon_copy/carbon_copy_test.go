package carbon_copy_test

import (
	"context"
	"testing"
	"time"

	"github.com/Bunny3th/easy-workflow/workflow/model"
	strategymocks "github.com/Duke1616/ecmdb/internal/event/mocks"
	"github.com/Duke1616/ecmdb/internal/event/service/strategy"
	"github.com/Duke1616/ecmdb/internal/event/service/strategy/carbon_copy"
	"github.com/Duke1616/ecmdb/internal/pkg/notification"
	sendermocks "github.com/Duke1616/ecmdb/internal/pkg/notification/mocks"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/gotomicro/ego/core/elog"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type CarbonCopyTestSuite struct {
	suite.Suite
	ctrl       *gomock.Controller
	mockBase   *strategymocks.MockService
	mockSender *sendermocks.MockNotificationSender
	strategy   strategy.SendStrategy
}

func (s *CarbonCopyTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockBase = strategymocks.NewMockService(s.ctrl)
	s.mockSender = sendermocks.NewMockNotificationSender(s.ctrl)

	s.strategy = carbon_copy.NewNotification(s.mockBase, s.mockSender)
	s.mockBase.EXPECT().Logger().Return(elog.DefaultLogger).AnyTimes()
}

func (s *CarbonCopyTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *CarbonCopyTestSuite) TestSend() {
	ctx := context.Background()
	info := strategy.Info{
		FlowContext: strategy.FlowContext{
			InstID: 300,
			CurrentNode: &model.Node{
				NodeID: "cc1",
			},
		},
	}

	// 1. Mock GetNodeProperty
	s.mockBase.EXPECT().GetNodeProperty(info, "cc1").Return([]easyflow.Node{}, map[string]interface{}{
		"is_cc": true,
	}, nil)

	// 2. Mock FetchRequiredData
	s.mockBase.EXPECT().FetchRequiredData(gomock.Any(), gomock.Any(), gomock.Any()).Return(&strategy.NotificationData{
		TName: "抄送模板",
		StartUser: user.User{
			DisplayName: "发起人",
		},
	}, nil)

	// 3. Mock ResolveAssignees
	s.mockBase.EXPECT().ResolveAssignees(gomock.Any(), gomock.Any(), gomock.Any()).Return([]user.User{
		{Username: "user1", FeishuInfo: user.FeishuInfo{UserId: "fs_user1"}},
	}, nil)

	// 4. Mock SafeGo to run synchronously for testing
	s.mockBase.EXPECT().SafeGo(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, timeout time.Duration, f func(context.Context)) {
		f(ctx)
	})

	// 5. Inside async part
	// Mock FetchTasksWithRetry
	tasks := []model.Task{{TaskID: 1001, UserID: "user1"}}
	s.mockBase.EXPECT().FetchTasksWithRetry(gomock.Any(), gomock.Any()).Return(tasks, nil)

	// Mock IsGlobalNotify
	s.mockBase.EXPECT().IsGlobalNotify(gomock.Any()).Return(true)

	// Mock PrepareCommonFields
	s.mockBase.EXPECT().PrepareCommonFields(gomock.Any(), gomock.Any()).Return([]notification.Field{})

	// Mock Sender.Send
	s.mockSender.EXPECT().Send(gomock.Any(), gomock.Any()).Return(notification.NewSuccessResponse(0, "success"), nil)

	// Mock PassTask
	s.mockBase.EXPECT().PassTask(gomock.Any(), 1001, "抄送节点自动通过").Return(nil)

	resp, err := s.strategy.Send(ctx, info)
	s.NoError(err)
	s.Equal("success", resp.Status)
}

func TestCarbonCopySuite(t *testing.T) {
	suite.Run(t, new(CarbonCopyTestSuite))
}
