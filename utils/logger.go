package utils

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/vivowares/octopus/Godeps/_workspace/src/gopkg.in/inconshreveable/log15.v2"
	"github.com/vivowares/octopus/Godeps/_workspace/src/gopkg.in/natefinch/lumberjack.v2"
	. "github.com/vivowares/octopus/configs"
	"io"
	"strings"
	"sync"
)

var Logger log15.Logger
var loggerChan chan *log15.Record
var loggerWg sync.WaitGroup

const (
	TimeFormat = "2006-01-02T15:04:05"
)

func InitialLogger() error {
	lvl, err := log15.LvlFromString(Config.Logging.Level)
	if err != nil {
		return err
	}

	rotatingLogger := &lumberjack.Logger{
		Filename:   Config.Logging.Filename,
		MaxSize:    Config.Logging.MaxSize,
		MaxBackups: Config.Logging.MaxBackups,
		MaxAge:     Config.Logging.MaxAge,
	}

	Logger = log15.New()

	loggerWg.Add(1)
	Logger.SetHandler(
		log15.LvlFilterHandler(lvl,
			AsyncHandler(Config.Logging.BufferSize,
				NewBufferedHandler(Config.Logging.BufferSize, rotatingLogger, Formatter(UncoloredTerminalFormatter))),
		),
	)

	return nil
}

func CloseLogger() {
	close(loggerChan)

	loggerWg.Wait()
}

type BufferedHandler struct {
	cur  int
	size int
	w    *bufio.Writer
	f    Formatter
}

func NewBufferedHandler(s int, w io.Writer, f Formatter) *BufferedHandler {
	return &BufferedHandler{
		size: s,
		w:    bufio.NewWriter(w),
		f:    f,
	}
}

func (l *BufferedHandler) Log(r *log15.Record) error {
	p := l.f(r)

	if l.cur <= l.size {
		l.cur += 1
		_, err := l.w.Write(p)
		return err
	}

	err := l.w.Flush()
	l.w.Write(p)
	l.cur = 1
	return err
}

func AsyncHandler(bufSize int, h log15.Handler) log15.Handler {
	loggerChan = make(chan *log15.Record, bufSize)

	go func() {
		defer loggerWg.Done()
		defer h.(*BufferedHandler).w.Flush()
		for {
			r, more := <-loggerChan
			if more {
				_ = h.Log(r)
			} else {
				break
			}
		}
	}()

	return log15.ChannelHandler(loggerChan)
}

func ChannelHandler(recs chan<- *log15.Record) log15.Handler {
	return log15.FuncHandler(func(r *log15.Record) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = errors.New("logger is closed")
			}
		}()

		recs <- r
		return
	})
}

type Formatter func(r *log15.Record) []byte

var UncoloredTerminalFormatter = func(r *log15.Record) []byte {
	lvl := strings.ToUpper(r.Lvl.String())
	b := &bytes.Buffer{}
	fmt.Fprintf(b, "%s: [%s] %s \n", lvl, r.Time.UTC().Format(TimeFormat), r.Msg)
	return b.Bytes()
}
