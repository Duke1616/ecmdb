package errs

import "errors"

var (
	// ErrNotificationFailed 通知发送失败
	ErrNotificationFailed = errors.New("通知发送失败")

	// ErrBuildMessage 构建消息失败
	ErrBuildMessage = errors.New("构建消息失败")

	// ErrParseMessage 解析消息失败
	ErrParseMessage = errors.New("解析消息失败")

	// ErrNotificationUnavailable 通知服务不可用
	ErrNotificationUnavailable = errors.New("通知服务不可用")

	// ErrNodeNotConfigured 节点未配置通知
	ErrNodeNotConfigured = errors.New("节点未配置通知")

	// ErrFetchData 获取数据失败
	ErrFetchData = errors.New("获取数据失败")

	// ErrResolveRule 解析规则失败
	ErrResolveRule = errors.New("解析规则失败")
)

// NotificationErrorCode 通知错误代码
type NotificationErrorCode string

const (
	// ErrorCodeBuildFailed 消息构建失败
	ErrorCodeBuildFailed NotificationErrorCode = "BUILD_FAILED"

	// ErrorCodeParseFailed 消息解析失败
	ErrorCodeParseFailed NotificationErrorCode = "PARSE_FAILED"

	// ErrorCodeServiceUnavailable 服务不可用
	ErrorCodeServiceUnavailable NotificationErrorCode = "SERVICE_UNAVAILABLE"

	// ErrorCodeTimeout 超时
	ErrorCodeTimeout NotificationErrorCode = "TIMEOUT"

	// ErrorCodeNodeNotConfigured 节点未配置
	ErrorCodeNodeNotConfigured NotificationErrorCode = "NODE_NOT_CONFIGURED"

	// ErrorCodeFetchDataFailed 获取数据失败
	ErrorCodeFetchDataFailed NotificationErrorCode = "FETCH_DATA_FAILED"

	// ErrorCodeResolveRuleFailed 解析规则失败
	ErrorCodeResolveRuleFailed NotificationErrorCode = "RESOLVE_RULE_FAILED"

	// ErrorCodeUnknown 未知错误
	ErrorCodeUnknown NotificationErrorCode = "UNKNOWN"
)

// StatusCode 通知状态代码
type StatusCode string

const (
	// StatusSuccess 成功
	StatusSuccess StatusCode = "SUCCESS"

	// StatusFailed 失败
	StatusFailed StatusCode = "FAILED"

	// StatusPending 待处理
	StatusPending StatusCode = "PENDING"
)
