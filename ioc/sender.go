package ioc

import (
	notificationv1 "github.com/Duke1616/ecmdb/api/proto/gen/ealert/notification/v1"
	"github.com/Duke1616/ecmdb/internal/pkg/notification"
	"github.com/Duke1616/ecmdb/internal/pkg/notification/channel"
	"github.com/Duke1616/ecmdb/internal/pkg/notification/provider"
	"github.com/Duke1616/ecmdb/internal/pkg/notification/provider/feishu"
	"github.com/Duke1616/ecmdb/internal/pkg/notification/provider/sequential"
	"github.com/Duke1616/ecmdb/internal/pkg/notification/sender"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/google/wire"
	lark "github.com/larksuite/oapi-sdk-go/v3"
)

var InitSender = wire.NewSet(
	sender.NewSender,
	newSelectorBuilder,
	newChannel,
)

func newChannel(builder *sequential.SelectorBuilder) channel.Channel {
	return channel.NewDispatcher(map[notification.Channel]channel.Channel{
		notification.ChannelFeishu: channel.NewLarkCardChannel(builder),
	})
}

func newSelectorBuilder(
	lark *lark.Client,
	notificationSvc notificationv1.NotificationServiceClient,
	workflowSvc workflow.Service,
) *sequential.SelectorBuilder {
	// 构建SMS供应商
	providers := make([]provider.Provider, 0)
	cardProvider, err := feishu.NewLarkCardProvider(lark)
	if err != nil {
		return nil
	}

	grpcNotificationProvider := feishu.NewGRPCProvider(notificationSvc, workflowSvc)
	providers = append(providers, grpcNotificationProvider, cardProvider)
	return sequential.NewSelectorBuilder(providers)
}
