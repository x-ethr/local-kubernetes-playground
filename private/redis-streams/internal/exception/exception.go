package exception

import (
	"errors"
)

var (
	Count = errors.New("unexpected runtime message count")
)

type Message interface {
	Count() error
}

type message struct{}

func (m *message) Count() error {
	return Count
}

type Exceptions interface {
	Message() Message
}

type exceptions struct{}

func (e *exceptions) Message() Message {
	return &message{}
}

func New() Exceptions {
	return &exceptions{}
}
