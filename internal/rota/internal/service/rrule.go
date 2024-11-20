package service

import (
	"fmt"
	"github.com/Duke1616/ecmdb/internal/rota/internal/domain"
	"github.com/ecodeclub/ekit/slice"
	"github.com/teambition/rrule-go"
	"time"
)

func RruleSchedule(rule domain.RotaRule, stime int64, etime int64) (domain.ShiftRostered, error) {
	// 创建一个位置对象，表示中国北京的位置
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return domain.ShiftRostered{}, err
	}

	// 如果设置了，截止时间，查询就以截止时间为止
	endTime := etime
	if rule.EndTime != 0 {
		endTime = rule.EndTime
	}

	var durationMilli int64
	resultTime := stime
	switch rule.Rotate.TimeUnit {
	case domain.DAILY:
		durationMilli = int64(rule.Rotate.TimeDuration) * 24 * 60 * 60 * 1000
		resultTime = stime - int64(rule.Rotate.TimeDuration)*24*60*60*1000
	case domain.HOURLY:
		durationMilli = int64(rule.Rotate.TimeDuration) * 60 * 60 * 1000
	default:
		return domain.ShiftRostered{}, fmt.Errorf("unsupported frequency: %d", rule.Rotate.TimeUnit)
	}

	set := rrule.Set{}
	option := rrule.ROption{
		Freq:     rrule.Frequency(rule.Rotate.TimeUnit.ToInt() - 1),
		Interval: int(rule.Rotate.TimeDuration),
		Dtstart:  time.UnixMilli(rule.StartTime).In(location),
		Until:    time.UnixMilli(endTime).In(location),
	}
	r, err := rrule.NewRRule(option)
	if err != nil {
		return domain.ShiftRostered{}, err
	}
	set.RRule(r)

	// 计算范围内排班
	data := set.Between(time.UnixMilli(rule.StartTime).In(location), time.UnixMilli(etime).In(location), true)
	var finalSchedule []domain.Schedule
	rotaGroups := rule.RotaGroups
	groupCount := len(rotaGroups)
	if groupCount == 0 {
		return domain.ShiftRostered{}, fmt.Errorf("RotaGroups is empty")
	}
	groupIndex := 0

	// 获取当前时间
	currentTime := time.Now().UnixMilli()

	// 初始化当前排班和下期排班
	var currentSchedule, nextSchedule *domain.Schedule

	for _, t := range data {
		schedule := domain.Schedule{
			Title:     "值班组",
			StartTime: t.UnixMilli(),
			EndTime:   t.UnixMilli() + durationMilli,
			RotaGroup: rotaGroups[groupIndex],
		}

		// 判断当前排班
		if currentTime >= t.UnixMilli() && currentTime < t.UnixMilli()+durationMilli && currentSchedule == nil {
			currentSchedule = &schedule
		}

		// 判断下期排班
		if currentTime >= t.UnixMilli() && currentTime < t.UnixMilli()+durationMilli && nextSchedule == nil {
			index := (groupIndex + 1) % groupCount
			nextSchedule = &domain.Schedule{
				Title:     "值班组",
				StartTime: t.UnixMilli() + durationMilli,
				EndTime:   t.UnixMilli() + durationMilli + durationMilli,
				RotaGroup: rotaGroups[index],
			}
		}

		groupIndex = (groupIndex + 1) % groupCount
		// 判断是否需要保存
		if t.UnixMilli() < resultTime {
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

	return domain.ShiftRostered{
		FinalSchedule:   finalSchedule,
		CurrentSchedule: *currentSchedule,
		NextSchedule:    *nextSchedule,
	}, nil
}

func RruleAdjustmentSchedule(rostered domain.ShiftRostered, adjustmentRules []domain.RotaAdjustmentRule) (domain.ShiftRostered, error) {
	rs := slice.ToMap(rostered.FinalSchedule, func(element domain.Schedule) int64 {
		return element.StartTime
	})

	title := "调班组"
	for _, rule := range adjustmentRules {
		_, ok := rs[rule.StartTime]
		if ok {
			d := domain.Schedule{
				Title:     title,
				StartTime: rule.StartTime,
				EndTime:   rule.EndTime,
				RotaGroup: rule.RotaGroup,
			}

			// 判定当前值班是否重复
			if rostered.CurrentSchedule.StartTime == rule.StartTime &&
				rostered.CurrentSchedule.EndTime == rule.EndTime {
				rostered.CurrentSchedule = domain.Schedule{
					Title:     title,
					StartTime: rule.StartTime,
					EndTime:   rule.EndTime,
					RotaGroup: rule.RotaGroup,
				}
			}

			// 判定下期值班是否重复
			if rostered.NextSchedule.StartTime == rule.StartTime &&
				rostered.NextSchedule.EndTime == rule.EndTime {
				rostered.NextSchedule = domain.Schedule{
					Title:     title,
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
	}

	rostered.FinalSchedule = FromMap(rs)

	return rostered, nil
}

func FromMap[T any, K comparable](input map[K]T) []T {
	result := make([]T, 0, len(input)) // 创建切片，初始容量等于 map 的大小
	for _, value := range input {
		result = append(result, value) // 将 map 中的值添加到切片
	}
	return result
}
