package utils

import "time"

func MilliSecToSec(milli int64) int64 {
	return milli * 1000000 / int64(time.Second)
}

func MilliSecToNano(milli int64) int64 {
	return milli * 1000000 % int64(time.Second)
}

func NanoToMilli(nano int64) int64 {
	return nano / int64(time.Millisecond)
}
