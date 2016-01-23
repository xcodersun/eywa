package connections

import (
	"reflect"
	"testing"

	. "github.com/vivowares/octopus/Godeps/_workspace/src/github.com/smartystreets/goconvey/convey"
)

func TestMessageSerializations(t *testing.T) {
	Convey("message can be marshaled/unmarshaled", t, func() {
		m := &Message{
			MessageType: SyncRequestMessage,
			MessageId:   "1",
			Payload:     []byte("\u2318|\u2317"),
		}

		raw, _ := Marshal(m)
		n, _ := Unmarshal(raw)
		So(reflect.DeepEqual(m, n), ShouldBeTrue)
	})

	Convey("message unmarshalling returns proper errors", t, func() {
		raw := []byte("1|test")
		_, err := Unmarshal(raw)
		So(err.Error(), ShouldContainSubstring, "pips")

		raw = []byte("2||")
		_, err = Unmarshal(raw)
		So(err.Error(), ShouldContainSubstring, "empty MessageId")

		raw = []byte("98|test|this is a test message")
		_, err = Unmarshal(raw)
		So(err.Error(), ShouldContainSubstring, "invalid MessageType")

		raw = []byte("-|test|this is a test message")
		_, err = Unmarshal(raw)
		So(err.Error(), ShouldContainSubstring, "strconv.ParseIn")
	})
}
