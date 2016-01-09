package utils

import (
	. "github.com/vivowares/octopus/configs"
	// "gopkg.in/inconshreveable/log15.v2"
	"github.com/vivowares/octopus/Godeps/_workspace/src/gopkg.in/natefinch/lumberjack.v2"
)

var AccessLog *lumberjack.Logger
var ConnectionLog *lumberjack.Logger

func InitialLoggers() {
	AccessLog = &lumberjack.Logger{
		Filename:   Config.Logs.AccessLog.Filename,
		MaxSize:    Config.Logs.AccessLog.MaxSize,
		MaxBackups: Config.Logs.AccessLog.MaxBackups,
		MaxAge:     Config.Logs.AccessLog.MaxAge,
	}

	ConnectionLog = &lumberjack.Logger{
		Filename:   Config.Logs.ConnectionLog.Filename,
		MaxSize:    Config.Logs.ConnectionLog.MaxSize,
		MaxBackups: Config.Logs.ConnectionLog.MaxBackups,
		MaxAge:     Config.Logs.ConnectionLog.MaxAge,
	}
}
