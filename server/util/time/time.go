package time

import (
	"time"
)

func Now() time.Time {
	return time.Now().UTC()
}

func NowUnixMillis() int64 {
	return Now().UnixNano() / 1000000
}
