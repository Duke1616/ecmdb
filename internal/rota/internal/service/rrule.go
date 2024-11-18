package service

import (
	"fmt"
	"github.com/Duke1616/ecmdb/internal/rota/internal/domain"
	"github.com/teambition/rrule-go"
	"time"
)

func RruleSchedule(rule domain.RotaRule, stime int64, etime int64) (domain.ShiftRostered, error) {
	// 创建一个位置对象，表示中国北京的位置
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return domain.ShiftRostered{}, err
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
		Until:    time.UnixMilli(etime).In(location),
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

	for i, t := range data {

		schedule := domain.Schedule{
			Title:     "空空如也",
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
			d := data[i+1]
			nextSchedule = &domain.Schedule{
				Title:     "空空如也",
				StartTime: d.UnixMilli(),
				EndTime:   d.UnixMilli() + durationMilli,
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
