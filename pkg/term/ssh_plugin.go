package term

import (
	"context"
	"fmt"
	"net"

	"github.com/Duke1616/ecmdb/pkg/term/sshx"
	"github.com/gorilla/websocket"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// sshConnector 是基于现有 sshx 多级网关能力的 Connector 实现。
type sshConnector struct{}

func (s *sshConnector) Name() string {
	return "ssh"
}

// Connect 将 GatewayChain 转换为 sshx 的配置，并返回抽象的 Session。
func (s *sshConnector) Connect(ctx context.Context, chain GatewayChain, opts ConnectOptions) (Session, error) {
	// 使用 SSHChainBuilder 构建多跳链路对应的 Transport
	builder := NewSSHChainBuilder()
	transport, err := builder.Build(chain)
	if err != nil {
		return nil, err
	}

	// 通过 Transport 确保链路打通并拿到底层 *ssh.Client
	sshTransport, ok := transport.(*sshChainTransport)
	if !ok {
		return nil, fmt.Errorf("unexpected transport type for ssh connector")
	}
	if err = sshTransport.ensureClient(ctx); err != nil {
		return nil, err
	}

	return &sshSession{client: sshTransport.client}, nil
}

// sshSession 是对底层 *ssh.Client 的抽象包装，实现 Session 接口以及能力接口。
type sshSession struct {
	client *ssh.Client
}

func (s *sshSession) Protocol() string {
	return "ssh"
}

func (s *sshSession) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

// Transport 返回基于当前 SSH Client 的传输实现。
func (s *sshSession) Transport() Transport {
	return &sshTransport{client: s.client}
}

// sshTransport 使用底层 *ssh.Client 作为传输层，根据 Endpoint 建立到目标的 TCP 连接。
type sshTransport struct {
	client *ssh.Client
}

func (t *sshTransport) Dial(ctx context.Context, ep Endpoint) (net.Conn, error) {
	if t.client == nil {
		return nil, fmt.Errorf("ssh transport client is nil")
	}
	address := fmt.Sprintf("%s:%d", ep.Host, ep.Port)
	return t.client.DialContext(ctx, "tcp", address)
}

// NewTerminal 实现 ShellCapable 能力，基于现有 sshx.SSHConnect 创建终端会话。
func (s *sshSession) NewTerminal(ws *websocket.Conn, rows, cols int) (TerminalSession, error) {
	sshConn, err := sshx.NewSSHConnect(s.client, ws, rows, cols)
	if err != nil {
		return nil, err
	}
	return &sshTerminalSession{SSHConnect: sshConn, client: s.client}, nil
}

// NewSFTP 实现 FileCapable 能力，返回基于 SSH 的 SFTP 客户端。
func (s *sshSession) NewSFTP() (*sftp.Client, error) {
	return sftp.NewClient(s.client)
}

// sshTerminalSession 将 sshx.SSHConnect 适配为通用 TerminalSession。
type sshTerminalSession struct {
	*sshx.SSHConnect
	client *ssh.Client
}

func (t *sshTerminalSession) Start() {
	t.SSHConnect.Start()
}

func (t *sshTerminalSession) Stop() {
	t.SSHConnect.Stop()
}

func (t *sshTerminalSession) Resize(rows, cols int) error {
	return t.WindowChange(rows, cols)
}

func (t *sshTerminalSession) Write(data []byte) error {
	_, err := t.StdinPipe.Write(data)
	return err
}

func (t *sshTerminalSession) Ping() error {
	if t.client == nil {
		return nil
	}
	_, _, err := t.client.Conn.SendRequest("PING", true, nil)
	return err
}

// 确保 sshSession 实现了 Session 以及能力接口。
var _ Session = (*sshSession)(nil)
var _ ShellCapable = (*sshSession)(nil)
var _ FileCapable = (*sshSession)(nil)

func init() {
	RegisterConnector(&sshConnector{})
}
