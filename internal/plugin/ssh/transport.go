package ssh

import (
	"context"
	"fmt"
	"net"

	"github.com/Duke1616/ecmdb/pkg/term"
	"github.com/Duke1616/ecmdb/pkg/term/sshx"
	golangssh "golang.org/x/crypto/ssh"
)

// SSHChainBuilder 是基于 SSH 多级网关的 ChainBuilder 实现。
type SSHChainBuilder struct{}

func NewSSHChainBuilder() term.ChainBuilder {
	return &SSHChainBuilder{}
}

func (b *SSHChainBuilder) Build(chain term.GatewayChain) (term.Transport, error) {
	return &sshChainTransport{gateways: chain}, nil
}

// sshChainTransport 会在第一次 Dial 时通过多级网关建立一个 SSH 客户端，
// 后续 Dial 复用该客户端上的 TCP 连接能力。
type sshChainTransport struct {
	gateways []term.Endpoint
	client   *golangssh.Client
}

func (t *sshChainTransport) Dial(ctx context.Context, ep term.Endpoint) (net.Conn, error) {
	if err := t.ensureClient(ctx); err != nil {
		return nil, err
	}

	address := fmt.Sprintf("%s:%d", ep.Host, ep.Port)
	return t.client.DialContext(ctx, "tcp", address)
}

// ensureClient 确保通过多级网关建立好 SSH 客户端，允许被其他组件复用。
func (t *sshChainTransport) ensureClient(ctx context.Context) error {
	if len(t.gateways) == 0 {
		return fmt.Errorf("no gateways configured")
	}

	if t.client != nil {
		return nil
	}

	client, err := sshx.Connect(ctx, t.gateways)
	if err != nil {
		return err
	}
	t.client = client
	return nil
}
