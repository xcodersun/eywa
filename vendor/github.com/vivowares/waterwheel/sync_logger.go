package waterwheel

import (
	"io"
	"sync"
	"time"
)

type SyncLogger struct {
	wc io.WriteCloser
	sync.Mutex
	level  Level
	f      Formatter
	buf    []byte
	closed bool
}

func (l *SyncLogger) Level() Level {
	return l.level
}

func NewSyncLogger(w io.WriteCloser, f Formatter, lvl string) *SyncLogger {
	return &SyncLogger{
		wc:    w,
		level: MapLevel(lvl),
		f:     f,
		buf:   []byte{},
	}
}

func (l *SyncLogger) Close() error {
	l.Lock()
	defer l.Unlock()

	if l.closed {
		return nil
	}

	l.closed = true
	return l.wc.Close()
}

func (l *SyncLogger) Log(lvl Level, msg string) (err error) {
	if lvl <= l.level {
		l.Lock()
		defer l.Unlock()

		if l.closed {
			return CloseErr
		}

		l.buf = l.buf[:0]
		l.f(&Record{
			Level:   lvl,
			Time:    time.Now(),
			Message: msg,
		}, &l.buf)

		_, err := l.wc.Write(l.buf)
		return err
	}

	return nil
}

func (l *SyncLogger) Critical(msg string) error {
	return l.Log(Critical, msg)
}

func (l *SyncLogger) Error(msg string) error {
	return l.Log(Error, msg)
}

func (l *SyncLogger) Warn(msg string) error {
	return l.Log(Warn, msg)
}

func (l *SyncLogger) Info(msg string) error {
	return l.Log(Info, msg)
}

func (l *SyncLogger) Debug(msg string) error {
	return l.Log(Debug, msg)
}
