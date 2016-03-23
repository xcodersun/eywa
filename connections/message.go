package connections

type MessageType uint8

const (
	TypeUploadMessage   MessageType = 1 // upstream
	TypeRequestMessage  MessageType = 2 // downstream
	TypeSendMessage     MessageType = 3 // downstream
	TypeResponseMessage MessageType = 4 // upstream

	// these two messages are only used for connection states internally
	TypeConnectMessage    MessageType = 8
	TypeDisconnectMessage MessageType = 9
)

var SupportedMessageTypes = map[MessageType]string{
	TypeUploadMessage:     "upload",
	TypeResponseMessage:   "response",
	TypeSendMessage:       "send",
	TypeRequestMessage:    "request",
	TypeConnectMessage:    "connect",
	TypeDisconnectMessage: "disconnect",
}

type Message interface {
	Type() MessageType
	TypeString() string
	Id() string
	Payload() []byte
	Raw() []byte
	Marshal() ([]byte, error)
	Unmarshal() error
}
