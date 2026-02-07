package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"strings"
	"time"

	"github.com/Bunny3th/easy-workflow/workflow/model"
	engineSvc "github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/order/internal/domain"
	"github.com/Duke1616/ecmdb/internal/order/internal/errs"
	"github.com/Duke1616/ecmdb/internal/order/internal/event"
	"github.com/Duke1616/ecmdb/internal/order/internal/service"
	"github.com/Duke1616/ecmdb/internal/pkg/rule"
	templateSvc "github.com/Duke1616/ecmdb/internal/template"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/Duke1616/enotify/notify"
	"github.com/Duke1616/enotify/notify/feishu"
	"github.com/Duke1616/enotify/notify/feishu/card"
	"github.com/Duke1616/enotify/template"
	"github.com/chromedp/chromedp"
	"github.com/ecodeclub/ekit/slice"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/spf13/viper"
	"golang.org/x/image/draw"

	"github.com/ecodeclub/mq-api"
	"github.com/gotomicro/ego/core/elog"
	lark "github.com/larksuite/oapi-sdk-go/v3"
)

const (
	LarkCardProgress            = "feishu-card-progress"
	LarkCardProgressImageResult = "feishu-card-progress-image-result"
)

type LarkCallbackEventConsumer struct {
	Svc              service.Service
	callback         callback
	tmpl             *template.Template
	handler          notify.Handler
	workflowSvc      workflow.Service
	userSvc          user.Service
	engineSvc        engineSvc.Service
	engineProcessSvc service.ProcessEngine
	templateSvc      templateSvc.Service
	consumer         mq.Consumer
	lark             *lark.Client
	logger           *elog.Component
}

func NewLarkCallbackEventConsumer(q mq.MQ, engineSvc engineSvc.Service, engineProcessSvc service.ProcessEngine, service service.Service,
	templateSvc templateSvc.Service, userSvc user.Service, workflowSvc workflow.Service, lark *lark.Client) (*LarkCallbackEventConsumer, error) {
	groupID := "lark_callback"
	consumer, err := q.Consumer(event.LarkCallbackEventName, groupID)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.FromGlobs([]string{})
	if err != nil {
		return nil, err
	}

	handler, err := feishu.NewHandler(lark)
	if err != nil {
		return nil, err
	}

	return &LarkCallbackEventConsumer{
		consumer:         consumer,
		engineSvc:        engineSvc,
		engineProcessSvc: engineProcessSvc,
		userSvc:          userSvc,
		workflowSvc:      workflowSvc,
		handler:          handler,
		callback:         getLarkCallbackConfig(),
		templateSvc:      templateSvc,
		Svc:              service,
		tmpl:             tmpl,
		lark:             lark,
		logger:           elog.DefaultLogger,
	}, nil
}

type callback struct {
	FrontendUrl string `mapstructure:"frontend_url"`
	Debug       bool   `mapstructure:"debug"`
}

func getLarkCallbackConfig() callback {
	var cfg callback
	if err := viper.UnmarshalKey("lark.callback", &cfg); err != nil {
		panic(fmt.Errorf("unable to decode into structure: %v", err))
	}

	return cfg
}

func (c *LarkCallbackEventConsumer) Start(ctx context.Context) {
	go func() {
		for {
			err := c.Consume(ctx)
			if err != nil {
				c.logger.Error("同步飞书回调事件失败", elog.Any("err", err))
			}
		}
	}()
}

