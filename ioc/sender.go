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
	newCardSelectorBuilder,
	newTextSelectorBuilder,
	newChannel,
)

func newChannel(card *sequential.CardSelectorBuilder, text *sequential.TextSelectorBuilder) channel.Channel {
	return channel.NewDispatcher(map[notification.Channel]channel.Channel{
		notification.ChannelLarkCard: channel.NewLarkCardChannel(card),
		notification.ChannelLarkText: channel.NewLarkTextChannel(text),
	})
}

func newCardSelectorBuilder(
	lark *lark.Client,
	notificationSvc notificationv1.NotificationServiceClient,
	workflowSvc workflow.Service,
) *sequential.CardSelectorBuilder {
	// 构建飞书卡片供应商
	providers := make([]provider.Provider, 0)
	cardProvider, err := feishu.NewLarkCardProvider(lark)
	if err != nil {
		return nil
	}

	grpcNotificationProvider := feishu.NewGRPCProvider(notificationSvc, workflowSvc)
	providers = append(providers, grpcNotificationProvider, cardProvider)
	// 构造 SelectorBuilder 并包装成 CardSelectorBuilder
	return &sequential.CardSelectorBuilder{SelectorBuilder: sequential.NewSelectorBuilder(providers)}
}

func newTextSelectorBuilder(
	lark *lark.Client,
) *sequential.TextSelectorBuilder {
	// 构建飞书文本供应商
	providers := make([]provider.Provider, 0)
	provideText, err := feishu.NewLarkTextProvider(lark)
	if err != nil {
		return nil
	}
	providers = append(providers, provideText)
	// 构造 SelectorBuilder 并包装成 TextSelectorBuilder
	return &sequential.TextSelectorBuilder{SelectorBuilder: sequential.NewSelectorBuilder(providers)}
}
