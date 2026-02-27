package domain

type TriggerPosition string

// ToString 转换为字符串
func (t TriggerPosition) ToString() string {
	return string(t)
}

const (
	// 正常流转与调度相关
	TriggerPositionTaskWaiting          TriggerPosition = "任务等待"
	TriggerPositionReadyToStartNode     TriggerPosition = "准备启动节点"
	TriggerPositionDispatchDelivered    TriggerPosition = "分发已送达执行端，当前任务执行中"
	TriggerPositionTaskExecutionSuccess TriggerPosition = "任务执行成功"
	TriggerPositionTaskExecutionFailed  TriggerPosition = "任务执行失败"

	// 重试与恢复相关
	TriggerPositionManualRetry            TriggerPosition = "人工手动重试"
	TriggerPositionAutoRetry              TriggerPosition = "自动补发任务"
	TriggerPositionAutoRetryLimitExceeded TriggerPosition = "超过最大重试次数"
	TriggerPositionManualSuccess          TriggerPosition = "手动修改状态为成功"

	// 错误处理相关
	TriggerPositionErrorGetOrder              TriggerPosition = "获取工单失败"
	TriggerPositionErrorGetProcessInst        TriggerPosition = "获取流程实例失败"
	TriggerPositionErrorGetProcessInfo        TriggerPosition = "获取流程信息失败"
	TriggerPositionErrorExtractAutomationInfo TriggerPosition = "提取自动化信息失败"
	TriggerPositionErrorGetDispatcherNode     TriggerPosition = "获取调度节点失败"
	TriggerPositionErrorGetTaskTemplate       TriggerPosition = "获取任务模版失败"
)
