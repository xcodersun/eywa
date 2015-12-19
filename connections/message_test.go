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
			Payload:     "test",
		}

		raw := Marshal(m)
		n, _ := Unmarshal(raw)
		So(reflect.DeepEqual(m, n), ShouldBeTrue)
	})

	Convey("message unmarshalling returns proper errors", t, func() {
		raw := "1|test"
		_, err := Unmarshal(raw)

		So(err.Error(), ShouldContainSubstring, "fields")

		raw = "2||"
		_, err = Unmarshal(raw)
		So(err.Error(), ShouldContainSubstring, "empty MessageId")

		raw = "0|test|this is a test message"
		_, err = Unmarshal(raw)
		So(err.Error(), ShouldContainSubstring, "invalid MessageType")
	})
}
