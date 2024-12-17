package web

import (
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/Duke1616/ecmdb/pkg/guacx"
	"github.com/Duke1616/ecmdb/pkg/sshx"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
)

type Handler struct {
	RRSvc        relation.RRSvc
	resourceSvc  resource.Service
	attributeSvc attribute.Service
}

func NewHandler(RRSvc relation.RRSvc, resourceSvc resource.Service, attributeSvc attribute.Service) *Handler {
	return &Handler{
		RRSvc:        RRSvc,
		resourceSvc:  resourceSvc,
		attributeSvc: attributeSvc,
	}
}

var UpGrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	Subprotocols: []string{"guacamole"},
}

func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/term")
	g.GET("/guac/tunnel", ginx.Wrap(h.ConnectGuacTunnel))
	g.GET("/ssh/tunnel", ginx.Wrap(h.ConnectSshTunnel))
}

func (h *Handler) ConnectSshTunnel(ctx *gin.Context) (ginx.Result, error) {
	var (
		conn    *websocket.Conn
		sshConn *sshx.SSHConnect
		err     error
	)

	// 升级 WebSocket 连接
	conn, err = UpGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return ginx.Result{
			Msg: "WebSocket upgrade failed",
		}, err
	}
	defer conn.Close()

	// 获取关联数据
	ids, err := h.RRSvc.ListDstRelated(ctx, "host", "gateway_default_host", 57)
	if err != nil {
		return ginx.Result{}, err
	}

	// 查看所有字段信息
	fields, err := h.attributeSvc.SearchAllAttributeFieldsByModelUid(ctx, "gateway")
	rs, err := h.resourceSvc.ListResourceByIds(ctx, fields, ids)
	if err != nil {
		return ginx.Result{}, err
	}

	// 组合所有网关
	var multiGateways []*sshx.GatewayConfig
	for _, item := range rs {
		gateway := &sshx.GatewayConfig{
			Username:   sshx.GetStringField(item.Data, "username", ""),
			Host:       sshx.GetStringField(item.Data, "host", ""),
			PrivateKey: sshx.GetStringField(item.Data, "private_key", ""),
			Port:       sshx.GetIntField(item.Data, "port", 22),
			Password:   sshx.GetStringField(item.Data, "password", "default_password"),
			AuthType:   sshx.GetStringField(item.Data, "auth_type", "passwd"),
			Passphrase: sshx.GetStringField(item.Data, "password", "default_password"),
			Sort:       sshx.GetIntField(item.Data, "sort", 0),
		}

		multiGateways = append(multiGateways, gateway)
	}
	manager := sshx.NewMultiGatewayManager(multiGateways)

	filesHost, err := h.attributeSvc.SearchAllAttributeFieldsByModelUid(ctx, "host")
	if err != nil {
		return ginx.Result{}, err
	}

	host, err := h.resourceSvc.FindResourceById(ctx, filesHost, 57)
	if err != nil {
		return ginx.Result{}, err
	}

	client, err := manager.Connect(&sshx.GatewayConfig{
		AuthType:   sshx.GetStringField(host.Data, "auth_type", ""),
		Host:       sshx.GetStringField(host.Data, "ip", ""),
		Port:       sshx.GetIntField(host.Data, "port", 22),
		Username:   sshx.GetStringField(host.Data, "username", ""),
		Password:   sshx.GetStringField(host.Data, "password", ""),
		PrivateKey: sshx.GetStringField(host.Data, "private_key", ""),
		Passphrase: sshx.GetStringField(host.Data, "password", "passwd"),
		Sort:       0,
	})
	if err != nil {
		return ginx.Result{}, err
	}

	cols := ctx.Query("cols")
	colsInt, err := strconv.Atoi(cols)
	rows := ctx.Query("rows")
	rowsInt, err := strconv.Atoi(rows)
	if sshConn, err = sshx.NewSSHConnect(client, conn, rowsInt, colsInt); err != nil {
		return ginx.Result{
			Msg: "WebSocket create client failed",
		}, err
	}

	sshConn.Start()
	defer sshConn.Stop()

	for {
		select {
		case <-ctx.Done():
			return ginx.Result{}, nil
		default:
			_, message, er := conn.ReadMessage()
			if er != nil {
				return ginx.Result{}, er
			}

			msg, er := sshx.ParseTerminalMessage(message)
			if er != nil {
				continue
			}

			switch msg.Operation {
			case "resize":
				if err = sshConn.WindowChange(msg.Rows, msg.Cols); err != nil {
				}
			case "stdin":
				_, err = sshConn.StdinPipe.Write([]byte(msg.Data))
				if err != nil {
					return ginx.Result{}, err
				}
			case "ping":
				_, _, err = client.Conn.SendRequest("PING", true, nil)
				if err != nil {
					return ginx.Result{}, err
				}
			}
		}
	}
}

func (h *Handler) ConnectGuacTunnel(ctx *gin.Context) (ginx.Result, error) {
	ws, err := UpGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	defer ws.Close()

	if err != nil {
		return ginx.Result{
			Msg: "websocket upgrade failed",
		}, err
	}

	cfg := guacx.NewConfig()
	cfg.SetParameter("width", ctx.Query("width"))
	cfg.SetParameter("height", ctx.Query("height"))
	cfg.SetParameter("dpi", ctx.Query("dpi"))

	cfg.SetParameter("hostname", "")
	cfg.SetParameter("port", "")
	cfg.SetParameter("username", "")
	cfg.SetParameter("password", "")
	cfg.SetParameter("scheme", "rdp")
	cfg.Protocol = "rdp"

	tunnel, err := guacx.NewTunnel("", cfg)
	if err != nil {

	}
	err = tunnel.Handshake()
	if err != nil {
		return ginx.Result{
			Msg: "tunnel handshake failed",
		}, err
	}

	guacHandler := guacx.NewGuacamoleHandler(ws, tunnel)
	guacHandler.Start()
	defer guacHandler.Stop()

	for {
		_, message, er := ws.ReadMessage()
		if er != nil {
			_ = tunnel.Close()
			return ginx.Result{}, er
		}
		_, er = tunnel.Write(message)
		if er != nil {
			return ginx.Result{}, er
		}
	}
}
