package utils

import (
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/vivowares/waterwheel"
	"github.com/vivowares/octopus/Godeps/_workspace/src/gopkg.in/natefinch/lumberjack.v2"
	. "github.com/vivowares/octopus/configs"
)

var Logger *waterwheel.AsyncLogger

func InitialLogger() {
	rl := &lumberjack.Logger{
		Filename:   Config().Logging.Filename,
		MaxSize:    Config().Logging.MaxSize,
		MaxBackups: Config().Logging.MaxBackups,
		MaxAge:     Config().Logging.MaxAge,
	}

	Logger = waterwheel.NewAsyncLogger(
		waterwheel.NewBufferedWriteCloser(Config().Logging.BufferSize, rl),
		waterwheel.SimpleFormatter,
		Config().Logging.BufferSize,
		Config().Logging.Level,
	)
}

func CloseLogger() {
	Logger.Close()
}
