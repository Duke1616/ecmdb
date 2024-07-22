package event

import (
	"github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/task/internal/service"
	"log"
)

type TaskEvent struct {
	svc service.Service
}

func (t *TaskEvent) EventAutomation(ProcessInstanceID int, CurrentNode *model.Node, PrevNode model.Node) error {
	//可以做一些处理，比如通知流程开始人，节点到了哪个步骤
	processName, err := engine.GetProcessNameByInstanceID(ProcessInstanceID)
	if err != nil {
		return err
	}
	log.Printf("--------流程[%s]节点[%s]结束-------", processName, CurrentNode.NodeName)
	return nil
}
