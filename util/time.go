package util

import (
	"time"
)

//CurMilis returns the time in miliseconds.
func GetMilis(t time.Time) int64 {
	return t.UnixNano() / 1000000
}

//CurMilis returns the current time in miliseconds.
func CurMilis() int64 {
	return GetMilis(time.Now())
}

//GetTime returns an instance of time.Time for the miliseconds provided.
func GetTime(miliseconds int64) time.Time {
	return time.Unix(0, miliseconds*1000000)
}

const layout = "2006-01-02 15:04:05"

//Date returns a string representation of the date
//represented by the miliseconds provided.
func Date(miliseconds int64) string {
	return GetTime(miliseconds).Format(layout)
}
