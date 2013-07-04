package util

import(
	"time"
)

func CurMilis()int64{
	return time.Now().UnixNano()/1000000
}

func GetTime(miliseconds int64) time.Time {
	return time.Unix(0, miliseconds*1000000)
}

const layout = "2006-01-02 15:04:05"

func Date(miliseconds int64) string{
	return GetTime(miliseconds).Format(layout)
}