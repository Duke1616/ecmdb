package plugin

import (
	"fmt"

	pluginx "github.com/Duke1616/ecmdb/pkg/plugin"
)

func buildRelationName(baseModelUID string, spec pluginx.ResourceSpec) (string, error) {
	if spec.RelationType == "" {
		return "", fmt.Errorf("插件关联类型不能为空: %s.%s", baseModelUID, spec.Name)
	}
	if !pluginx.ValidRelationType(spec.RelationType) {
		return "", fmt.Errorf("插件关联类型不支持: %s", spec.RelationType)
	}
	if spec.ModelUID == "" {
		return "", fmt.Errorf("插件关联模型不能为空: %s.%s", baseModelUID, spec.Name)
	}

	switch spec.Direction {
	case pluginx.DirectionToTarget:
		return fmt.Sprintf("%s_%s_%s", baseModelUID, spec.RelationType, spec.ModelUID), nil
	case pluginx.DirectionToSource:
		return fmt.Sprintf("%s_%s_%s", spec.ModelUID, spec.RelationType, baseModelUID), nil
	default:
		return "", fmt.Errorf("插件关联方向不支持: %s", spec.Direction)
	}
}
