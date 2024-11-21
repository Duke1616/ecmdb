package web

// AddRoleReq 新增值班规则
type AddRoleReq struct {
	Id       int64    `json:"id"`
	RotaRule RotaRule `json:"rota_rule"`
}

// AddOrUpdateAdjustmentRoleReq 新增修改值班规则
type AddOrUpdateAdjustmentRoleReq struct {
	Id       int64              `json:"id"`
	GroupId  int64              `json:"group_id"`
	RotaRule RotaAdjustmentRule `json:"rota_rule"`
}

type DetailById struct {
	Id int64 `json:"id"`
}

type Rota struct {
	Id              int64                `json:"id"`
	Name            string               `json:"name"`
	Desc            string               `json:"desc"`
	Enabled         bool                 `json:"enabled"`
	Owner           int64                `json:"owner"`
	Rules           []RotaRule           `json:"rules"`
	AdjustmentRules []RotaAdjustmentRule `json:"adjustment_rules"`
}

// RotaRule 值班规则
type RotaRule struct {
	StartTime  int64       `json:"start_time"`
	EndTime    int64       `json:"end_time"`
	RotaGroups []RotaGroup `json:"rota_groups"`
	Rotate     Rotate      `json:"rotate"`
}

// RotaAdjustmentRule 临时值班规则
type RotaAdjustmentRule struct {
	StartTime int64     `json:"start_time"`
	EndTime   int64     `json:"end_time"`
	RotaGroup RotaGroup `json:"rota_group"`
}

// Rotate 轮换相关参数
type Rotate struct {
	TimeUnit     uint8 `json:"time_unit"`
	TimeDuration uint8 `json:"time_duration"`
}

// RotaGroup 值班组
type RotaGroup struct {
	Id      int64   `json:"id"`
	Name    string  `json:"name"`
	Members []int64 `json:"members"`
}

// CreateRotaReq 创建值班请求
type CreateRotaReq struct {
	Name    string `json:"name"`
	Desc    string `json:"desc"`
	Enabled bool   `json:"enabled"`
	Owner   int64  `json:"owner"`
}

type ListReq struct {
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty" validate:"required"`
}

type DetailReq struct {
	Id int64 `json:"id"`
}

type DeleteReq struct {
	Id int64 `json:"id"`
}

type UpdateReq struct {
	Id      int64  `json:"id"`
	Name    string `json:"name"`
	Desc    string `json:"desc"`
	Enabled bool   `json:"enabled"`
	Owner   int64  `json:"owner"`
}

type UpdateShiftRuleReq struct {
	Id        int64      `json:"id"`
	RotaRules []RotaRule `json:"rota_rules"`
}

// GenerateShiftRosteredReq 生成排班表请求
type GenerateShiftRosteredReq struct {
	Id        int64 `json:"id"`
	StartTime int64 `json:"start_time"` // 开始时间
	EndTime   int64 `json:"end_time"`   // 结束时间
}

type RetrieveRotas struct {
	Rotas []Rota `json:"rotas"`
	Total int64  `json:"total"`
}

type RetrieveShiftRostered struct {
	FinalSchedule   []Schedule `json:"final_schedule"`
	CurrentSchedule Schedule   `json:"current_schedule"`
	NextSchedule    Schedule   `json:"next_schedule"`
	Members         []int64    `json:"members"`
}

type Schedule struct {
	Title     string    `json:"title"`
	StartTime int64     `json:"start_time"`
	EndTime   int64     `json:"end_time"`
	RotaGroup RotaGroup `json:"rota_group"`
}

type DeleteAdjustmentRoleReq struct {
	Id      int64 `json:"id"`
	GroupId int64 `json:"group_id"`
}
