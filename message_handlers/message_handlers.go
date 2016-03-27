package message_handlers

import (
	"errors"
	"fmt"
	. "github.com/vivowares/eywa/connections"
	"github.com/vivowares/eywa/models"
	"strings"
	"time"
)

var SupportedMessageHandlers = map[string]*Middleware{"indexer": Indexer, "logger": Logger}

var channelNotFound = errors.New("channel not found when indexing data")

func findCachedChannel(idStr string) (*models.Channel, bool) {
	id := models.DecodeHashId(idStr)
	ch, found := models.FetchCachedChannelById(id)
	return ch, found
}

func formatError(e error) string {
	buf := []byte{}
	buf = append(buf, []byte(fmt.Sprintf("%-12s", "ERROR"))...)
	formatTime(time.Now().UTC(), &buf)
	buf = append(buf, []byte("  ")...)
	buf = append(buf, []byte(e.Error())...)
	buf = append(buf, '\n')
	return string(buf)
}

func formatMessage(m Message) string {
	buf := []byte{}
	t := m.TypeString()
	if len(t) == 0 {
		t = "WRONG TYPE"
	}
	buf = append(buf, []byte(
		fmt.Sprintf("%-12s", strings.ToUpper(t)),
	)...)
	formatTime(time.Now().UTC(), &buf)
	buf = append(buf, []byte("  ")...)
	raw := []byte{}
	if m.Raw() != nil {
		raw = m.Raw()
	}
	buf = append(buf, raw...)
	buf = append(buf, '\n')
	return string(buf)
}

func formatIndex(index []byte) string {
	buf := []byte{}
	buf = append(buf, []byte(fmt.Sprintf("%-12s", "INDEX"))...)
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
