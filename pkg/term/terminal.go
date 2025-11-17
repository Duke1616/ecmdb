package term

import (
	"context"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pkg/sftp"
)

// Endpoint 表示一个通用的远程访问端点（可以是网关，也可以是最终目标主机）。
type Endpoint struct {
	Host       string
	Port       int
	Username   string
	Password   string
	PrivateKey string
	Passphrase string
	AuthType   string
	Sort       int
}

// GatewayChain 代表从入口到目标主机的一条多级跳板链路。
type GatewayChain []Endpoint

// ConnectOptions 为不同协议预留的扩展参数。
type ConnectOptions map[string]any

// Connector 抽象出终端连接能力，不关心具体协议实现（SSH/RDP/VNC/K8s exec 等）。
type Connector interface {
	// Name 返回协议/插件名称，例如 "ssh"、"rdp"、"k8s_exec" 等。
	Name() string

	// Connect 根据多级网关链路建立到目标的连接，返回抽象的 Session
	Connect(ctx context.Context, chain GatewayChain, opts ConnectOptions) (Session, error)
}

// Session 表示一次建立好的远程会话（可能是 SSH、RDP、K8s exec 等）。
// 具体能力通过能力接口进行扩展。
type Session interface {
	Protocol() string
	Close() error
	Transport() Transport
}

// TerminalSession 抽象了一个与终端交互的会话能力，用于绑定 websocket。
type TerminalSession interface {
	Start()
	Stop()
	Resize(rows, cols int) error
	Write(data []byte) error
	Ping() error
}

// ShellCapable 表示 Session 具备创建交互式终端的能力。
type ShellCapable interface {
	NewTerminal(ws *websocket.Conn, rows, cols int) (TerminalSession, error)
}

// FileCapable 表示 Session 具备文件管理能力，例如 SFTP。
type FileCapable interface {
	NewSFTP() (*sftp.Client, error)
}

var (
	registryMu        sync.RWMutex
	connectorRegistry = make(map[string]Connector)
)

// RegisterConnector 注册一个终端连接插件。
func RegisterConnector(c Connector) {
	registryMu.Lock()
	defer registryMu.Unlock()

	connectorRegistry[c.Name()] = c
}

// GetConnector 按名称获取插件实现。
func GetConnector(name string) (Connector, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	c, ok := connectorRegistry[name]
	return c, ok
}
