package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"strings"
	"time"

	"github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	engineSvc "github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/order/internal/domain"
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
	feishuMsg "github.com/Duke1616/enotify/notify/feishu/message"
	"github.com/Duke1616/enotify/template"
	"github.com/chromedp/chromedp"
	"github.com/ecodeclub/ekit/slice"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/spf13/viper"
	"golang.org/x/image/draw"

	"strconv"

	"github.com/ecodeclub/mq-api"
	"github.com/gotomicro/ego/core/elog"
	"github.com/larksuite/oapi-sdk-go/v3"
)

const (
	FeishuCardProgress            = "feishu-card-progress"
	FeishuCardProgressImageResult = "feishu-card-progress-image-result"
)

type FeishuCallbackEventConsumer struct {
	patchNc      notify.Notifier[*larkim.PatchMessageReq]
	createNc     notify.Notifier[*larkim.CreateMessageReq]
	Svc          service.Service
	logicFlowUrl string
	tmpl         *template.Template
	workflowSvc  workflow.Service
	userSvc      user.Service
	engineSvc    engineSvc.Service
	templateSvc  templateSvc.Service
	consumer     mq.Consumer
	lark         *lark.Client
	logger       *elog.Component
}

func NewFeishuCallbackEventConsumer(q mq.MQ, engineSvc engineSvc.Service, service service.Service,
	templateSvc templateSvc.Service, userSvc user.Service, workflowSvc workflow.Service, lark *lark.Client) (*FeishuCallbackEventConsumer, error) {
	groupID := "feishu_callback"
	consumer, err := q.Consumer(event.FeishuCallbackEventName, groupID)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.FromGlobs([]string{})
	if err != nil {
		return nil, err
	}

	patchNc, err := feishu.NewPatchFeishuNotifyByClient(lark)
	if err != nil {
		return nil, err
	}

	createNc, err := feishu.NewCreateFeishuNotifyByClient(lark)
	if err != nil {
		return nil, err
	}

	return &FeishuCallbackEventConsumer{
		consumer:     consumer,
		engineSvc:    engineSvc,
		userSvc:      userSvc,
		workflowSvc:  workflowSvc,
		patchNc:      patchNc,
		createNc:     createNc,
		logicFlowUrl: getLogicFlowUrl(),
		templateSvc:  templateSvc,
		Svc:          service,
		tmpl:         tmpl,
		lark:         lark,
		logger:       elog.DefaultLogger,
	}, nil
}

func getLogicFlowUrl() string {
	type Config struct {
		LogicFlowUrl string `mapstructure:"logicflow_url"`
	}

	var cfg Config
	if err := viper.UnmarshalKey("frontend", &cfg); err != nil {
		panic(fmt.Errorf("unable to decode into struct: %v", err))
	}

	return cfg.LogicFlowUrl
}

func (c *FeishuCallbackEventConsumer) Start(ctx context.Context) {
	go func() {
		for {
			err := c.Consume(ctx)
			if err != nil {
				c.logger.Error("同步飞书回调事件失败", elog.Any("err", err))
			}
		}
	}()
}

