package template

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	notificationv1 "github.com/Duke1616/ecmdb/api/proto/gen/ealert/notification/v1"
	templatev1 "github.com/Duke1616/ecmdb/api/proto/gen/ealert/template/v1"
	"github.com/Duke1616/ecmdb/cmd/initial/ioc"
	"github.com/Duke1616/ecmdb/internal/workflow"
)

//go:embed fs/approval.tmpl
var LarkApprovalTemplate string

//go:embed fs/carbon_copy.tmpl
var LarkApprovalCCTemplate string

//go:embed fs/progress.tmpl
var LarkApprovalProgressTemplate string

//go:embed fs/progress_image_result.tmpl
var LarkApprovalProgressImageResultTemplate string

//go:embed fs/revoke.tmpl
var LarkApprovalRevokeTemplate string

type InitialNotifyTemplate struct {
	App *ioc.App
}

func NewInitial(app *ioc.App) *InitialNotifyTemplate {
	return &InitialNotifyTemplate{
		App: app,
	}
}

type templateConfig struct {
	Name        string
	Desc        string
	Channel     notificationv1.Channel
	VersionName string
	Content     string
	NotifyType  workflow.NotifyType
}

func (i *InitialNotifyTemplate) InitTemplate() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	fmt.Printf("📋 开始初始化工单消息通知模版...\n")

	templates := []templateConfig{
		{
			Name:        "工单审批通知",
			Desc:        "工单审批流程通知",
			Channel:     notificationv1.Channel_FEISHU_CARD,
			VersionName: "v1.0.0",
			Content:     LarkApprovalTemplate,
			NotifyType:  workflow.NotifyTypeApproval,
		},
		{
			Name:        "工单抄送通知",
			Desc:        "工单抄送通知",
			Channel:     notificationv1.Channel_FEISHU_CARD,
			VersionName: "v1.0.0",
			Content:     LarkApprovalCCTemplate,
			NotifyType:  workflow.NotifyTypeCC,
		},
		{
			Name:        "工单进度通知",
			Desc:        "工单进度通知",
			Channel:     notificationv1.Channel_FEISHU_CARD,
			VersionName: "v1.0.0",
			Content:     LarkApprovalProgressTemplate,
			NotifyType:  workflow.NotifyTypeProgress,
		},
		{
			Name:        "工单进度图片通知",
			Desc:        "工单进度图片结果通知",
			Channel:     notificationv1.Channel_FEISHU_CARD,
			VersionName: "v1.0.0",
			Content:     LarkApprovalProgressImageResultTemplate,
			NotifyType:  workflow.NotifyTypeProgressImageResult,
		},
		{
			Name:        "工单撤回通知",
			Desc:        "工单撤回通知",
			Channel:     notificationv1.Channel_FEISHU_CARD,
			VersionName: "v1.0.0",
			Content:     LarkApprovalRevokeTemplate,
			NotifyType:  workflow.NotifyTypeRevoke,
		},
	}

	for _, t := range templates {
		if err := i.processTemplate(ctx, t); err != nil {
			// 错误已在内部处理并打印，继续下一个
			continue
		}
	}

	fmt.Printf("✅ 所有工单模版处理完成!\n")
	return nil
}

// checkTemplateExists 检查模板是否已存在
// 返回: templateId, exists, error
func (i *InitialNotifyTemplate) checkTemplateExists(ctx context.Context, notifyType workflow.NotifyType, channel string) (int64, bool, error) {
	binding, err := i.App.WorkflowSvc.AdminNotifyBinding().GetEffective(ctx, 0, notifyType, channel)
	if err != nil || binding.Id == 0 {
		return 0, false, nil
	}
	return binding.TemplateId, true, nil
}

