package waterwheel

import (
	"errors"
	"io"
	"sync"
	"time"
)

var CloseErr = errors.New("logger is closed")

type AsyncLogger struct {
	sync.WaitGroup
	wc    io.WriteCloser
	buf   []byte
	level Level
	ch    chan *Record
	f     Formatter
}

func (l *AsyncLogger) Level() Level {
	return l.level
}

func NewAsyncLogger(w io.WriteCloser, f Formatter, bufSize int, lvl string) *AsyncLogger {
	ch := make(chan *Record, bufSize)

	l := &AsyncLogger{
		ch:    ch,
		f:     f,
		wc:    w,
		level: MapLevel(lvl),
		buf:   []byte{},
	}

	l.Add(1)

	go func() {
		defer l.Done()
		defer l.wc.Close()
		for {
			r, more := <-ch
			if more {
				l.buf = l.buf[:0]
				l.f(r, &l.buf)
				l.wc.Write(l.buf)
			} else {
				break
			}
		}
	}()

	return l
}

func (l *AsyncLogger) Close() {
	close(l.ch)
	l.Wait()
}

func (l *AsyncLogger) Log(lvl Level, msg string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = CloseErr
		}
	}()

	if lvl <= l.level {
		l.ch <- &Record{
			Level:   lvl,
			Time:    time.Now(),
			Message: msg,
		}
	}
	return
}

func (l *AsyncLogger) Critical(msg string) error {
	return l.Log(Critical, msg)
}

func (l *AsyncLogger) Error(msg string) error {
	return l.Log(Error, msg)
}

func (l *AsyncLogger) Warn(msg string) error {
	return l.Log(Warn, msg)
}

func (l *AsyncLogger) Info(msg string) error {
	return l.Log(Info, msg)
}

func (l *AsyncLogger) Debug(msg string) error {
	return l.Log(Debug, msg)
}