func (c *LarkCallbackEventConsumer) Consume(ctx context.Context) error {
	cm, err := c.consumer.Consume(ctx)
	if err != nil {
		return fmt.Errorf("获取消息失败: %w", err)
	}
	var evt event.LarkCallback
	if err = json.Unmarshal(cm.Value, &evt); err != nil {
		return fmt.Errorf("解析消息失败: %w", err)
	}

	// 获取任务ID
	taskId, err := evt.GetTaskIdInt()
	if err != nil {
		return err
	}

	// 获取工单ID
	orderId, err := evt.GetOrderIdInt()
	if err != nil {
		return err
	}

	// 处理消息
	comment := evt.GetComment()
	if comment == "" {
		comment = "无"
	}

	c.logger.Debug("获取飞书回调信息", elog.Any("evt", evt),
		elog.Any("order_id", orderId),
		elog.Any("task_id", taskId),
	)

	var remark string
	switch evt.GetAction() {
	case event.Pass:
		remark = fmt.Sprintf("你已同意该申请, 批注：%s", comment)
		if err = c.engineProcessSvc.Pass(ctx, taskId, comment, evt.GetFormValue()); err != nil {
			if strings.Contains(err.Error(), "已处理，无需操作") {
				remark = "你的节点任务已经结束，无法进行审批，详情登录 ECMDB 平台查看"
				c.logger.Error("飞书回调消息，同意工单失败", elog.FieldErr(err))
				break
			}

			if errors.Is(err, errs.ValidationError) {
				content := fmt.Sprintf(`{"text": "%s"}`, err.Error())
				msg := feishu.NewCreateBuilder(evt.UserId).SetReceiveIDType(feishu.ReceiveIDTypeUserID).
					SetContent(feishu.NewFeishuCustom("text", content)).Build()

				if err = c.handler.Send(ctx, msg); err != nil {
					return fmt.Errorf("触发发送信息失败: %w, 任务ID: %d, 工单ID: %d", err, taskId, orderId)
				}

				return err
			}

			c.logger.Error("飞书回调消息，同意工单失败", elog.FieldErr(err),
				elog.Int("任务ID", taskId),
				elog.Int64("工单ID", orderId),
			)

			return err
		}
	case event.Reject:
		remark = fmt.Sprintf("你已驳回该申请, 批注：%s", comment)
		err = c.engineProcessSvc.Reject(ctx, taskId, comment)
		if err != nil {
			if strings.Contains(err.Error(), "已处理，无需操作") {
				remark = "你的节点任务已经结束，无法进行审批，详情登录 ECMDB 平台查看"
				c.logger.Error("飞书回调消息，同意工单失败", elog.FieldErr(err))
				break
			}

			c.logger.Error("飞书回调消息，驳回工单失败", elog.FieldErr(err),
				elog.String("任务ID", evt.GetTaskId()),
				elog.String("工单ID", evt.GetOrderId()),
			)

			return err
		}
	case event.Progress:
		remark = fmt.Sprintf("你已驳回该申请, 批注：%s", comment)
		err = c.progress(orderId, evt.GetUserId())
		if err != nil {
			c.logger.Error("查看流程进度失败", elog.FieldErr(err))
			return err
		}

		return nil
	case event.Revoke:
		remark = fmt.Sprintf("你已撤销该申请, 批注：%s", evt.GetComment())
		// 获取工单详情
		var orderResp domain.Order
		orderResp, err = c.Svc.Detail(ctx, orderId)
		if err != nil {
			return err
		}

		// 查找用户详情
		var userResp user.User
		userResp, err = c.userSvc.FindByFeishuUserId(ctx, evt.GetUserId())
		if err != nil {
			return err
		}

		// 撤销流程
		err = c.engineProcessSvc.Revoke(ctx, orderResp.Process.InstanceId, userResp.Username, true)
		if err != nil {
			remark = "你的节点任务已经结束，无法进行审批，详情登录 ECMDB 平台查看"
			c.logger.Error("飞书回调消息，驳回工单失败", elog.FieldErr(err))
		}

		err = c.Svc.UpdateStatusByInstanceId(ctx, orderResp.Process.InstanceId, domain.WITHDRAW.ToUint8())
		if err != nil {
			c.logger.Error("撤销变更流程状态失败", elog.FieldErr(err))
		}
	default:
		c.logger.Error("没有匹配到任何选项")
		return nil
	}

	return c.withdraw(ctx, evt, remark)
}

func (c *LarkCallbackEventConsumer) progress(orderId int64, userId string) error {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoSandbox,
		chromedp.NoFirstRun,
		chromedp.DisableGPU,
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("font-render-hinting", "none"),
		chromedp.Flag("force-color-profile", "srgb"),
	)

	if !c.callback.Debug {
		opts = append(opts, chromedp.Headless)
	} else {
		opts = append(opts, chromedp.Flag("headless", false))
	}

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancelBrowser()

	taskCtx, cancelTask := context.WithTimeout(browserCtx, 30*time.Second)
	defer cancelTask()
	// 存储截图的 buffer
	var buf []byte

	// 1. 获取工单详情
	orderDetail, err := c.Svc.Detail(taskCtx, orderId)
	if err != nil {
		return err
	}

	// 2. 获取流程实例详情，拿到对应的版本号
	inst, err := c.engineSvc.GetInstanceByID(taskCtx, orderDetail.Process.InstanceId)
	if err != nil {
		return err
	}

	// 3. 尝试获取历史快照 (Version-Aware)
	wf, err := c.workflowSvc.FindInstanceFlow(taskCtx, orderDetail.WorkflowId, inst.ProcID, inst.ProcVersion)
	if err != nil {
		return err
	}

	// 解析连接线、SRC => DST、标注为通过
	edges, approvalUsers, err := c.parserEdges(taskCtx, orderDetail, wf.ProcessId)
	if err != nil {
		return err
	}

	// 获取代码
	injectData, err := c.getJsCode(wf, edges)
	if err != nil {
		return err
	}

	// 进行截图
	err = chromedp.Run(taskCtx,
		chromedp.EmulateViewport(1920, 1080, chromedp.EmulateScale(1)),
		chromedp.Navigate(c.callback.FrontendUrl),
		chromedp.WaitReady("body"),
		// 等待数据加载
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("Page loaded, waiting for LF-preview...")
			return nil
		}),
		// 注入数据
		chromedp.Evaluate(injectData, nil),

		// 等待 LogicFlow 容器可见
		chromedp.WaitVisible("#LF-preview", chromedp.ByID),

		// 等待前端设置的 data-rendered 标志
		chromedp.WaitVisible(`#LF-preview[data-rendered="true"]`, chromedp.ByQuery),

		// 再等 300ms，防止动画残影
		chromedp.Sleep(300*time.Millisecond),

		// 准备开始截图
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("LF-preview is visible, capturing screenshot...")
			return nil
		}),
		// 截取容器截图（比全屏更精准）
		chromedp.Screenshot("#LF-preview", &buf, chromedp.NodeVisible, chromedp.ByID),
	)
	if err != nil {
		return err
	}

	// 上传文件到飞书
	imageKey, err := c.uploadImage(taskCtx, buf)
	if err != nil {
		return err
	}

	//发送图片消息
	return c.sendImage(taskCtx, imageKey, approvalUsers, userId)
}

