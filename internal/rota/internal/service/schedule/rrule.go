package schedule

import (
	"fmt"
	"github.com/Duke1616/ecmdb/internal/rota/internal/domain"
	"github.com/ecodeclub/ekit/slice"
	"github.com/teambition/rrule-go"
	"time"
)

const TITLE = "调班组"

type rruleSchedule struct {
	location      *time.Location
	rule          *domain.RotaRule
	durationMilli int64
	resultTime    int64
}

func NewRruleSchedule(location *time.Location) Scheduler {
	return &rruleSchedule{
		location: location,
	}
}

func (r *rruleSchedule) GetCurrentSchedule(rule domain.RotaRule,
	adjustmentRules []domain.RotaAdjustmentRule) (domain.Schedule, error) {
	r.rule = &rule

	// 解析规则
	err := r.parseRotate(time.Now().UnixMilli())
	if err != nil {
		return domain.Schedule{}, err
	}

	// 执行 rrule 获取范围值班数据
	endTime := r.computeEndTime(time.Now().UnixMilli())
	set, err := r.setRrule(endTime)
	if err != nil {
		return domain.Schedule{}, err
	}

	// 计算范围内排班
	times := set.Between(r.parserTime(rule.StartTime), r.parserTime(endTime), true)

	// 返回当前排班
	rostered, err := r.generate(times, adjustmentRules)
	return rostered.CurrentSchedule, err
}

// GenerateSchedule 根据规则生成排期表
// stime 前端选中日志范围 - 开始时间
// etime 前端选中日志范围 - 结束时间
func (r *rruleSchedule) GenerateSchedule(rule domain.RotaRule, adjustmentRules []domain.RotaAdjustmentRule,
	stime int64, etime int64) (domain.ShiftRostered, error) {
	// 结构体赋值
	r.rule = &rule

	// 解析规则
	err := r.parseRotate(stime)
	if err != nil {
		return domain.ShiftRostered{}, err
	}

	// 执行 rrule 获取范围值班数据
	endTime := r.computeEndTime(etime)
	set, err := r.setRrule(endTime)

	// 计算范围内排班
	finalTimes := set.Between(r.parserTime(rule.StartTime), r.parserTime(endTime), true)

	// 生成返回数据
	return r.generate(finalTimes, adjustmentRules)
}

// parseRotate 根据排班规则进行解析
// durationMilli: 排班间隔时间
// resultTime: 如果前端选择日期靠后，防止多余返回数据，比如当月6月，查询8月，只需要返回8月所需的数据
func (r *rruleSchedule) parseRotate(stime int64) error {
	switch r.rule.Rotate.TimeUnit {
	case domain.DAILY:
		r.durationMilli = int64(r.rule.Rotate.TimeDuration) * 24 * 60 * 60 * 1000
		r.resultTime = stime - int64(r.rule.Rotate.TimeDuration)*24*60*60*1000
	case domain.HOURLY:
		r.durationMilli = int64(r.rule.Rotate.TimeDuration) * 60 * 60 * 1000
		r.resultTime = stime - int64(r.rule.Rotate.TimeDuration)*60*60*1000
	default:
		return fmt.Errorf("unsupported frequency: %d", r.rule.Rotate.TimeUnit)
	}

	return nil
}

func (r *rruleSchedule) computeEndTime(etime int64) int64 {
	// 如果设置了截止时间，查询就以截止时间为止
	endTime := etime
	if r.rule.EndTime != 0 {
		endTime = r.rule.EndTime
	}

	// 如果前端选择结束时间，是小于计算规则的开始时间，那么默认只需要计算当前排班和下期排班
	if etime < r.rule.StartTime {
		endTime = time.Now().UnixMilli()
	}

	return endTime
}

