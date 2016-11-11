package waterwheel

import "time"

type Formatter func(r *Record, buf *[]byte)

// inspired by standard lib "log"
var SimpleFormatter = func(r *Record, buf *[]byte) {
	// lvl := r.Level.String()
	// *buf = append(*buf, lvl...)
	formatLevel(r.Level, buf)
	*buf = append(*buf, ' ')

	formatTime(r.Time.UTC(), buf)

	*buf = append(*buf, r.Message...)
	*buf = append(*buf, '\n')
}

func formatLevel(lvl Level, buf *[]byte) {
	var b [8]byte
	asBytes := []byte(lvl.String())
	for i := 0; i < len(b); i++ {
		if i < len(asBytes) {
			b[i] = asBytes[i]
		} else {
			b[i] = ' '
		}
	}

	*buf = append(*buf, b[:]...)
}

// copied from standard lib.
// Cheap integer to fixed-width decimal ASCII.  Give a negative width to avoid zero-padding.
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
