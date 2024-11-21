package rota

import (
	"fmt"
	"github.com/teambition/rrule-go"
	"testing"
	"time"
)

func printTimeSlice(ts []time.Time) {
	for _, t := range ts {
		fmt.Println(t)
	}
}

func TestRRule(t *testing.T) {
	// Daily, for 10 occurrences.
	r, _ := rrule.NewRRule(rrule.ROption{
		Freq:    rrule.DAILY,
		Count:   10,
		Dtstart: time.Date(1997, 9, 2, 9, 0, 0, 0, time.UTC),
	})

	fmt.Println(r.String())
	// DTSTART:19970902T090000Z
	// RRULE:FREQ=DAILY;COUNT=10

	printTimeSlice(r.All())
	// 1997-09-02 09:00:00 +0000 UTC
	// 1997-09-03 09:00:00 +0000 UTC
	// ...
	// 1997-09-07 09:00:00 +0000 UTC

	printTimeSlice(r.Between(
		time.Date(1997, 9, 6, 0, 0, 0, 0, time.UTC),
		time.Date(1997, 9, 8, 0, 0, 0, 0, time.UTC), true))
	// [1997-09-06 09:00:00 +0000 UTC
	//  1997-09-07 09:00:00 +0000 UTC]

	// Every four years, the first Tuesday after a Monday in November, 3 occurrences (U.S. Presidential Election day).
	r, _ = rrule.NewRRule(rrule.ROption{
		Freq:     rrule.YEARLY,
		Interval: 4,
		Count:    3,
		//Bymonth:    []int{11},
		//Byweekday:  []rrule.Weekday{rrule.TU},
		//Bymonthday: []int{2, 3, 4, 5, 6, 7, 8},
		Dtstart: time.Date(1996, 11, 5, 9, 0, 0, 0, time.UTC),
	})

	fmt.Println(r.String())
	// DTSTART:19961105T090000Z
	// RRULE:FREQ=YEARLY;INTERVAL=4;COUNT=3;BYMONTH=11;BYMONTHDAY=2,3,4,5,6,7,8;BYDAY=TU

	printTimeSlice(r.All())
	// 1996-11-05 09:00:00 +0000 UTC
	// 2000-11-07 09:00:00 +0000 UTC
	// 2004-11-02 09:00:00 +0000 UTC
}

func TestSet(t *testing.T) {
	set := rrule.Set{}
	r, _ := rrule.NewRRule(rrule.ROption{
		Freq:      rrule.DAILY,
		Interval:  1,
		Bymonth:   []int{10, 9},
		Byweekday: []rrule.Weekday{rrule.SA, rrule.SU},
		Dtstart:   time.Date(2024, 9, 2, 9, 0, 0, 0, time.UTC),
		Until:     time.Date(2025, 9, 2, 9, 0, 0, 0, time.UTC),
	})

	set.RRule(r)

	printTimeSlice(set.Between(time.Date(2024, 9, 2, 9, 0, 0, 0, time.UTC),
		time.Date(2024, 12, 2, 9, 0, 0, 0, time.UTC),
		true))
}

func TestExampleSet(t *testing.T) {
	// Daily, for 7 days, jumping Saturday and Sunday occurrences.
	set := rrule.Set{}
	r, _ := rrule.NewRRule(rrule.ROption{
		Freq:    rrule.DAILY,
		Count:   7,
		Dtstart: time.Date(1997, 9, 2, 9, 0, 0, 0, time.UTC)})
	set.RRule(r)

	fmt.Println(set.String())
	// DTSTART:19970902T090000Z
	// RRULE:FREQ=DAILY;COUNT=7

	printTimeSlice(set.Between(time.Date(1997, 9, 2, 9, 0, 0, 0, time.UTC), time.Date(1997, 12, 2, 9, 0, 0, 0, time.UTC), true))
	// 1997-09-02 09:00:00 +0000 UTC
	// 1997-09-03 09:00:00 +0000 UTC
	// 1997-09-04 09:00:00 +0000 UTC
	// 1997-09-05 09:00:00 +0000 UTC
	// 1997-09-06 09:00:00 +0000 UTC
	// 1997-09-07 09:00:00 +0000 UTC
	// 1997-09-08 09:00:00 +0000 UTC

	// Weekly, for 4 weeks, plus one time on day 7, and not on day 16.
	set = rrule.Set{}
	r, _ = rrule.NewRRule(rrule.ROption{
		Freq:    rrule.WEEKLY,
		Count:   4,
		Dtstart: time.Date(1997, 9, 2, 9, 0, 0, 0, time.UTC)})
	set.RRule(r)
	set.RDate(time.Date(1997, 9, 7, 9, 0, 0, 0, time.UTC))
	set.ExDate(time.Date(1997, 9, 16, 9, 0, 0, 0, time.UTC))

	fmt.Println(set.String())
	// DTSTART:19970902T090000Z
	// RRULE:FREQ=WEEKLY;COUNT=4
	// RDATE:19970907T090000Z
	// EXDATE:19970916T090000Z

	printTimeSlice(set.All())
	// 1997-09-02 09:00:00 +0000 UTC
	// 1997-09-07 09:00:00 +0000 UTC
	// 1997-09-09 09:00:00 +0000 UTC
	// 1997-09-23 09:00:00 +0000 UTC
}

type MyInterface interface {
	SetValue(int)
	GetValue() int
}

type MyStruct struct {
	value int
}

func (m *MyStruct) SetValue(v int) {
	m.value = v
}

func (m *MyStruct) GetValue() int {
	return m.value
}

func TestEa(*testing.T) {
	// 创建一个结构体实例
	s := MyStruct{}
	t := MyStruct{}

	// 将结构体指针赋值给接口
	var iface1 MyInterface = &s
	var iface2 MyInterface = &t

	// 通过接口调用方法
	iface1.SetValue(20)
	iface2.SetValue(40)

	// 打印原始结构体的值
	fmt.Println("Original struct value:", s.value) // 输出: Original struct value: 20
	fmt.Println("Original struct value:", t.value)

	// 打印接口中的值
	fmt.Println("Interface value:", iface1.GetValue()) // 输出: Interface value: 20
	fmt.Println("Original struct value:", iface2.GetValue())

}
