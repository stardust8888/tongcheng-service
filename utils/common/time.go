package common

import (
	"fmt"

	"tongcheng/utils/x/timex"
)

type TimeRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

func (obj *TimeRange) BindDefault() error {
	if obj.Start == "" {
		obj.Start = timex.GetBeijingTime().Add(-24 * timex.Hour).Format()
	}
	if obj.End == "" {
		obj.End = timex.GetBeijingTime().Format()
	}
	if _, err := timex.ParseBeijingTime(obj.Start, timex.TimeFormat); err != nil {
		return fmt.Errorf("StartParse err:%v", err)
	}
	if _, err := timex.ParseBeijingTime(obj.End, timex.TimeFormat); err != nil {
		return fmt.Errorf("EndParse err:%v", err)
	}
	return nil
}
