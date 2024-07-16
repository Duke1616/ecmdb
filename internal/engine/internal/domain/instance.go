package domain

import (
	"github.com/Bunny3th/easy-workflow/workflow/database"
)

type Instance struct {
	TaskID          int                 //任务ID
	ProcInstID      int                 //流程实例ID
	ProcID          int                 //流程ID
	ProcName        string              //流程名称
	ProcVersion     int                 //流程版本号
	BusinessID      string              //业务ID
	Starter         string              //流程发起人用户ID
	CurrentNodeID   string              //当前进行节点ID
	CurrentNodeName string              //当前进行节点名称
	CreateTime      *database.LocalTime //创建时间
	ApprovedBy      []string            //当前处理人
	Status          int                 //0:未完成(审批中) 1:已完成(通过) 2:撤销
}
