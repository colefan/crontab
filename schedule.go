package crontab

import "time"
import "fmt"

// Schedule 定时任务的调度器接口
type Schedule interface {
	// 返回下次激活时间
	// Next 初始化时被调用一次，之后每次Run时别别用以获取下一次执行时间
	Next(time.Time) time.Time
	ShowFormat()
}

// SpecSchedule 调度器
type SpecSchedule struct {
	Second uint64 //0 0-59
	Minute uint64 //0 0-59
	Hour   uint64 //0 0-23
	Dom    uint64 //*? 0-31
	Month  uint64 //* 1-12
	Dow    uint64 //?* 1-6
}

// Next next run time
func (s *SpecSchedule) Next(t time.Time) time.Time {

	//获得当前时间的下一整秒
	t = t.Add(1*time.Second - time.Duration(t.Nanosecond())*time.Nanosecond)
	timeAdded := false
	yearLimit := t.Year() + 2

FindNextRunTime:
	if t.Year() > yearLimit {
		//2年之内找不到合适的之行时间，则放弃
		return time.Time{}
	}
	//先检查month
	for ((1 << uint64(t.Month())) & s.Month) == 0 {
		//这个月不需要执行 ，则需要跳过这个月,时间设置为下个月的开始时间
		if !timeAdded {
			timeAdded = true
			//
			t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
		}
		t = t.AddDate(0, 1, 0)

		if t.Month() == time.January {
			goto FindNextRunTime
		}
	}

	//已经找到执行的月份了，要找dom，dow
	for !dayMatches(s, t) {
		if !timeAdded {
			timeAdded = true
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		}
		t = t.AddDate(0, 0, 1)

		if t.Day() == 1 {
			goto FindNextRunTime
		}

	}

	//找小时
	for 1<<uint(t.Hour())&s.Hour == 0 {
		if !timeAdded {
			timeAdded = true
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
		}
		t = t.Add(1 * time.Hour)

		if t.Hour() == 0 {
			goto FindNextRunTime
		}
	}

	for 1<<uint(t.Minute())&s.Minute == 0 {
		if !timeAdded {
			timeAdded = true
			t = t.Truncate(time.Minute)
		}
		t = t.Add(1 * time.Minute)

		if t.Minute() == 0 {
			goto FindNextRunTime
		}
	}

	for 1<<uint(t.Second())&s.Second == 0 {
		if !timeAdded {
			timeAdded = true
			t = t.Truncate(time.Second)
		}
		t = t.Add(1 * time.Second)

		if t.Second() == 0 {
			goto FindNextRunTime
		}
	}

	return t
}

func (s *SpecSchedule) ShowFormat() {
	fmt.Printf("second :%b\n", s.Second)
	fmt.Printf("Minute :%b\n", s.Minute)
	fmt.Printf("Hour   :%b\n", s.Hour)
	fmt.Printf("Dom    :%b\n", s.Dom)
	fmt.Printf("Month  :%b\n", s.Month)
	fmt.Printf("Mow    :%b\n", s.Dow)

}

func dayMatches(s *SpecSchedule, t time.Time) bool {
	var domMatch bool
	var dowMatch bool

	if s.Dom != 0 {
		domMatch = ((1 << uint(t.Day())) & s.Dom) > 0
	}

	if s.Dow != 0 {
		dowMatch = ((1 << uint(t.Weekday())) & s.Dow) > 0
	}

	if s.Dom == 0 && s.Dow == 0 {
		return true
	}

	return domMatch || dowMatch
}
