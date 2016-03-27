package pubsub

import (
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/olebedev/emitter"
	"time"
)

var capacity uint = 512
var EM = emitter.New(capacity)
var EywaLogPublisher = NewBasicPublisher("log/eywa")

func Close() {
	EM.Off("*")
}

type Publisher interface {
	Topic() string
	Attached() bool
	Attach()
	Detach()
	Publish(string)
}

func FormatError(e error) string {
	buf := []byte{}
	buf = append(buf, []byte("ERROR  ")...)
	formatTime(time.Now().UTC(), &buf)
	buf = append(buf, []byte("  ")...)
	buf = append(buf, []byte(e.Error())...)
	buf = append(buf, '\n')
	return string(buf)
}

func FormatRaw(raw []byte) string {
	buf := []byte{}
	buf = append(buf, []byte("RAW    ")...)
	formatTime(time.Now().UTC(), &buf)
	buf = append(buf, []byte("  ")...)
	buf = append(buf, raw...)
	buf = append(buf, '\n')
	return string(buf)
}

func FormatIndex(index []byte) string {
	buf := []byte{}
	buf = append(buf, []byte("INDEX  ")...)
	formatTime(time.Now().UTC(), &buf)
	buf = append(buf, []byte("  ")...)
	buf = append(buf, index...)
	buf = append(buf, '\n')
	return string(buf)
}

//copied from waterwheel and stdlib
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

//copied from waterwheel and stdlib
func formatTime(t time.Time, buf *[]byte) {
	year, month, day := t.Date()
	itoa(buf, year, 4)
	*buf = append(*buf, '-')
	itoa(buf, int(month), 2)
	*buf = append(*buf, '-')
	itoa(buf, day, 2)
	*buf = append(*buf, 'T')
	hour, min, sec := t.Clock()
	itoa(buf, hour, 2)
	*buf = append(*buf, ':')
	itoa(buf, min, 2)
	*buf = append(*buf, ':')
	itoa(buf, sec, 2)
	*buf = append(*buf, '.')
	itoa(buf, t.Nanosecond()/1e3, 6)
	*buf = append(*buf, ' ')
}
