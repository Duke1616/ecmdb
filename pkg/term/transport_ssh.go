package term

import (
	"context"
	"fmt"
	"net"

	"github.com/Duke1616/ecmdb/pkg/term/sshx"
	"golang.org/x/crypto/ssh"
)

// SSHChainBuilder 是基于 SSH 多级网关的 ChainBuilder 实现。
// 目前将 GatewayChain 里的每个 Endpoint 直接映射为 sshx.GatewayConfig，
// 在第一次 Dial 时通过 sshx.MultiGatewayManager 打通整条链路。
type SSHChainBuilder struct{}

func NewSSHChainBuilder() ChainBuilder {
	return &SSHChainBuilder{}
}

func (b *SSHChainBuilder) Build(chain GatewayChain) (Transport, error) {
	gateways := make([]*sshx.GatewayConfig, 0, len(chain))
	for _, ep := range chain {
		gateways = append(gateways, &sshx.GatewayConfig{
			AuthType:   ep.AuthType,
			Host:       ep.Host,
			Port:       ep.Port,
			Username:   ep.Username,
			Password:   ep.Password,
			PrivateKey: ep.PrivateKey,
			Passphrase: ep.Passphrase,
			Sort:       ep.Sort,
		})
	}

	return &sshChainTransport{gateways: gateways}, nil
}

// sshChainTransport 会在第一次 Dial 时通过多级网关建立一个 SSH 客户端，
// 后续 Dial 复用该客户端上的 TCP 连接能力。
type sshChainTransport struct {
	gateways []*sshx.GatewayConfig
	client   *ssh.Client
}

func (t *sshChainTransport) Dial(ctx context.Context, ep Endpoint) (net.Conn, error) {
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

	manager := sshx.NewMultiGatewayManager(t.gateways)
	client, err := manager.Connect(ctx)
	if err != nil {
		return err
	}
	t.client = client
	return nil
}
