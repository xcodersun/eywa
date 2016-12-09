package waterwheel

import (
	"strings"
	"time"
)

type Level int

const (
	Critical Level = iota
	Error
	Warn
	Info
	Debug
)

func (l Level) String() string {
	switch l {
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warn:
		return "WARN"
	case Error:
		return "ERROR"
	case Critical:
		return "CRITICAL"
	default:
		panic("bad level")
	}
}

func MapLevel(lvl string) Level {
	switch strings.ToUpper(lvl) {
	case "CRITICAL":
		return Critical
	case "ERROR":
		return Error
	case "WARN":
		return Warn
	case "INFO":
		return Info
	case "DEBUG":
		return Debug
	default:
		return Debug
	}
}

type Record struct {
	Level   Level
	Time    time.Time
	Message string
}
