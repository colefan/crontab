package crontab

import "fmt"
import "strings"
import "strconv"

//ParseOption 解析参数类型
type ParseOption int

const (
	Second     ParseOption = 1 << iota // Seconds field, default 0
	Minute                             // Minutes field, default 0
	Hour                               // Hours field, default 0
	Dom                                // Day of month field, default *
	Month                              // Month field, default *
	Dow                                // Day of week field, default *
	Descriptor                         // Allow descriptors such as @monthly, @weekly, etc.
)

var places = []ParseOption{
	Second,
	Minute,
	Hour,
	Dom,
	Month,
	Dow,
}

var defaults = []string{
	"0", //秒 default 0
	"0", //分 default 0
	"0", //小时 default 0
	"*", //Day of month default ＊
	"*", //Day of Week default *
	"*",
}

// Parser 解析器
type Parser struct {
	options   ParseOption
	optionals int
}

// NewParser  Creates a custom Parser with custom options.
func NewParser(options ParseOption) Parser {
	optionals := 0
	return Parser{options, optionals}
}

// Parse 解析
// "秒 分 时 DOM天(月） 月 DOW天（星期) "
// "0 0 0 ？ ＊ ＊"
// ? 代表这位没有意义
// 1/5 :从1开始每隔5
// 14,18 :表示14和18
// dayOfMonth 和 dayOfWeek 时冲突的，必须有一位设置为？
func Parse(spec string) (Schedule, error) {
	if len(spec) == 0 {
		return nil, fmt.Errorf("Empty spec string")
	}

	splitSpec := strings.Split(spec, " ")
	if len(splitSpec) < 6 {
		return nil, fmt.Errorf("spec string invalid should by [0 0 0 ? * *],now is [%s]", spec)
	}

	s := &SpecSchedule{}
	//0 seconds  0-12  1,2,3  0/5 9-17
	if err := parseHMS(splitSpec[0], &s.Second, 0, 59); err != nil {
		// fmt.Printf("parse Second error %v\n", err)
		return nil, err
	}
	//1 minutes 0-12  1,2,3  0/5 9-17
	// fmt.Println("Minutes = " + splitSpec[1])
	if err := parseHMS(splitSpec[1], &s.Minute, 0, 59); err != nil {
		// fmt.Printf("parse Minute error %v\n", err)

		return nil, err
	}
	//2 Hours 0-24  1,2,3 0/4 9-17
	if err := parseHMS(splitSpec[2], &s.Hour, 0, 23); err != nil {
		// fmt.Printf("parse Hour error %v\n", err)

		return nil, err
	}

	if splitSpec[3] != "?" && splitSpec[5] != "?" {
		return nil, fmt.Errorf("spec string invalid Dom & Dow should be ? at least one options")
	}
	//3 Dom *  ? 1-31 1,2,3  1-5
	if err := parseDays(splitSpec[3], &s.Dom, 1, 31); err != nil {
		// fmt.Printf("parse Dom error %v\n", err)

		return nil, err
	}
	//4 Month * 1-12 1,2,3 1-2

	if err := parseDays(splitSpec[4], &s.Month, 1, 12); err != nil {
		// fmt.Printf("parse Day error %v\n", err)
		return nil, err
	}
	//5 Dow * ? 0-6 1,2,3,5,6,0 1-2

	if err := parseDays(splitSpec[5], &s.Dow, 0, 6); err != nil {
		// fmt.Printf("parse Dow error %v\n", err)
		return nil, err
	}

	return s, nil
}

func parseDays(format string, options *uint64, min uint64, max uint64) error {
	if len(format) == 0 {
		return fmt.Errorf("parse Dom,Month,Dow format error ,is empty")
	}
	if format == "?" {
		*options = 0
		return nil
	}
	if format == "*" {
		for i := min; i <= max; i++ {
			*options |= 1 << i
		}
		return nil
	}

	if strings.Index(format, ",") >= 0 {
		splits := strings.Split(format, ",")
		for _, tmp := range splits {
			if len(tmp) > 0 {
				if val, err := strconv.ParseUint(tmp, 10, 64); err != nil {
					return err
				} else {
					*options |= (1 << val)
				}
			}
		}

	} else if strings.Index(format, "/") >= 0 {
		splits := strings.Split(format, "/")
		if len(splits) != 2 {
			return fmt.Errorf("too many / for format :" + format)
		}
		var start uint64
		var step uint64
		var err error
		if start, err = strconv.ParseUint(splits[0], 10, 64); err != nil {
			return err
		}

		if step, err = strconv.ParseUint(splits[1], 10, 64); err != nil {
			return err
		}

		for ; start <= max; start += step {
			*options |= 1 << start
		}

	} else if strings.Index(format, "-") >= 0 {
		splits := strings.Split(format, "-")
		if len(splits) != 2 {
			return fmt.Errorf("too many - for format :" + format)
		}

		var start uint64
		var stop uint64
		var err error
		if start, err = strconv.ParseUint(splits[0], 10, 64); err != nil {
			return err
		}

		if stop, err = strconv.ParseUint(splits[1], 10, 64); err != nil {
			return err
		}

		for ; start <= max && start <= stop; start++ {
			*options |= 1 << start
		}

	} else {
		return fmt.Errorf("unknown dom,month,dow format : " + format)
	}

	return nil
}

func parseHMS(format string, options *uint64, min uint64, max uint64) error {
	if len(format) == 0 {
		return fmt.Errorf("parse second,minutes,hour format error ,is empty")
	}
	// fmt.Println("format = " + format)
	if format == "*" {
		for i := min; i <= max; i++ {
			*options |= (1 << i)
		}
		return nil
	}

	if len(format) <= 2 {
		if val, err := strconv.ParseUint(format, 10, 64); err != nil {
			return err
		} else {
			*options |= (1 << val)
			return nil
		}
	}

	if strings.Index(format, ",") >= 0 {
		splits := strings.Split(format, ",")
		for _, tmp := range splits {
			if len(tmp) > 0 {
				if val, err := strconv.ParseUint(tmp, 10, 64); err != nil {
					return err
				} else {
					*options |= (1 << val)
				}
			}
		}

	} else if strings.Index(format, "/") >= 0 {
		splits := strings.Split(format, "/")
		if len(splits) != 2 {
			return fmt.Errorf("too many / for format :" + format)
		}
		var start uint64
		var step uint64
		var err error
		// fmt.Println("splits = " + splits[0] + "/" + splits[1])
		if start, err = strconv.ParseUint(splits[0], 10, 64); err != nil {
			return err
		}

		if step, err = strconv.ParseUint(splits[1], 10, 64); err != nil {
			return err
		}

		for ; start <= max; start += step {
			*options |= 1 << start
		}

	} else if strings.Index(format, "-") >= 0 {
		splits := strings.Split(format, "-")
		if len(splits) != 2 {
			return fmt.Errorf("too many - for format :" + format)
		}

		var start uint64
		var stop uint64
		var err error
		if start, err = strconv.ParseUint(splits[0], 10, 64); err != nil {
			return err
		}

		if stop, err = strconv.ParseUint(splits[1], 10, 64); err != nil {
			return err
		}

		for ; start <= max && start <= stop; start++ {
			*options |= 1 << start
		}

	} else {
		return fmt.Errorf("unknown second,minutes,hours format : " + format)
	}

	return nil
}
