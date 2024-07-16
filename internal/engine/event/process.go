package event

import (
	"context"
	"github.com/Bunny3th/easy-workflow/workflow/database"
	"github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/gotomicro/ego/core/elog"
	"log"
)

// RoleUser 这里创建了一个角色-用户的人员库，用来模拟数据库中存储的角色-用户对应关系
var RoleUser = make(map[string][]string)

func init() {
	//初始化人事数据
	RoleUser["主管"] = []string{"张经理"}
	RoleUser["人事经理"] = []string{"人事老刘"}
	RoleUser["老板"] = []string{"李老板", "老板娘"}
	RoleUser["副总"] = []string{"赵总", "钱总", "孙总"}
}

type ProcessEvent struct {
	producer OrderStatusModifyEventProducer
	logger   *elog.Component
}

func NewProcessEvent(producer OrderStatusModifyEventProducer) *ProcessEvent {
	return &ProcessEvent{
		producer: producer,
		logger:   elog.DefaultLogger,
	}
}

// EventStart 节点结束事件
func (e *ProcessEvent) EventStart(ProcessInstanceID int, CurrentNode *model.Node, PrevNode model.Node) error {
	//可以做一些处理，比如通知流程开始人，节点到了哪个步骤
	processName, err := engine.GetProcessNameByInstanceID(ProcessInstanceID)
	if err != nil {
		return err
	}
	log.Printf("--------流程[%s]节点[%s]结束-------", processName, CurrentNode.NodeName)
	return nil
}

// EventEnd 节点结束事件
func (e *ProcessEvent) EventEnd(ProcessInstanceID int, CurrentNode *model.Node, PrevNode model.Node) error {
	processName, err := engine.GetProcessNameByInstanceID(ProcessInstanceID)
	if err != nil {
		return err
	}

	e.logger.Info("节点结束了", elog.Any("processName", processName))
	log.Printf("--------流程[%s]节点[%s]结束-------", processName, CurrentNode.NodeName)
	return nil
}

// EventClose 流程结束，修改 Order 状态为已完成
func (e *ProcessEvent) EventClose(ProcessInstanceID int, CurrentNode *model.Node, PrevNode model.Node) error {
	evt := OrderStatusModifyEvent{
		ProcessInstanceId: ProcessInstanceID,
		Status:            END,
	}

	err := e.producer.Produce(context.Background(), evt)
	if err != nil {
		// 要做好监控和告警
		e.logger.Error("发送修改 Order 事件失败",
			elog.FieldErr(err),
			elog.Any("evt", evt))
	}
	return nil
}

// EventNotify 通知
func (e *ProcessEvent) EventNotify(ProcessInstanceID int, CurrentNode *model.Node, PrevNode model.Node) error {
	processName, err := engine.GetProcessNameByInstanceID(ProcessInstanceID)
	if err != nil {
		return err
	}
	log.Printf("--------流程[%s]节点[%s]，通知节点中对应人员--------", processName, CurrentNode.NodeName)
	if CurrentNode.NodeType == model.EndNode {
		log.Printf("============================== 流程[%s]结束 ==============================", processName)
		variables, err := engine.ResolveVariables(ProcessInstanceID, []string{"$starter"})
		if err != nil {
			return err
		}
		log.Printf("通知流程创建人%s,流程[%s]已完成", variables["$starter"], processName)

	} else {
		for _, user := range CurrentNode.UserIDs {
			log.Printf("通知用户[%s],抓紧去处理", user)
		}
	}
	return nil
}

// EventTaskForceNodePass 任务事件
// 在示例流程中，"副总审批"是一个会签节点，需要3个副总全部通过，节点才算通过
// 现在通过任务事件改变会签通过人数，设为只要2人通过，即算通过
func (e *ProcessEvent) EventTaskForceNodePass(TaskID int, CurrentNode *model.Node, PrevNode model.Node) error {
	taskInfo, err := engine.GetTaskInfo(TaskID)
	if err != nil {
		return err
	}

	processName, err := engine.GetProcessNameByInstanceID(taskInfo.ProcInstID)
	if err != nil {
		return err
	}
	log.Printf("--------流程[%s]节点[%s],开始任务ID[%d]结束事件--------", processName, CurrentNode.NodeName, taskInfo.TaskID)
	_, PassNum, _, err := engine.TaskNodeStatus(taskInfo.TaskID)
	if err != nil {
		return err
	}
	//如果通过数>=2，则:
	//1、直接把节点中所有任务都置为通过and结束，这样节点就强制被完成
	//2、自动生成comment，以免其他被代表的用户疑惑

	if PassNum >= 2 {
		tx := engine.DB.Begin()

		//找到本节点那些还没有通过的task
		var tasks []database.ProcTask
		result := tx.Where("proc_inst_id=? AND node_id=? AND batch_code=? AND is_finished=0",
			taskInfo.ProcInstID, taskInfo.NodeID, taskInfo.BatchCode).Find(&tasks)
		if result.Error != nil {
			return result.Error
		}

		//代表他们通过
		result = tx.Model(&database.ProcTask{}).
			Where("proc_inst_id=? AND node_id=? AND batch_code=? AND is_finished=0", taskInfo.ProcInstID, taskInfo.NodeID, taskInfo.BatchCode).
			Updates(database.ProcTask{Comment: "通过人数已满2人，系统自动代表你通过", IsFinished: 1, Status: 1})
		if result.Error != nil {
			return result.Error
		}
		tx.Commit()
	}

	return nil
}

// EventRevoke 流程撤销
func (e *ProcessEvent) EventRevoke(ProcessInstanceID int, RevokeUserID string) error {
	processName, err := engine.GetProcessNameByInstanceID(ProcessInstanceID)
	if err != nil {
		return err
	}
	log.Printf("--------流程[%s],由[%s]发起撤销--------", processName, RevokeUserID)

	return nil
}