func (c *FeishuCallbackEventConsumer) Consume(ctx context.Context) error {
	cm, err := c.consumer.Consume(ctx)
	if err != nil {
		return fmt.Errorf("获取消息失败: %w", err)
	}
	var evt event.FeishuCallback
	if err = json.Unmarshal(cm.Value, &evt); err != nil {
		return fmt.Errorf("解析消息失败: %w", err)
	}

	taskId, err := strconv.Atoi(evt.TaskId)
	if err != nil {
		return err
	}

	// 处理消息
	if evt.Comment == "" {
		evt.Comment = "无"
	}

	var wantResult string
	switch evt.Action {
	case "pass":
		wantResult = fmt.Sprintf("你已同意该申请, 批注：%s", evt.Comment)
		err = c.engineSvc.Pass(ctx, taskId, evt.Comment)
		if err != nil {
			wantResult = "你的节点任务已经结束，无法进行审批，详情登录 ECMDB 平台查看"
			c.logger.Error("飞书回调消息，同意工单失败", elog.FieldErr(err))
		}
	case "reject":
		wantResult = fmt.Sprintf("你已驳回该申请, 批注：%s", evt.Comment)
		err = c.engineSvc.Reject(ctx, taskId, evt.Comment)
		if err != nil {
			wantResult = "你的节点任务已经结束，无法进行审批，详情登录 ECMDB 平台查看"
			c.logger.Error("飞书回调消息，驳回工单失败", elog.FieldErr(err))
		}
	case "progress":
		wantResult = fmt.Sprintf("你已驳回该申请, 批注：%s", evt.Comment)
		var orderId int64
		orderId, err = strconv.ParseInt(evt.OrderId, 10, 64)
		if err != nil {
			c.logger.Error("查看流程进度失败", elog.FieldErr(err))
			return err
		}
		err = c.progress(orderId, evt.FeishuUserId)
		if err != nil {
			c.logger.Error("查看流程进度失败", elog.FieldErr(err))
			return err
		}

		return nil
	case "revoke":
		wantResult = fmt.Sprintf("你已撤销该申请, 批注：%s", evt.Comment)
		var orderId int64
		orderId, err = strconv.ParseInt(evt.OrderId, 10, 64)
		if err != nil {
			c.logger.Error("查看流程进度失败", elog.FieldErr(err))
			return err
		}

		// 获取工单详情
		var orderResp domain.Order
		orderResp, err = c.Svc.Detail(ctx, orderId)
		if err != nil {
			return err
		}

		// 查找用户详情
		var userResp user.User
		userResp, err = c.userSvc.FindByFeishuUserId(ctx, evt.FeishuUserId)
		if err != nil {
			return err
		}

		// 撤销流程
		err = c.engineSvc.Revoke(ctx, orderResp.Process.InstanceId, userResp.Username, true)
		if err != nil {
			wantResult = "你的节点任务已经结束，无法进行审批，详情登录 ECMDB 平台查看"
			c.logger.Error("飞书回调消息，驳回工单失败", elog.FieldErr(err))
		}

		err = c.Svc.UpdateStatusByInstanceId(ctx, orderResp.Process.InstanceId, domain.WITHDRAW.ToUint8())
		if err != nil {
			c.logger.Error("撤销变更流程状态失败", elog.FieldErr(err))
		}
	default:
		return nil
	}

	return c.withdraw(ctx, evt, wantResult)
}

func (c *FeishuCallbackEventConsumer) progress(orderId int64, userId string) error {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Headless,
		chromedp.NoSandbox,
		chromedp.NoFirstRun,
		chromedp.DisableGPU,
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("single-process", true),
		chromedp.Flag("no-zygote", true),
		chromedp.Flag("font-render-hinting", "none"),
		chromedp.Flag("force-color-profile", "srgb"),
	)

	ctx, cancel = chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx, chromedp.WithLogf(log.Printf))
	defer cancel()

	// 设置超时时间
	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 存储截图的 buffer
	var buf []byte

	orderDetail, err := c.Svc.Detail(ctx, orderId)
	if err != nil {
		return err
	}

	wf, err := c.workflowSvc.Find(ctx, orderDetail.WorkflowId)
	if err != nil {
		return err
	}

	// 解析连接线、SRC => DST、标注为通过
	edges, approvalUsers, err := c.parserEdges(ctx, orderDetail, wf.ProcessId)
	if err != nil {
		return err
	}

	// 获取代码
	jsCode, err := c.getJsCode(wf, edges)
	if err != nil {
		return err
	}

	// 进行截图
	err = chromedp.Run(ctx,
		chromedp.EmulateViewport(1920, 1080, chromedp.EmulateScale(1)),
		chromedp.Navigate(c.logicFlowUrl),
		chromedp.WaitReady("body"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("Page loaded, waiting for LF-preview...")
			return nil
		}),
		chromedp.Evaluate(jsCode, nil),
		chromedp.WaitVisible("#LF-preview", chromedp.ByID),
		chromedp.Sleep(1*time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("LF-preview is visible, capturing screenshot...")
			return nil
		}),
		chromedp.FullScreenshot(&buf, 100),
	)
	if err != nil {
		return err
	}

	// 上传文件到飞书
	imageKey, err := c.uploadImage(ctx, buf)
	if err != nil {
		return err
	}

	// 发送图片消息
	return c.sendImage(ctx, imageKey, approvalUsers, userId)
}

