package ssh

import (
	"fmt"
	"sort"

	"github.com/Duke1616/ecmdb/pkg/plugin"
	"github.com/Duke1616/ecmdb/pkg/term"
)

const (
	ActionTerminal = "terminal"
	ActionSFTP     = "sftp"
)

const (
	pluginUID     = "builtin.ssh"
	connectorName = "ssh"
	inputTarget   = "target"
)

// Endpoint 描述 SSH 连接需要从资源属性中解析出的字段。
type Endpoint struct {
	Host       string `plugin:"host,field=ip,required"`
	Port       int    `plugin:"port,default=22"`
	Username   string `plugin:"username,required"`
	Password   string `plugin:"password"`
	PrivateKey string `plugin:"private_key"`
	AuthType   string `plugin:"auth_type"`
	Sort       int    `plugin:"sort"`
}

// Gateway 描述 SSH 网关模型需要解析出的字段。
type Gateway struct {
	Host       string `plugin:"host,field=host,required"`
	Port       int    `plugin:"port,default=22"`
	Username   string `plugin:"username,required"`
	Password   string `plugin:"password"`
	PrivateKey string `plugin:"private_key"`
	AuthType   string `plugin:"auth_type"`
	Sort       int    `plugin:"sort"`
}

// Target 描述以内置 host 模型为中心的 SSH 输入树。
type Target struct {
	Endpoint
	Gateways []Gateway `plugin:"gateways,model=AuthGateway,in=default"`
}

// Definition 返回内置 SSH 插件的声明，导入命令会将它保存到指定租户。
func Definition() plugin.Definition {
	return plugin.NewRegistry(
		pluginUID,
		"SSH",
		plugin.Type("builtin"),
		plugin.Version("1.0.0"),
	).
		Action(
			ActionTerminal,
			"SSH 终端",
			plugin.Icon("terminal"),
			plugin.UI(plugin.UIBuiltinTerminal),
		).
		Action(
			ActionSFTP,
			"文件管理",
			plugin.Icon("folder"),
			plugin.UI(plugin.UIBuiltinSFTP),
		).
		Setup(
			plugin.ModelGroup("主机模型"),
			plugin.ModelGroup("网关模型"),
			plugin.RelationTypes(plugin.BasicRelationTypes()...),
			hostModel(),
			gatewayModel(),
			plugin.Relation("AuthGateway", plugin.RelationTypeDefault, "host").OneToMany(),
		).
		Bind(plugin.Center[Target]("host")).
		MustDefinition()
}

// ResolveRequest 创建内置 SSH 插件动作的解析请求。
func ResolveRequest(action string, resourceID int64) plugin.ResolveRequest {
	return plugin.ResolveRequest{
		PluginID:   pluginUID,
		Action:     action,
		ResourceID: resourceID,
	}
}

// DecodeTarget 从插件动作上下文中读取 SSH 目标输入。
func DecodeTarget(actionCtx plugin.ActionContext) (Target, error) {
	return plugin.InputOne[Target](actionCtx, inputTarget)
}

// Connector 返回 SSH 终端连接器。
func Connector() (term.Connector, error) {
	connector, ok := term.GetConnector(connectorName)
	if !ok {
		return nil, fmt.Errorf("ssh connector not registered")
	}
	return connector, nil
}

// ToGatewayChain 将插件解析出的强类型输入转换为终端连接层可消费的网关链路。
func (t Target) ToGatewayChain() term.GatewayChain {
	sort.SliceStable(t.Gateways, func(i, j int) bool {
		return t.Gateways[i].Sort < t.Gateways[j].Sort
	})

	chain := make(term.GatewayChain, 0, len(t.Gateways)+1)
	for _, gateway := range t.Gateways {
		chain = append(chain, gateway.ToEndpoint())
	}

	target := t.Endpoint.ToEndpoint()
	target.Sort = len(chain) + 1
	chain = append(chain, target)
	return chain
}

// ToEndpoint 将资源字段转换为通用终端端点。
func (e Endpoint) ToEndpoint() term.Endpoint {
	return toEndpoint(e.Host, e.Port, e.Username, e.Password, e.PrivateKey, e.AuthType, e.Sort)
}

// ToEndpoint 将网关资源字段转换为通用终端端点。
func (g Gateway) ToEndpoint() term.Endpoint {
	return toEndpoint(g.Host, g.Port, g.Username, g.Password, g.PrivateKey, g.AuthType, g.Sort)
}

func toEndpoint(host string, port int, username, password, privateKey, authType string, sort int) term.Endpoint {
	if authType == "" {
		authType = "passwd"
	}

	return term.Endpoint{
		Host:       host,
		Port:       port,
		Username:   username,
		Password:   password,
		PrivateKey: privateKey,
		AuthType:   authType,
		Passphrase: password,
		Sort:       sort,
	}
}

func hostModel() plugin.ModelSpec {
	return plugin.Model(
		"host",
		"主机",
		plugin.ModelIcon("monitor-host"),
		plugin.ModelGroupName("主机模型"),
	).
		AttrGroup("基础属性", 0,
			plugin.String("name", "名称").Required().Display().Index(0),
			plugin.String("ip", "IP地址").Required().Display().Index(1),
			plugin.String("port", "端口").Display().Index(2),
			plugin.String("username", "用户名").Display().Index(3),
			plugin.List("auth_type", "认证类型", authOptions()).Required().Display().Index(6),
		).
		AttrGroup("加密属性", 2,
			plugin.String("password", "密码").Secure().Index(1),
			plugin.Multiline("private_key", "私钥").Secure().Index(2),
		).
		Build()
}

func gatewayModel() plugin.ModelSpec {
	return plugin.Model(
		"AuthGateway",
		"登陆网关",
		plugin.ModelIcon("ops-oneterm-login"),
		plugin.ModelGroupName("网关模型"),
	).
		AttrGroup("基础属性", 0,
			plugin.String("name", "名称").Required().Display().Index(0),
			plugin.String("host", "地址").Required().Display().Index(1),
			plugin.String("port", "端口").Display().Index(2),
			plugin.String("username", "用户名").Display().Index(3),
		).
		AttrGroup("分类属性", 1,
			plugin.List("auth_type", "认证类型", authOptions()).Display().Index(1),
			plugin.String("sort", "排序").Display().Index(2),
		).
		AttrGroup("加密属性", 2,
			plugin.String("password", "密码").Secure().Index(1),
			plugin.Multiline("private_key", "私钥").Secure().Index(2),
		).
		Build()
}

func authOptions() []string {
	return []string{"passwd", "publickey", "passphrase"}
}
