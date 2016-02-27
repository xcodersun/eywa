package utils

import "time"
import "strconv"

func MilliSecToSec(milli int64) int64 {
	return milli * 1000000 / int64(time.Second)
}

func MilliSecToNano(milli int64) int64 {
	return milli * 1000000 % int64(time.Second)
}

func NanoToMilli(nano int64) int64 {
	return nano / int64(time.Millisecond)
}

type JSONDuration struct {
	time.Duration
}

func (d *JSONDuration) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(d.String())), nil
}

func (d *JSONDuration) UnmarshalJSON(data []byte) error {
	str, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}
	_d, err := time.ParseDuration(str)
	if err != nil {
		return err
	}

	d.Duration = _d
	return nil
}
