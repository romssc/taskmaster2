package standarttime

import "time"

type Time struct{}

func New() *Time {
	return &Time{}
}

func (t *Time) TimeNow() int64 {
	return time.Now().Unix()
}
