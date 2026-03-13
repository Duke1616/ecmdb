package user_test

import (
	"context"
	"testing"

	"github.com/Bunny3th/easy-workflow/workflow/model"
	strategymocks "github.com/Duke1616/ecmdb/internal/event/mocks"
	"github.com/Duke1616/ecmdb/internal/event/service/strategy"
	"github.com/Duke1616/ecmdb/internal/event/service/strategy/user"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserNotification_Send(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBase := strategymocks.NewMockService(ctrl)

	n := user.NewUserNotification(mockBase, nil, nil, nil)

	ctx := context.Background()
	info := strategy.Info{
		FlowContext: strategy.FlowContext{
			InstID: 100,
			CurrentNode: &model.Node{
				NodeID: "node1",
			},
		},
	}

	// 1. Mock GetNodeProperty
	mockBase.EXPECT().GetNodeProperty(gomock.Any(), "node1").Return([]easyflow.Node{}, map[string]interface{}{
		"assignees": []interface{}{
			map[string]interface{}{
				"rule":   "LEADER",
				"values": []string{},
			},
		},
	}, nil)

	// 2. Mock ResolveAssignees
	mockBase.EXPECT().ResolveAssignees(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)

	// 3. Mock FetchRequiredData
	mockBase.EXPECT().FetchRequiredData(gomock.Any(), gomock.Any(), gomock.Any()).Return(&strategy.NotificationData{}, nil)

	// 4. Mock SafeGo
	mockBase.EXPECT().SafeGo(gomock.Any(), gomock.Any(), gomock.Any()).Return()

	resp, err := n.Send(ctx, info)
	assert.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
}
