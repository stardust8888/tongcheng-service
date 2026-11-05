package timex

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"tongcheng/utils/tool/jsoniter"
)

// 常用时间格式
const (
	ZeroTime          = "0001-01-01 00:00:00"
	ZeroDateTime      = "0001-01-01"
	DateMonthFormat   = "2006-01"
	DateFormat        = "2006-01-02"
	DateFormatCompact = "20060102"
	TimeFormat        = "2006-01-02 15:04:05"
	TimeMilliFormat   = "2006-01-02 15:04:05.000"
	TimeFormatCompact = "20060102150405"
	DateMonthFormatCN = "2006年01月"
	TimeFormatCNShort = "2006年01月02日"
	TimeFormatSlash   = "2006/01/02 15:04:05"
	TimeFormatZone    = "2006-01-02T15:04:05Z07:00"
	TimeLocation      = "Asia/Shanghai"
)

// 常用时间单位（以纳秒为单位）
const (
	Nanosecond  time.Duration = 1
	Microsecond               = 1000 * Nanosecond
	Millisecond               = 1000 * Microsecond
	Second                    = 1000 * Millisecond
	Minute                    = 60 * Second
	Hour                      = 60 * Minute
	Day                       = 24 * Hour
	Week                      = 7 * Day
	Month                     = 30 * Day
)

type Time time.Time

// MarshalJSON 实现 MarshalJSON 接口，用于 JSON 序列化
func (t Time) MarshalJSON() ([]byte, error) {
	if time.Time(t).IsZero() {
		return []byte(`""`), nil
	}
	beijingLoc, err := time.LoadLocation(TimeLocation)
	if err != nil {
		beijingLoc = time.FixedZone("CST", 8*3600)
	}
	formatted := time.Time(t).In(beijingLoc).Format(TimeFormat)
	return jsoniter.Marshal(formatted)
}

// UnmarshalJSON 实现 UnmarshalJSON 接口，用于 JSON 序列化
func (t *Time) UnmarshalJSON(data []byte) error {
	date := strings.Trim(string(data), `"`)
	if date == ZeroTime || date == "" {
		*t = Time{}
		return nil
	}

	beijingLoc, err := time.LoadLocation(TimeLocation)
	if err != nil {
		beijingLoc = time.FixedZone("CST", 8*3600)
	}
	tt, err := time.ParseInLocation(TimeFormat, date, beijingLoc)
	if err != nil {
		return err
	}
	*t = Time(tt)

	return nil
}

// Value 实现 Value 接口，用于数据库存储
func (t Time) Value() (driver.Value, error) {
	tt := time.Time(t)
	if tt.Format(TimeFormat) == ZeroTime {
		return nil, nil
	}
	return tt, nil
}

// Scan 实现 Scan 接口，用于从数据库读取时间
func (t *Time) Scan(value any) error {
	if value == nil {
		*t = Time{}
		return nil
	}

	if v, ok := value.(time.Time); ok {
		*t = Time(v)
		return nil
	}

	return fmt.Errorf("无法将 %v 转换为 Time", value)
}

func (t Time) Format() string {
	beijingLoc, err := time.LoadLocation(TimeLocation)
	if err != nil {
		beijingLoc = time.FixedZone("CST", 8*3600)
	}
	return time.Time(t).In(beijingLoc).Format(TimeFormat)
}

func (t Time) Unix() int64 {
	return time.Time(t).Unix()
}

func (t Time) Add(d time.Duration) Time {
	return Time(time.Time(t).Add(d))
}

func GetBeijingLocation() (*time.Location, error) {
	return time.LoadLocation(TimeLocation)
}

// GetBeijingTime 获取当前北京时间
func GetBeijingTime() Time {
	beijingLoc, err := time.LoadLocation(TimeLocation)
	if err != nil {
		// 如果加载时区失败，则使用UTC时间加8小时作为替代
		return Time(time.Now().UTC().Add(8 * time.Hour))
	}
	return Time(time.Now().In(beijingLoc))
}

func GetBjTimeFormat() string {
	return time.Time(GetBeijingTime()).Format(TimeFormat)
}

// ParseBeijingTime 将时间字符串解析为北京时间
func ParseBeijingTime(value string, format string) (Time, error) {
	var (
		t          time.Time
		beijingLoc *time.Location
		err        error
	)
	beijingLoc, err = time.LoadLocation(TimeLocation)
	if err != nil {
		// 如果加载时区失败，则使用UTC时间加8小时作为替代
		t, err = time.Parse(format, value)
		if err != nil {
			return Time{}, err
		}
		return Time(t.UTC().Add(8 * time.Hour)), nil
	}

	t, err = time.ParseInLocation(format, value, beijingLoc)
	if err != nil {
		return Time{}, err
	}

	return Time(t), nil
}

func ToTime(t time.Time) Time {
	return Time(t)
}