func (c *FeishuCallbackEventConsumer) parserEdges(ctx context.Context, o domain.Order,
	processId int) (map[string]string, []string, error) {
	// 查看审批记录
	record, _, err := c.engineSvc.TaskRecord(ctx, o.Process.InstanceId, 0, 20)
	if err != nil {
		return nil, nil, err
	}

	// 获取当前所处的节点 ID
	currentNode := record[len(record)-1].NodeID

	users := slice.FilterMap(record, func(idx int, src model.Task) (string, bool) {
		if src.Status == 0 && src.IsFinished == 0 {
			return src.UserID, true
		}

		return "", false
	})

	// 查看流程定义及节点
	define, err := engine.GetProcessDefine(processId)
	if err != nil {
		return nil, nil, err
	}
	var endNodeId string
	nodesMap := slice.ToMap(define.Nodes, func(element model.Node) string {
		if element.NodeType == model.EndNode {
			endNodeId = element.NodeID
		}
		return element.NodeID
	})

	// 如果工单已经结束，则以结束节点为查询起点
	if o.Status == domain.END {
		currentNode = endNodeId
	}

	// 过滤 record 只保留正常节点，驳回节点暂不处理
	filterRecord := slice.FilterMap(record, func(idx int, src model.Task) (model.Task, bool) {
		if src.Status == 2 {
			return model.Task{}, false
		}
		return src, true
	})
	recordMap := slice.ToMap(filterRecord, func(element model.Task) string {
		return element.NodeID
	})

	edges := make(map[string]string)
	visited := make(map[string]bool)

	processNode(currentNode, nodesMap, recordMap, edges, visited)
	return edges, users, nil
}

// 定义一个递归函数来处理节点
func processNode(nodeID string, nodesMap map[string]model.Node,
	recordMap map[string]model.Task, edges map[string]string, visited map[string]bool) {
	if visited[nodeID] {
		return
	}
	visited[nodeID] = true

	// 获取当前节点
	node, exists := nodesMap[nodeID]
	if !exists {
		return
	}

	// 如果上级只有一个节点则直接处理
	if len(node.PrevNodeIDs) == 1 {
		prevNodeID := node.PrevNodeIDs[0]

		// 将前置节点添加到 edges 中
		edges[prevNodeID] = nodeID

		// 递归处理前置节点
		processNode(prevNodeID, nodesMap, recordMap, edges, visited)
	}

	// 处理当前节点的前置节点, 网关节点正常不会同时出现两个
	taskNodes := make([]string, 0)
	var gatewayNode string
	for _, prevNodeID := range node.PrevNodeIDs {
		n, _ := nodesMap[prevNodeID]

		switch n.NodeType {
		case model.TaskNode:
			taskNodes = append(taskNodes, n.NodeID)
		case model.GateWayNode:
			gatewayNode = n.NodeID
		}
	}

	// 优先处理任务节点
	taskNodesProcessed := processTaskNodes(taskNodes, nodeID, nodesMap, recordMap, edges, visited)

	// 如果任务节点处理成功，则不处理网关节点
	if !taskNodesProcessed {
		// 处理网关节点
		processGatewayNode(gatewayNode, nodeID, nodesMap, recordMap, edges, visited)
	}
}

