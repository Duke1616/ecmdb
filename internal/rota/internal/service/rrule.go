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

	set := rrule.Set{}
	option := rrule.ROption{
		Freq:     rrule.Frequency(rule.Rotate.TimeUnit.ToInt() - 1),
		Interval: int(rule.Rotate.TimeDuration),
		//Bymonth:   []int{10, 9},
		//Byweekday: []rrule.Weekday{rrule.SA, rrule.SU},
		Dtstart: time.UnixMilli(rule.StartTime).In(location),
		Until:   time.UnixMilli(etime).In(location),
	}
	r, err := rrule.NewRRule(option)
	if err != nil {
		return domain.ShiftRostered{}, err
	}
	set.RRule(r)
	// 计算当前排班

	// 计算下期排班

	// 计算范围内排班
	data := set.Between(time.UnixMilli(rule.StartTime).In(location), time.UnixMilli(etime).In(location), true)

	var finalSchedule []domain.Schedule
	rotaGroups := rule.RotaGroups
	groupCount := len(rotaGroups)
	if groupCount == 0 {
		return domain.ShiftRostered{}, fmt.Errorf("RotaGroups is empty")
	}
	groupIndex := 0
	for _, t := range data {
		var durationMilli int64
		switch option.Freq {
		case rrule.DAILY:
			durationMilli = int64(option.Interval) * 24 * 60 * 60 * 1000
		case rrule.HOURLY:
			durationMilli = int64(option.Interval) * 60 * 60 * 1000
		default:
			return domain.ShiftRostered{}, fmt.Errorf("unsupported frequency: %d", option.Freq)
		}

		schedule := domain.Schedule{
			Title:     "空空如也",
			StartTime: t.UnixMilli(),
			EndTime:   t.UnixMilli() + durationMilli,
			RotaGroup: rotaGroups[groupIndex],
		}
		finalSchedule = append(finalSchedule, schedule)
		groupIndex = (groupIndex + 1) % groupCount
	}

	return domain.ShiftRostered{
		FinalSchedule: finalSchedule,
	}, nil
}
