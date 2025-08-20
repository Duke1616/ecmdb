package order

import (
	notificationv1 "github.com/Duke1616/ecmdb/api/proto/gen/notification/v1"
	"github.com/Duke1616/ecmdb/internal/order/internal/domain"
	"github.com/Duke1616/ecmdb/internal/order/internal/grpc"
	"github.com/Duke1616/ecmdb/internal/order/internal/service"
	"github.com/Duke1616/ecmdb/internal/order/internal/web"
)

type Service = service.Service

type RpcServer = grpc.WorkOrderServer

type Handler = web.Handler

const (
	EndProcess    = domain.END
	SystemProvide = domain.SYSTEM
	WechatProvide = domain.WECHAT
)

type Order = domain.Order

var protoToDomain = map[notificationv1.Channel]domain.Channel{
	notificationv1.Channel_FEISHU_CARD: domain.ChannelFeishuCard,
	notificationv1.Channel_EMAIL:       domain.ChannelEmail,
	notificationv1.Channel_IN_APP:      domain.ChannelInApp,
}

var domainToProto = map[domain.Channel]notificationv1.Channel{
	domain.ChannelFeishuCard: notificationv1.Channel_FEISHU_CARD,
	domain.ChannelEmail:      notificationv1.Channel_EMAIL,
	domain.ChannelInApp:      notificationv1.Channel_IN_APP,
}

func ChannelToDomainProto(ch domain.Channel) notificationv1.Channel {
	if v, ok := domainToProto[ch]; ok {
		return v
	}
	return notificationv1.Channel_CHANNEL_UNSPECIFIED
}

func ChannelToProtoDomain(ch notificationv1.Channel) domain.Channel {
	if v, ok := protoToDomain[ch]; ok {
		return v
	}
	return ""
}