// processTaskNodes 处理任务节点，返回是否成功处理
func processTaskNodes(taskNodes []string, nodeID string, nodesMap map[string]model.Node,
	recordMap map[string]model.Task, edges map[string]string, visited map[string]bool) bool {
	if len(taskNodes) == 0 {
		return false
	}

	for _, taskNodeID := range taskNodes {
		// 检查任务节点是否处理成功
		if _, exists := recordMap[taskNodeID]; !exists {
			continue
		}

		processNode(taskNodeID, nodesMap, recordMap, edges, visited)
		edges[taskNodeID] = nodeID
		return true
	}

	// 如果所有任务节点都未成功，返回 false
	return false
}

func processGatewayNode(gatewayNode string, nodeID string, nodesMap map[string]model.Node,
	recordMap map[string]model.Task, edges map[string]string, visited map[string]bool) {
	if gatewayNode != "" {
		edges[gatewayNode] = nodeID
		processNode(gatewayNode, nodesMap, recordMap, edges, visited)
	}
}

func (c *FeishuCallbackEventConsumer) sendImage(ctx context.Context, imageKey *string, approvalUsers []string, userId string) error {
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

	notifyWrap := notify.WrapNotifierDynamic(c.createNc, func() (notify.BasicNotificationMessage[*larkim.CreateMessageReq], error) {
		return feishuMsg.NewCreateFeishuMessage(
			"user_id", userId,
			feishu.NewFeishuCustomCard(c.tmpl, FeishuCardProgressImageResult,
				card.NewApprovalCardBuilder().
					SetToTitle("工单流程进度查看").
					SetImageKey(*imageKey).
					SetToFields(fields).
					Build(),
			),
		), nil
	})

	ok, err := notifyWrap.Send(ctx)
	if !ok {
		c.logger.Error("发送流程进度失败")
	}
	return err
}

func (c *FeishuCallbackEventConsumer) uploadImage(ctx context.Context, buf []byte) (*string, error) {
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

func (c *FeishuCallbackEventConsumer) getJsCode(wf workflow.Workflow, edgeMap map[string]string) (string, error) {
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

		val, ok := edgeMap[edge.SourceNodeId]
		if !ok {
			continue
		}

		if val != edge.TargetNodeId {
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

func (c *FeishuCallbackEventConsumer) withdraw(ctx context.Context, callback event.FeishuCallback, wantResult string) error {
	// 获取模版详情信息
	orderIdInt, _ := strconv.ParseInt(callback.OrderId, 10, 64)

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
	fields := rule.GetFields(rules, fOrder.Provide.ToUint8(), fOrder.Data)
	userInfo, err := c.userSvc.FindByUsername(ctx, fOrder.CreateBy)
	if err != nil {
		return err
	}

	notifyWrap := notify.WrapNotifierDynamic(c.patchNc, func() (notify.BasicNotificationMessage[*larkim.PatchMessageReq], error) {
		return feishuMsg.NewPatchFeishuMessage(
			callback.MessageId,
			feishu.NewFeishuCustomCard(c.tmpl, FeishuCardProgress,
				card.NewApprovalCardBuilder().
					SetToTitle(rule.GenerateTitle(userInfo.DisplayName, t.Name)).
					SetToFields(fields).
					SetToCallbackValue(getCallbackValue(callback)).
					SetWantResult(wantResult).
					Build(),
			),
		), nil
	})

	ok, err := notifyWrap.Send(ctx)
	if !ok {
		c.logger.Error("修改飞书消息失败")
	}
	return err
}

func getCallbackValue(callback event.FeishuCallback) []card.Value {
	fields := []struct {
		Key   string
		Value string
	}{
		{"order_id", callback.OrderId},
		{"task_id", callback.TaskId},
		{"feishu_user_id", callback.FeishuUserId},
		{"action", "progress"},
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

func (c *FeishuCallbackEventConsumer) Stop(_ context.Context) error {
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
