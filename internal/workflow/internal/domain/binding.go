package domain

type NotifyType string

const (
	// NotifyTypeApproval 审批通知
	NotifyTypeApproval NotifyType = "approval"
	// NotifyTypeCC 抄送通知
	NotifyTypeCC NotifyType = "carbon-copy"
	// NotifyTypeProgress 进度通知
	NotifyTypeProgress NotifyType = "progress"
	// NotifyTypeProgressImageResult 结果通知
	NotifyTypeProgressImageResult NotifyType = "progress-image-result"
	// NotifyTypeRevoke 撤销通知
	NotifyTypeRevoke NotifyType = "revoke"
)

type NotifyBinding struct {
	Id         int64
	WorkflowId int64
	NotifyType NotifyType
	Channel    string
	TemplateId int64
	Ctime      int64
	Utime      int64
}