func (r *rruleSchedule) generate(finalTimes []time.Time, adjustmentRules []domain.RotaAdjustmentRule) (domain.ShiftRostered, error) {
	var finalSchedule []domain.Schedule
	groupCount := len(r.rule.RotaGroups)
	if groupCount == 0 {
		return domain.ShiftRostered{}, fmt.Errorf("RotaGroups is empty")
	}
	groupIndex := 0

	// 获取当前时间
	currentTime := time.Now().UnixMilli()

	// 初始化当前排班和下期排班
	var currentSchedule, nextSchedule *domain.Schedule

	for _, t := range finalTimes {
		schedule := domain.Schedule{
			Title:     "值班组",
			StartTime: t.UnixMilli(),
			EndTime:   t.UnixMilli() + r.durationMilli,
			RotaGroup: r.rule.RotaGroups[groupIndex],
		}

		// 判断当前排班
		if currentSchedule == nil && currentTime >= t.UnixMilli() && currentTime < t.UnixMilli()+r.durationMilli {
			currentSchedule = &schedule

			// 获取下期排班
			nextGroupIndex := (groupIndex + 1) % groupCount
			nextScheduleStartTime := schedule.EndTime
			nextSchedule = &domain.Schedule{
				Title:     "值班组",
				StartTime: nextScheduleStartTime,
				EndTime:   nextScheduleStartTime + r.durationMilli,
				RotaGroup: r.rule.RotaGroups[nextGroupIndex],
			}
		}

		groupIndex = (groupIndex + 1) % groupCount

		// 判断是否需要进行丢弃
		if t.UnixMilli() < r.resultTime {
			continue
		}

		finalSchedule = append(finalSchedule, schedule)
	}

	// 如果没有计算
	if currentSchedule == nil && nextSchedule == nil {
		return domain.ShiftRostered{
			FinalSchedule: finalSchedule,
		}, nil
	}

	// 过滤临时调班进行返回
	return r.filterAdjustmentSchedule(domain.ShiftRostered{
		FinalSchedule:   finalSchedule,
		CurrentSchedule: *currentSchedule,
		NextSchedule:    *nextSchedule,
	}, adjustmentRules), nil
}

// setRrule 通过 rrule 自动计算出范围时间
func (r *rruleSchedule) setRrule(endTime int64) (rrule.Set, error) {
	set := rrule.Set{}
	option := rrule.ROption{
		Freq:     rrule.Frequency(r.rule.Rotate.TimeUnit.ToInt() - 1),
		Interval: int(r.rule.Rotate.TimeDuration),
		Dtstart:  r.parserTime(r.rule.StartTime),
		Until:    r.parserTime(endTime),
	}

	rr, err := rrule.NewRRule(option)
	if err != nil {
		return rrule.Set{}, err
	}

	set.RRule(rr)
	return set, nil
}

// parserTime 时间转换解析
func (r *rruleSchedule) parserTime(unixTime int64) time.Time {
	return time.UnixMilli(unixTime).In(r.location)
}

// filterAdjustmentSchedule 过滤临时调班
func (r *rruleSchedule) filterAdjustmentSchedule(rostered domain.ShiftRostered,
	adjustmentRules []domain.RotaAdjustmentRule) domain.ShiftRostered {
	rs := slice.ToMap(rostered.FinalSchedule, func(element domain.Schedule) int64 {
		return element.StartTime
	})

	for _, rule := range adjustmentRules {
		if _, ok := rs[rule.StartTime]; !ok {
			continue
		}

		d := domain.Schedule{
			Title:     TITLE,
			StartTime: rule.StartTime,
			EndTime:   rule.EndTime,
			RotaGroup: rule.RotaGroup,
		}

		// 判定当前值班是否重复
		if rostered.CurrentSchedule.StartTime == rule.StartTime &&
			rostered.CurrentSchedule.EndTime == rule.EndTime {
			rostered.CurrentSchedule = domain.Schedule{
				Title:     TITLE,
				StartTime: rule.StartTime,
				EndTime:   rule.EndTime,
				RotaGroup: rule.RotaGroup,
			}
		}

		// 判定下期值班是否重复
		if rostered.NextSchedule.StartTime == rule.StartTime &&
			rostered.NextSchedule.EndTime == rule.EndTime {
			rostered.NextSchedule = domain.Schedule{
				Title:     TITLE,
				StartTime: rule.StartTime,
				EndTime:   rule.EndTime,
				RotaGroup: rule.RotaGroup,
			}
		}

		// 使用 map 来检查元素是否已经存在
		memberMap := make(map[int64]bool)
		for _, member := range rostered.Members {
			memberMap[member] = true
		}

		// 追加新成员，避免重复
		for _, member := range rule.RotaGroup.Members {
			if !memberMap[member] {
				rostered.Members = append(rostered.Members, member)
				memberMap[member] = true
			}
		}
		rs[rule.StartTime] = d
	}

	rostered.FinalSchedule = fromMap(rs)

	return rostered
}

func fromMap[T any, K comparable](input map[K]T) []T {
	result := make([]T, 0, len(input)) // 创建切片，初始容量等于 map 的大小
	for _, value := range input {
		result = append(result, value) // 将 map 中的值添加到切片
	}
	return result
}
