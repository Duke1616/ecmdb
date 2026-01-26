package web

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/relation"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/Duke1616/ecmdb/pkg/term"
	"github.com/Duke1616/ecmdb/pkg/term/guacx"
	"github.com/Duke1616/ecmdb/pkg/term/sshx"
	sftpFinder "github.com/Duke1616/vuefinder-go/pkg/provider/sftp"
	finderWeb "github.com/Duke1616/vuefinder-go/pkg/web"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/sftp"
	"golang.org/x/sync/errgroup"
)

type Handler struct {
	RRSvc        relation.RRSvc
	resourceSvc  resource.EncryptedSvc
	attributeSvc attribute.Service
	session      *term.SessionPool
	timeout      time.Duration
	finderWeb    *finderWeb.Handler
}

func NewHandler(RRSvc relation.RRSvc, resourceSvc resource.EncryptedSvc, attributeSvc attribute.Service) *Handler {
	return &Handler{
		RRSvc:        RRSvc,
		resourceSvc:  resourceSvc,
		attributeSvc: attributeSvc,
		session:      term.NewSessionPool(),
		timeout:      5 * time.Second,
		finderWeb:    finderWeb.NewHandler(),
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

	// SSH 连接服务器，支持多层网关跳转
	g.GET("/ssh/session", ginx.Ws(h.SshSessionTunnel))

	// 主要用于连接管理成功后，存储到Session中，不需要重复建立连接
	g.POST("/connect", ginx.WrapBody(h.Connect))

	// 注册 FinderWeb 路由，实现 SFTP 能力
	h.finderWeb.RegisterRoutes(server)
	h.finderWeb.RegisterUploadRoute(server)
}

func (h *Handler) Connect(ctx *gin.Context, req ConnectReq) (ginx.Result, error) {
	switch req.Type {
	case ConnectTypeRDP:
		return ginx.Result{Msg: "不支持RDP协议"}, fmt.Errorf("暂不支持 RDP 协议")
	case ConnectTypeVNC:
		return ginx.Result{Msg: "不支持VNC协议"}, fmt.Errorf("暂不支持 VNC 协议")
	case ConnectTypeSSH:
		_, err := h.connectSSh(ctx, req.ResourceId)
		if err != nil {
			return ginx.Result{}, err
		}
	case ConnectTypeWebSftp:
		sess, err := h.connectSSh(ctx, req.ResourceId)
		if err != nil {
			return ginx.Result{}, err
		}

		// 通过能力接口获取 SFTP client
		fileCapable, ok := sess.(term.FileCapable)
		if !ok {
			return ginx.Result{Msg: "当前会话不支持 SFTP"}, fmt.Errorf("session does not implement FileCapable")
		}

		var sftpClient *sftp.Client
		sftpClient, err = fileCapable.NewSFTP()
		if err != nil {
			return ginx.Result{}, err
		}

		// 添加 sftp 信息
		h.finderWeb.SetFinder(req.ResourceId, sftpFinder.NewSftpFinder(sftpClient))
	default:
		return ginx.Result{Msg: fmt.Sprintf("不支持的连接类型: %s", req.Type)}, fmt.Errorf("unsupported connect type: %s", req.Type)
	}

	return ginx.Result{
		Msg: "SSH 连接成功",
	}, nil
}

func (h *Handler) connectSSh(ctx context.Context, resourceId int64) (term.Session, error) {
	// 获取指定资产关联网关数据
	hostResource, gatewayRs, err := h.queryResource(ctx, resourceId)
	if err != nil {
		return nil, fmt.Errorf("获取基本连接信息失败")
	}

	// 组合所有网关为通用 GatewayChain
	var chain term.GatewayChain
	for _, item := range gatewayRs {
		endpoint := term.Endpoint{
			Username:   sshx.GetStringField(item.Data, "username", ""),
			Host:       sshx.GetStringField(item.Data, "host", ""),
			PrivateKey: sshx.GetStringField(item.Data, "private_key", ""),
			Port:       sshx.GetIntField(item.Data, "port", 22),
			Password:   sshx.GetStringField(item.Data, "password", "default_password"),
			AuthType:   sshx.GetStringField(item.Data, "auth_type", "passwd"),
			Passphrase: sshx.GetStringField(item.Data, "password", "default_password"),
			Sort:       sshx.GetIntField(item.Data, "sort", 0),
		}
		chain = append(chain, endpoint)
	}

	// 组合真实的目标节点
	chain = append(chain, term.Endpoint{
		AuthType:   sshx.GetStringField(hostResource.Data, "auth_type", ""),
		Host:       sshx.GetStringField(hostResource.Data, "ip", ""),
		Port:       sshx.GetIntField(hostResource.Data, "port", 22),
		Username:   sshx.GetStringField(hostResource.Data, "username", ""),
		Password:   sshx.GetStringField(hostResource.Data, "password", ""),
		PrivateKey: sshx.GetStringField(hostResource.Data, "private_key", ""),
		Passphrase: sshx.GetStringField(hostResource.Data, "password", "passwd"),
		Sort:       len(chain) + 1,
	})

	// 通过插件连接网关和目标节点
	connector, ok := term.GetConnector("ssh")
	if !ok {
		return nil, fmt.Errorf("ssh connector not registered")
	}

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	sess, err := connector.Connect(ctxWithTimeout, chain, nil)
	if err != nil {
		return nil, fmt.Errorf("ssh connector fail")
	}

	// 每次连接都重新替换Session
	h.session.SetSession(resourceId, term.NewSessions(sess))

	return sess, nil
}

func (h *Handler) SshSessionTunnel(ctx *gin.Context) error {
	// 传递参数
	resourceId := ctx.Query("resource_id")
	resourceIdInt, err := strconv.ParseInt(resourceId, 10, 64)
	if err != nil {
		return err
	}

	cols := ctx.Query("cols")
	colsInt, err := strconv.Atoi(cols)
	if err != nil {
		return err
	}

	rows := ctx.Query("rows")
	rowsInt, err := strconv.Atoi(rows)
	if err != nil {
		return err
	}

	return h.wsSShSession(ctx, resourceIdInt, colsInt, rowsInt)
}

func (h *Handler) wsSShSession(ctx *gin.Context, resourceIdInt int64, colsInt, rowsInt int) error {
	var (
		err  error
		conn *websocket.Conn
	)

	// 升级 WebSocket 连接
	conn, err = UpGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	// 获取抽象 Session
	sessionWrapper, err := h.session.GetSession(resourceIdInt)
	if err != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		return err
	}

	sess := sessionWrapper.Session
	shellCapable, ok := sess.(term.ShellCapable)
	if !ok {
		_ = conn.WriteMessage(websocket.TextMessage, []byte("session not support shell"))
		return fmt.Errorf("session does not implement ShellCapable")
	}

	// 创建终端会话
	terminalSession, err := shellCapable.NewTerminal(conn, rowsInt, colsInt)
	if err != nil {
		return err
	}

	// 监听终端输出写入 websocket
	terminalSession.Start()
	defer terminalSession.Stop()

	// 接收 websocket 信息处理
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			var message []byte
			_, message, err = conn.ReadMessage()
			if err == io.EOF {
				return nil
			}

			if err != nil {
				return err
			}

			msg, er := sshx.ParseTerminalMessage(message)
			if er != nil {
				continue
			}

			switch msg.Operation {
			case "resize":
				if err = terminalSession.Resize(msg.Rows, msg.Cols); err != nil {
					return err
				}
			case "stdin":
				if err = terminalSession.Write([]byte(msg.Data)); err != nil {
					return err
				}
			case "ping":
				if err = terminalSession.Ping(); err != nil {
					return err
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

func (h *Handler) queryResource(ctx context.Context, resourceId int64) (resource.Resource, []resource.Resource, error) {
	var (
		eg           errgroup.Group
		hostResource resource.Resource
		gatewayRs    []resource.Resource
	)
	eg.Go(func() error {
		ids, err := h.RRSvc.ListDstRelated(ctx, "host", "AuthGateway_default_host", resourceId)
		if err != nil {
			return err
		}

		if len(ids) == 0 {
			return nil
		}

		fields, err := h.attributeSvc.SearchAllAttributeFieldsByModelUid(ctx, "AuthGateway")
		if err != nil {
			return err
		}

		gatewayRs, err = h.resourceSvc.ListResourceByIds(ctx, fields, ids)
		if err != nil {
			return err
		}
		return nil
	})
	eg.Go(func() error {
		fields, err := h.attributeSvc.SearchAllAttributeFieldsByModelUid(ctx, "host")
		if err != nil {
			return err
		}

		hostResource, err = h.resourceSvc.FindResourceById(ctx, fields, resourceId)
		if err != nil {
			return err
		}

		return nil
	})
	if err := eg.Wait(); err != nil {
		return resource.Resource{}, nil, err
	}

	return hostResource, gatewayRs, nil
}
