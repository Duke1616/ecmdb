package web

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	sshplugin "github.com/Duke1616/ecmdb/internal/plugin/ssh"
	pluginservice "github.com/Duke1616/ecmdb/internal/service/plugin"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/Duke1616/ecmdb/pkg/term"
	"github.com/Duke1616/ecmdb/pkg/term/guacx"
	"github.com/Duke1616/ecmdb/pkg/term/sshx"
	"github.com/Duke1616/eiam/pkg/web/capability"
	sftpFinder "github.com/Duke1616/vuefinder-go/pkg/provider/sftp"
	finderWeb "github.com/Duke1616/vuefinder-go/pkg/web"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/sftp"
)

type Handler struct {
	pluginSvc pluginservice.Service
	session   *term.SessionPool
	timeout   time.Duration
	finderWeb *finderWeb.Handler
	capability.IRegistry
}

func NewHandler(pluginSvc pluginservice.Service) *Handler {
	return &Handler{
		pluginSvc: pluginSvc,
		session:   term.NewSessionPool(),
		timeout:   5 * time.Second,
		finderWeb: finderWeb.NewHandler(),
		IRegistry: capability.NewRegistry("cmdb", "terminal", "资产仓库/在线终端"),
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
	g.GET("/guac/tunnel", h.Capability("远程连接", "guac_tunnel").
		Handle(ginx.Wrap(h.ConnectGuacTunnel)),
	)

	// SSH 连接服务器，支持多层网关跳转
	g.GET("/ssh/session", h.Capability("终端会话", "ssh_session").
		Handle(ginx.Ws(h.SshSessionTunnel)),
	)

	// 主要用于连接管理成功后，存储到Session中，不需要重复建立连接
	g.POST("/connect", h.Capability("终端连接验证", "connect").
		Handle(ginx.WrapBody(h.Connect)),
	)

	// 自行接管 FinderWeb (SFTP) 路由注册并注入权限控制cat mig
	gFinder := server.Group("/api/finder")

	gFinder.GET("/files", h.Capability("查看文件", "sftp_files").
		Handle(wrapFinder(h.finderWeb.Index)),
	)
	gFinder.GET("/download", h.Capability("下载文件", "sftp_download").
		Handle(h.finderWeb.DownloadStream),
	)
	gFinder.GET("/search", h.Capability("搜索文件", "sftp_search").
		Handle(wrapFinder(h.finderWeb.Search)),
	)
	gFinder.GET("/preview", h.Capability("预览文件", "sftp_preview").
		Handle(wrapFinderBuff(h.finderWeb.Preview)),
	)
	gFinder.POST("/new_folder", h.Capability("创建目录", "sftp_new_folder").
		Handle(wrapFinderBody(h.finderWeb.NewFolder)),
	)
	gFinder.POST("/new_file", h.Capability("创建文件", "sftp_new_file").
		Handle(wrapFinderBody(h.finderWeb.NewFile)),
	)
	gFinder.POST("/rename", h.Capability("重命名文件", "sftp_rename").
		Handle(wrapFinderBody(h.finderWeb.Rename)),
	)
	gFinder.POST("/move", h.Capability("移动文件", "sftp_move").
		Handle(wrapFinderBody(h.finderWeb.Move)),
	)
	gFinder.POST("/archive", h.Capability("压缩文件", "sftp_archive").
		Handle(wrapFinderBody(h.finderWeb.Archive)),
	)
	gFinder.POST("/unarchive", h.Capability("解压文件", "sftp_unarchive").
		Handle(wrapFinderBody(h.finderWeb.Unarchive)),
	)
	gFinder.POST("/save", h.Capability("保存文件内容", "sftp_save").
		Handle(wrapFinderBuffBody(h.finderWeb.Save)),
	)
	gFinder.POST("/delete", h.Capability("删除文件", "sftp_delete").
		Handle(wrapFinderBody(h.finderWeb.Delete)),
	)
	// WebSocket 上传路由
	gFinder.GET("/upload/ws", h.Capability("上传文件", "sftp_upload_ws").
		Handle(func(ctx *gin.Context) {
			finderWeb.UploadHandler(h.finderWeb)(ctx.Writer, ctx.Request)
		}),
	)
}

func (h *Handler) Connect(ctx *gin.Context, req ConnectReq) (ginx.Result, error) {
	switch req.Type {
	case ConnectTypeRDP:
		return ginx.Result{Msg: "不支持RDP协议"}, fmt.Errorf("暂不支持 RDP 协议")
	case ConnectTypeVNC:
		return ginx.Result{Msg: "不支持VNC协议"}, fmt.Errorf("暂不支持 VNC 协议")
	case ConnectTypeSSH:
		_, err := h.connectSSh(ctx, req.ResourceId, sshplugin.ActionTerminal)
		if err != nil {
			return ginx.Result{}, err
		}
	case ConnectTypeWebSftp:
		sess, err := h.connectSSh(ctx, req.ResourceId, sshplugin.ActionSFTP)
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

func (h *Handler) connectSSh(ctx context.Context, resourceId int64, action string) (term.Session, error) {
	actionCtx, err := h.pluginSvc.ResolveActionContext(ctx, sshplugin.ResolveRequest(action, resourceId))
	if err != nil {
		return nil, fmt.Errorf("获取 SSH 插件输入失败: %w", err)
	}

	target, err := sshplugin.DecodeTarget(actionCtx)
	if err != nil {
		return nil, fmt.Errorf("解析 SSH 插件输入失败: %w", err)
	}
	chain := target.ToGatewayChain()

	// 通过插件连接网关和目标节点
	connector, err := sshplugin.Connector()
	if err != nil {
		return nil, err
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	sess, err := connector.Connect(ctxWithTimeout, chain, nil)
	if err != nil {
		return nil, fmt.Errorf("ssh connector fail")
	}

	// 每次连接都重新替换Session
	h.session.SetSession(resourceId, sess)

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
	sess, err := h.session.GetSession(resourceIdInt)
	if err != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		return err
	}
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
