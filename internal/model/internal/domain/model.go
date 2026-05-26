package domain

import (
	"fmt"
	"time"
)

type Model struct {
	ID      int64
	GroupId int64
	Name    string
	UID     string
	Icon    string
	Builtin bool
	Ctime   time.Time
	Utime   time.Time
}

type ModelGroup struct {
	ID   int64
	Name string
}

// EnsureDeletable 校验模型是否允许删除
// NOTE: 内置模型禁止删除，将此业务规则内聚到领域对象
func (m *Model) EnsureDeletable() error {
	if m.Builtin {
		return fmt.Errorf("内置模型不允许删除")
	}
	return nil
}

func (m *Model) SheetName() string {
	name := fmt.Sprintf("%s(%s)", m.Name, m.UID)
	if len(name) > 31 {
		// 如果超长,优先保留 model_name
		maxNameLen := 31 - len(m.UID) - 2 // 减去括号和 UID 的长度
		if maxNameLen > 0 {
			name = fmt.Sprintf("%s(%s)", m.Name[:maxNameLen], m.UID)
		} else {
			name = name[:31]
		}
	}
	return name
}
