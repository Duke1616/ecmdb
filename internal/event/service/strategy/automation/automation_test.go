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
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAutomationNotification_Send(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBase := strategymocks.NewMockService(ctrl)
	mockSender := sendermocks.NewMockNotificationSender(ctrl)
	n := automation.NewAutomationNotification(mockBase, mockSender)

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
	mockBase.EXPECT().GetNodeProperty(info, "auto1").Return([]easyflow.Node{}, map[string]interface{}{
		"is_notify":     true,
		"notify_method": []int64{2}, // ProcessNowSend
	}, nil).Times(2)

	// 2. Mock IsGlobalNotify
	mockBase.EXPECT().IsGlobalNotify(gomock.Any()).Return(true)

	// 3. Mock FetchRequiredData
	mockBase.EXPECT().FetchRequiredData(gomock.Any(), gomock.Any(), gomock.Any()).Return(&strategy.NotificationData{
		TName: "测试模板",
		StartUser: user.User{
			FeishuInfo: user.FeishuInfo{
				UserId: "user123",
			},
		},
	}, nil)

	// 4. Mock Sender
	mockSender.EXPECT().Send(gomock.Any(), gomock.Any()).Return(notification.NewSuccessResponse(0, "success"), nil)

	resp, err := n.Send(ctx, info)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "success", resp.Status)
}