func (c *LarkCallbackEventConsumer) parserEdges(ctx context.Context, o domain.Order,
	processId int) (map[string][]string, []string, error) {
	// 查看审批记录（用于获取当前审批人）
	record, _, err := c.engineSvc.TaskRecord(ctx, o.Process.InstanceId, 0, 20)
	if err != nil {
		return nil, nil, err
	}

	users := slice.FilterMap(record, func(idx int, src model.Task) (string, bool) {
		if src.Status == 0 && src.IsFinished == 0 {
			return src.UserID, true
		}

		return "", false
	})

	// 使用 Engine Service 获取已遍历的边
	edges, err := c.engineSvc.GetTraversedEdges(ctx, o.Process.InstanceId, processId, o.Status.ToUint8())
	if err != nil {
		return nil, nil, err
	}

	return edges, users, nil
}

func (c *LarkCallbackEventConsumer) sendImage(ctx context.Context, imageKey *string, approvalUsers []string, userId string) error {
	var fields []card.Field
	us, err := c.userSvc.FindByUsernames(ctx, approvalUsers)
	if err != nil {
		return err
	}

	approval := `**审批人：Null**`
	status := `**状态：<font color='green'> 已结束 </font>**`
	var atTags []string
	if len(us) > 0 {
		for _, u := range us {
			atTags = append(atTags, fmt.Sprintf("<at id=%s></at>", u.FeishuInfo.UserId))
		}

		approval = fmt.Sprintf("**审批人：** %s", strings.Join(atTags, " "))
		status = `**状态：<font color='green'> 审批中 </font>**`
	}

	fields = append(fields, card.Field{
		Tag:     "markdown",
		Content: approval,
	})

	fields = append(fields, card.Field{
		Tag:     "markdown",
		Content: status,
	})

	msg := feishu.NewCreateBuilder(userId).SetReceiveIDType(feishu.ReceiveIDTypeUserID).
		SetContent(feishu.NewFeishuCustomCard(c.tmpl, LarkCardProgressImageResult, card.NewApprovalCardBuilder().
			SetToTitle("工单流程进度查看").
			SetImageKey(*imageKey).
			SetToFields(fields).
			Build())).
		Build()

	if err = c.handler.Send(ctx, msg); err != nil {
		return fmt.Errorf("发送流程进度失败: %w", err)
	}

	return nil
}

func (c *LarkCallbackEventConsumer) uploadImage(ctx context.Context, buf []byte) (*string, error) {
	req := larkim.NewCreateImageReqBuilder().
		Body(larkim.NewCreateImageReqBodyBuilder().
			ImageType(`message`).
			Image(bytes.NewReader(buf)).
			Build()).
		Build()

	// 发起请求
	resp, err := c.lark.Im.Image.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	// 服务端错误处理
	if !resp.Success() {
		return nil, fmt.Errorf("logId: %s, error response: \n%s",
			resp.RequestId(), larkcore.Prettify(resp.CodeError))
	}

	return resp.Data.ImageKey, nil
}

