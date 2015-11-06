package connections

import "fmt"

type MessageSendingError struct {
	message string
}

func (e *MessageSendingError) Error() string {
	return fmt.Sprintf("[MessageSendingError] %s", e.message)
}

type MessageReadingError struct {
	message string
}

func (e *MessageReadingError) Error() string {
	return fmt.Sprintf("[MessageReadingError] %s", e.message)
}

type MessageTypeError struct {
	message string
}

func (e *MessageTypeError) Error() string {
	return fmt.Sprintf("[MessageTypeError] %s", e.message)
}

type ConnectionCloseError struct {
	message string
}

func (e *ConnectionCloseError) Error() string {
	return fmt.Sprintf("[ConnectionCloseError] %s", e.message)
}

type MessageParsingError struct {
	message string
}

func (e *MessageParsingError) Error() string {
	return fmt.Sprintf("[MessageParsingError] %s", e.message)
}

type ResponseTimeoutError struct {
	message string
}

func (e *ResponseTimeoutError) Error() string {
	return fmt.Sprintf("[ResponseTimeoutError] %s", e.message)
}

type MessageResponseError struct {
	message string
}

func (e *MessageResponseError) Error() string {
	return fmt.Sprintf("[MessageResponseError] %s", e.message)
}

type ConnectionRegisterError struct {
	message string
}

func (e *ConnectionRegisterError) Error() string {
	return fmt.Sprintf("[ConnectionRegisterError] %s", e.message)
}

type ConnectionUnregisterError struct {
	message string
}

func (e *ConnectionUnregisterError) Error() string {
	return fmt.Sprintf("[ConnectionUnregisterError] %s", e.message)
}

type ConnectionManagerStartingError struct {
	message string
}

func (e *ConnectionManagerStartingError) Error() string {
	return fmt.Sprintf("[ConnectionManagerStartingError] %s", e.message)
}

type UpdateRegisteredConnectionError struct {
	message string
}

func (e *UpdateRegisteredConnectionError) Error() string {
	return fmt.Sprintf("[UpdateRegisteredConnectionError] %s", e.message)
}

type WebsocketError struct {
	message string
}

func (e *WebsocketError) Error() string {
	return fmt.Sprintf("[WebsocketError] %s", e.message)
}
