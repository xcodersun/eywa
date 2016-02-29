package utils

import (
	"errors"
	"reflect"
	"strconv"
	"time"
)

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

var JSONDurationAssignReader = func(v interface{}, isPtr bool) (reflect.Value, error) {
	var _d time.Duration
	var err error
	if f, ok := v.(float64); ok {
		_d = time.Duration(int64(f))
	} else if i, ok := v.(int64); ok {
		_d = time.Duration(i)
	} else if s, ok := v.(string); ok {
		_d, err = time.ParseDuration(s)
		if err != nil {
			return reflect.Value{}, err
		}
	} else {
		return reflect.Value{}, errors.New("interface is not a duration")
	}

	d := JSONDuration{Duration: _d}
	if isPtr {
		return reflect.ValueOf(&d), nil
	} else {
		return reflect.ValueOf(d), nil
	}
}