func (c *LarkCallbackEventConsumer) getJsCode(wf workflow.Workflow, edgeMap map[string][]string) (string, error) {
	edgesJSON, err := json.Marshal(wf.FlowData.Edges)
	if err != nil {
		return "", err
	}

	var edges []easyflow.Edge
	err = json.Unmarshal(edgesJSON, &edges)
	if err != nil {
		return "", err
	}

	for i, edge := range edges {
		properties, ok := edge.Properties.(map[string]interface{})
		if !ok {
			properties = make(map[string]interface{})
		}

		targetNodeIDs, ok := edgeMap[edge.SourceNodeId]
		if !ok {
			continue
		}

		// 检查当前边的 Target 是否在 targets 列表中
		isMatched := false
		for _, tid := range targetNodeIDs {
			if tid == edge.TargetNodeId {
				isMatched = true
				break
			}
		}

		if !isMatched {
			continue
		}

		// 只更新 IsPass 字段，保留其他字段
		properties["is_pass"] = true
		edges[i].Properties = properties
	}

	var edgesMap []map[string]interface{}
	for _, edge := range edges {
		edgesMap = append(edgesMap, edgeToMap(edge))
	}

	wf.FlowData.Edges = edgesMap
	easyFlowData, err := json.Marshal(wf.FlowData)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`window.__DATA__ = %s;`, easyFlowData), nil
}

func edgeToMap(edge easyflow.Edge) map[string]interface{} {
	return map[string]interface{}{
		"type":         edge.Type,
		"sourceNodeId": edge.SourceNodeId,
		"targetNodeId": edge.TargetNodeId,
		"properties":   edge.Properties,
		"id":           edge.ID,
		"startPoint":   edge.StartPoint,
		"endPoint":     edge.EndPoint,
		"pointsList":   edge.PointsList,
		"text":         edge.Text,
	}
}

func (c *LarkCallbackEventConsumer) withdraw(ctx context.Context, callback event.LarkCallback, wantResult string) error {
	// 获取模版详情信息
	orderIdInt, err := callback.GetOrderIdInt()
	if err != nil {
		return err
	}

	fOrder, err := c.Svc.Detail(ctx, orderIdInt)
	if err != nil {
		return err
	}

	t, err := c.templateSvc.DetailTemplate(ctx, fOrder.TemplateId)
	if err != nil {
		return err
	}

	rules, err := rule.ParseRules(t.Rules)
	if err != nil {
		return err
	}
	ruleFields := rule.GetFields(rules, fOrder.Provide.ToUint8(), fOrder.Data)
	userInfo, err := c.userSvc.FindByUsername(ctx, fOrder.CreateBy)
	if err != nil {
		return err
	}

	msg := feishu.NewPatchBuilder(callback.GetMessageId()).SetReceiveIDType(feishu.ReceiveIDTypeUserID).
		SetContent(feishu.NewFeishuCustomCard(c.tmpl, LarkCardProgress,
			card.NewApprovalCardBuilder().
				SetToTitle(rule.GenerateTitle(userInfo.DisplayName, t.Name)).
				SetToFields(toCardFields(ruleFields)).
				SetToCallbackValue(getCallbackValue(callback)).
				SetWantResult(wantResult).
				Build(),
		)).Build()

	if err = c.handler.Send(ctx, msg); err != nil {
		return fmt.Errorf("修改飞书消息失败: %w", err)
	}

	return nil
}

func getCallbackValue(callback event.LarkCallback) []card.Value {
	fields := []struct {
		Key   string
		Value string
	}{
		{"order_id", callback.GetOrderId()},
		{"task_id", callback.GetTaskId()},
		{"user_id", callback.GetUserId()},
		{"action", string(event.Progress)},
	}

	value := make([]card.Value, 0, len(fields))
	for _, field := range fields {
		value = append(value, card.Value{
			Key:   field.Key,
			Value: field.Value,
		})
	}

	return value
}

func (c *LarkCallbackEventConsumer) Stop(_ context.Context) error {
	return c.consumer.Close()
}

func scaleImage(buf []byte) ([]byte, error) {
	_, format, err := image.DecodeConfig(bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}

	if format != "jpeg" && format != "png" {
		return nil, fmt.Errorf("unsupported image format: %s", format)
	}

	// 解码截图
	img, _, err := image.Decode(bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}

	// 缩放比例
	scale := 0.5
	dst := image.NewRGBA(image.Rect(0, 0, int(float64(img.Bounds().Dx())*scale), int(float64(img.Bounds().Dy())*scale)))

	// 使用双线性插值缩放图片
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

	// 将处理后的图片编码为 []byte
	var outBuf bytes.Buffer
	if err = jpeg.Encode(&outBuf, dst, nil); err != nil { // 使用 jpeg 编码
		return nil, err
	}

	// 返回处理后的字节流
	return outBuf.Bytes(), nil
}

func toCardFields(fields []rule.Field) []card.Field {
	var cardFields []card.Field
	for _, f := range fields {
		cardFields = append(cardFields, card.Field{
			IsShort: f.IsShort,
			Tag:     f.Tag,
			Content: f.Content,
		})
	}
	return cardFields
}