// createAndPublishTemplate 创建并发布模板
// 返回: templateId, versionId, error
func (i *InitialNotifyTemplate) createAndPublishTemplate(ctx context.Context, cfg templateConfig) (int64, int64, error) {
	// 创建模板
	createResp, err := i.App.TemplateClient.CreateTemplate(ctx, &templatev1.CreateTemplateRequest{
		Template: &templatev1.ChannelTemplate{
			Name:        cfg.Name,
			Description: cfg.Desc,
			Channel:     cfg.Channel,
			Versions: []*templatev1.ChannelTemplateVersion{
				{
					Name:    cfg.VersionName,
					Content: cfg.Content,
					Desc:    "初始化版本",
				},
			},
		},
	})
	if err != nil {
		return 0, 0, fmt.Errorf("创建模版失败: %w", err)
	}

	templateId := createResp.Template.Id
	if len(createResp.Template.Versions) == 0 {
		return 0, 0, fmt.Errorf("创建模版成功，但没有返回版本信息")
	}
	versionId := createResp.Template.Versions[0].Id

	// 发布版本
	_, err = i.App.TemplateClient.PublishTemplate(ctx, &templatev1.PublishTemplateRequest{
		TemplateId: templateId,
		VersionId:  versionId,
	})
	if err != nil {
		return 0, 0, fmt.Errorf("发布模版版本失败: %w", err)
	}

	return templateId, versionId, nil
}

// createNotifyBinding 创建通知绑定
func (i *InitialNotifyTemplate) createNotifyBinding(ctx context.Context, notifyType workflow.NotifyType, channel string, templateId int64) error {
	_, err := i.App.WorkflowSvc.AdminNotifyBinding().Create(ctx, workflow.NotifyBinding{
		WorkflowId: 0, // 0 表示默认全局配置
		NotifyType: notifyType,
		Channel:    channel,
		TemplateId: templateId,
	})
	return err
}

// createNewVersionAndPublish 为已存在的模板创建新版本并发布
// 返回: versionId, error
func (i *InitialNotifyTemplate) createNewVersionAndPublish(ctx context.Context, templateId int64, cfg templateConfig) (int64, error) {
	// 创建新版本
	createVerResp, err := i.App.TemplateClient.CreateTemplateVersion(ctx, &templatev1.CreateTemplateVersionRequest{
		TemplateId: templateId,
		Name:       cfg.VersionName,
		Content:    cfg.Content,
		Desc:       "初始化更新版本",
	})
	if err != nil {
		return 0, fmt.Errorf("创建新版本失败: %w", err)
	}

	// 获取新创建的版本ID
	if createVerResp.Version == nil {
		return 0, fmt.Errorf("创建新版本成功，但没有返回版本信息")
	}
	versionId := createVerResp.Version.Id

	// 发布新版本
	_, err = i.App.TemplateClient.PublishTemplate(ctx, &templatev1.PublishTemplateRequest{
		TemplateId: templateId,
		VersionId:  versionId,
	})
	if err != nil {
		return 0, fmt.Errorf("发布新版本失败: %w", err)
	}

	return versionId, nil
}

// processTemplate 处理单个模板的初始化流程
func (i *InitialNotifyTemplate) processTemplate(ctx context.Context, cfg templateConfig) error {
	fmt.Printf("🔄正在处理模版: %s\n", cfg.Name)

	// 1. 检查是否已存在
	existingTemplateId, exists, err := i.checkTemplateExists(ctx, cfg.NotifyType, cfg.Channel.String())
	if err != nil {
		fmt.Printf("❌ 检查模版 [%s] 失败: %v\n", cfg.Name, err)
		return err
	}

	var templateId int64
	var versionId int64

	if exists {
		// 模板已存在,创建新版本并发布
		templateId = existingTemplateId
		versionId, err = i.createNewVersionAndPublish(ctx, templateId, cfg)
		if err != nil {
			fmt.Printf("❌ 模版 [%s] %v\n", cfg.Name, err)
			return err
		}
		fmt.Printf("✅ 模版 [%s] 已更新新版本并发布 (TemplateID: %d, VersionID: %d)\n", cfg.Name, templateId, versionId)
		return nil
	}

	// 2. 创建并发布模板
	templateId, versionId, err = i.createAndPublishTemplate(ctx, cfg)
	if err != nil {
		fmt.Printf("❌ 模版 [%s] %v\n", cfg.Name, err)
		return err
	}
	fmt.Printf("✅ 模版 [%s] 创建并发布成功 (TemplateID: %d, VersionID: %d)\n", cfg.Name, templateId, versionId)

	// 3. 创建 NotifyBinding
	if err = i.createNotifyBinding(ctx, cfg.NotifyType, cfg.Channel.String(), templateId); err != nil {
		fmt.Printf("❌ 创建 NotifyBinding [%s] 失败: %v\n", cfg.Name, err)
		return err
	}
	fmt.Printf("✅ NotifyBinding [%s] 创建成功\n", cfg.Name)

	return nil
}
