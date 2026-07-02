package plugin

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Duke1616/ecmdb/internal/domain"
	pluginx "github.com/Duke1616/ecmdb/pkg/plugin"
)

var errRequiredInputMissing = errors.New("plugin required input missing")

type missingInputError struct {
	Reasons []string
}

func (e *missingInputError) Error() string {
	if len(e.Reasons) == 0 {
		return errRequiredInputMissing.Error()
	}
	return fmt.Sprintf("%s: %s", errRequiredInputMissing.Error(), strings.Join(e.Reasons, "; "))
}

func (e *missingInputError) Is(target error) bool {
	return target == errRequiredInputMissing
}

func newMissingInputError(reasons ...string) error {
	filtered := make([]string, 0, len(reasons))
	for _, reason := range reasons {
		reason = strings.TrimSpace(reason)
		if reason == "" {
			continue
		}
		filtered = append(filtered, reason)
	}
	if len(filtered) == 0 {
		return errRequiredInputMissing
	}
	return &missingInputError{Reasons: filtered}
}

func missingInputMessage(err error) string {
	var missingErr *missingInputError
	if errors.As(err, &missingErr) && len(missingErr.Reasons) > 0 {
		return strings.Join(missingErr.Reasons, "; ")
	}
	return "请检查资源字段和关联资源"
}

func joinSpecPath(parentPath string, name string) string {
	if parentPath == "" {
		return name
	}
	return parentPath + "." + name
}

func missingFieldReasons(path string, fieldNames []string) []string {
	reasons := make([]string, 0, len(fieldNames))
	for _, fieldName := range fieldNames {
		reasons = append(reasons, fmt.Sprintf("%s.%s 不能为空", path, fieldName))
	}
	return reasons
}

func missingSpecReason(path string, spec pluginx.ResourceSpec, base domain.Resource) string {
	if spec.RelationType != "" {
		return fmt.Sprintf("%s 缺少必需关联资源", path)
	}
	if spec.ModelUID != "" && base.ModelUID != spec.ModelUID {
		return fmt.Sprintf("%s 需要模型 %s", path, spec.ModelUID)
	}
	if len(spec.Filters) > 0 {
		return fmt.Sprintf("%s 不满足过滤条件", path)
	}
	return fmt.Sprintf("%s 缺少必需输入", path)
}
