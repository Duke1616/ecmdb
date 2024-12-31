package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	engineSvc "github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/order/internal/event"
	"github.com/Duke1616/ecmdb/internal/order/internal/service"
	"github.com/Duke1616/ecmdb/internal/pkg/rule"
	templateSvc "github.com/Duke1616/ecmdb/internal/template"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/enotify/notify"
	"github.com/Duke1616/enotify/notify/feishu"
	"github.com/Duke1616/enotify/notify/feishu/card"
	feishuMsg "github.com/Duke1616/enotify/notify/feishu/message"
	"github.com/Duke1616/enotify/template"
	"github.com/chromedp/chromedp"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/spf13/viper"
	"log"
	"time"

	"github.com/ecodeclub/mq-api"
	"github.com/gotomicro/ego/core/elog"
	"github.com/larksuite/oapi-sdk-go/v3"
	"strconv"
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
		err = c.progress(ctx, orderId, evt.FeishuUserId)
		if err != nil {
			c.logger.Error("查看流程进度失败", elog.FieldErr(err))
			return err
		}

		return nil

	default:
		return nil
	}

	return c.withdraw(ctx, evt, wantResult)
}

func (c *FeishuCallbackEventConsumer) progress(gCtx context.Context, orderId int64, userId string) error {
	ctx, cancel := chromedp.NewContext(gCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	// 设置超时时间
	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 存储截图的 buffer
	var buf []byte

	// 获取代码
	jsCode, err := c.getJsCode(ctx, orderId)
	if err != nil {
		return err
	}

	// 进行截图
	err = chromedp.Run(ctx,
		chromedp.Navigate(c.logicFlowUrl),
		chromedp.WaitReady("body"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("Page loaded, waiting for LF-preview...")
			return nil
		}),
		chromedp.Evaluate(jsCode, nil),
		chromedp.WaitVisible("#LF-preview", chromedp.ByID),
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("LF-preview is visible, capturing screenshot...")
			return nil
		}),
		chromedp.FullScreenshot(&buf, 2000),
	)
	if err != nil {
		return err
	}

	// 上传文件到飞书
	imageKey, err := c.uploadImage(ctx, buf)
	if err != nil {
		return err
	}

	fmt.Println(*imageKey, "imagekey")

	// 发送图片消息
	return c.sendImage(ctx, imageKey, userId)
}

func (c *FeishuCallbackEventConsumer) sendImage(ctx context.Context, imageKey *string, userId string) error {
	var fields []card.Field
	fields = append(fields, card.Field{
		Tag:     "markdown",
		Content: `**审批人：** <at id=></at>`,
	})

	fields = append(fields, card.Field{
		Tag:     "markdown",
		Content: `**状态：<font color='green'> 审批中 </font>**`,
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

func (c *FeishuCallbackEventConsumer) getJsCode(ctx context.Context, orderId int64) (string, error) {
	o, err := c.Svc.Detail(ctx, orderId)
	if err != nil {
		return "", err
	}

	wf, err := c.workflowSvc.Find(ctx, o.WorkflowId)
	if err != nil {
		return "", err
	}

	easyFlowData, err := json.Marshal(wf.FlowData)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`window.__DATA__ = %s;`, easyFlowData), nil
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
