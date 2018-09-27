package actions

import (
	"bytes"
	"encoding/binary"
	"errors"
)

var (
	// ErrUnknowAction is returned when the action is unknown
	ErrUnknowAction = errors.New("unknow action")
)

// Reader allow to retrieve an action from a given payload
type Reader interface {
	Read(buf []byte) (Action, error)
}

type reader struct {
	data []byte
}

var _ Reader = &reader{}
var _ Reader = &DummyReader{}

// NewReader creates a new Reader
func NewReader() Reader {
	return &reader{}
}

// Read attempt to read an action from given payload
func (ar *reader) Read(buf []byte) (Action, error) {
	r := bytes.NewReader(buf)
	actionByte, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	var action Action

	switch ID(actionByte) {
	case AnnounceID:
		action = &Announce{}
	default:
		return nil, ErrUnknowAction
	}

	if err := binary.Read(r, binary.BigEndian, action); err != nil {
		return nil, err
	}

	return action, nil
}

// DummyReader provides a configurable Reader
type DummyReader struct {
	ReadFunc func(buf []byte) (Action, error)
}

// Read calls ReadFunc
func (d *DummyReader) Read(buf []byte) (Action, error) {
	return d.ReadFunc(buf)
}
