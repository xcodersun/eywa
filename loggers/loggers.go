package loggers

import (
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/vivowares/waterwheel"
	"github.com/vivowares/eywa/Godeps/_workspace/src/gopkg.in/natefinch/lumberjack.v2"
	. "github.com/vivowares/eywa/configs"
	"log"
)

var Logger *waterwheel.AsyncLogger

var esWc *waterwheel.BufferedWriteCloser
var ESLogger *log.Logger

var dbWc *waterwheel.BufferedWriteCloser
var DBLogger *log.Logger

func InitialLogger() {
	rl := &lumberjack.Logger{
		Filename:   Config().Logging.Eywa.Filename,
		MaxSize:    Config().Logging.Eywa.MaxSize,
		MaxBackups: Config().Logging.Eywa.MaxBackups,
		MaxAge:     Config().Logging.Eywa.MaxAge,
	}

	Logger = waterwheel.NewAsyncLogger(
		waterwheel.NewBufferedWriteCloser(Config().Logging.Eywa.BufferSize, rl),
		waterwheel.SimpleFormatter,
		Config().Logging.Eywa.BufferSize,
		Config().Logging.Eywa.Level,
	)

	esWc = waterwheel.NewBufferedWriteCloser(Config().Logging.Indices.BufferSize,
		&lumberjack.Logger{
			Filename:   Config().Logging.Indices.Filename,
			MaxSize:    Config().Logging.Indices.MaxSize,
			MaxBackups: Config().Logging.Indices.MaxBackups,
			MaxAge:     Config().Logging.Indices.MaxAge,
		},
	)

	ESLogger = log.New(
		esWc,
		"",
		log.LUTC|log.Lmicroseconds|log.Ldate|log.Lshortfile|log.Ltime,
	)

	dbWc = waterwheel.NewBufferedWriteCloser(Config().Logging.Database.BufferSize,
		&lumberjack.Logger{
			Filename:   Config().Logging.Database.Filename,
			MaxSize:    Config().Logging.Database.MaxSize,
			MaxBackups: Config().Logging.Database.MaxBackups,
			MaxAge:     Config().Logging.Database.MaxAge,
		},
	)

	DBLogger = log.New(
		dbWc,
		"",
		log.LUTC|log.Lmicroseconds|log.Ldate|log.Lshortfile|log.Ltime,
	)
}

func CloseLogger() {
	Logger.Close()
	esWc.Close()
	dbWc.Close()
}
