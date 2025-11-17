package term

import (
	"context"
	"net"
)

// Transport 抽象了多跳链路上的传输能力，负责根据 Endpoint 建立底层连接。
// 后续可以有多种实现：直连、SSH 多跳、代理隧道等。
type Transport interface {
	Dial(ctx context.Context, ep Endpoint) (net.Conn, error)
}

// ChainBuilder 负责将一条 GatewayChain 构建为一个 Transport。
// 不同的实现可以支持不同的多跳策略（例如纯 SSH、多协议混合等）。
type ChainBuilder interface {
	Build(chain GatewayChain) (Transport, error)
}
