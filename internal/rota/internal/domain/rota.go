package domain

type TimeUnit uint8

func (s TimeUnit) ToUint8() uint8 {
	return uint8(s)
}

func (s TimeUnit) ToInt() int {
	return int(s)
}

const (
	YEARLY   TimeUnit = 1
	MONTHLY  TimeUnit = 2
	WEEKLY   TimeUnit = 3
	DAILY    TimeUnit = 4
	HOURLY   TimeUnit = 5
	MINUTELY TimeUnit = 6
	SECONDLY TimeUnit = 7
)

// Rota 值班信息
type Rota struct {
	Id        int64
	Name      string     // 名称
	Desc      string     // 描述
	Enabled   bool       // 是否启用
	Owner     int64      // 管理员
	Rules     []RotaRule // 值班规则
	TempRules []RotaRule // 临时调班
}

// RotaRule 值班规则
type RotaRule struct {
	StartTime  int64       // 开始时间
	EndTime    int64       // 结束时间
	RotaGroups []RotaGroup // 值班人员组
	Rotate     Rotate      // 轮换相关参数
}

// Rotate 轮换相关参数
type Rotate struct {
	TimeUnit     TimeUnit // 轮换单位
	TimeDuration uint8    // 轮换时间
}

// RotaGroup 值班组
type RotaGroup struct {
	Id      int64
	Name    string  // 组名称
	Members []int64 // 值班人员
}

// ShiftRostered 排班表
type ShiftRostered struct {
	FinalSchedule   []Schedule // 总排班
	CurrentSchedule Schedule   // 当前排班
	NextSchedule    Schedule   // 下期排班
}

type Schedule struct {
	Title     string    // 标题
	StartTime int64     // 开始时间
	EndTime   int64     // 结束时间
	RotaGroup RotaGroup // 分组
}
