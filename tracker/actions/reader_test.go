package actions

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"reflect"
	"testing"
)

func TestReader(t *testing.T) {
	t.Run("Readers returns EOF error when there is nothing to read", func(t *testing.T) {
		r := NewReader()

		a, err := r.Read(nil)
		if err == nil {
			t.Fatal("An error was expected")
		}
		if a != nil {
			t.Fatalf("Expected nil action, got %#v", a)
		}
	})

	t.Run("Readers return ErrUnknowActions when the action ID is unknown", func(t *testing.T) {
		r := NewReader()
		a, err := r.Read([]byte{0})
		if err != ErrUnknowAction {
			t.Fatalf("Expected err to be %s, got %s", ErrUnknowAction, err)
		}
		if a != nil {
			t.Fatalf("Expected nil action, got %#v", a)
		}
	})

	t.Run("Readers properly reads Announce actions", func(t *testing.T) {
		r := NewReader()

		expectedAction := &Announce{
			InfoHash: sha1.Sum([]byte("foobar")),
		}

		buf := bytes.NewBuffer(nil)
		buf.WriteByte(byte(expectedAction.ID()))
		binary.Write(buf, binary.BigEndian, expectedAction)

		a, err := r.Read(buf.Bytes())
		if err != nil {
			t.Fatalf("Expected err to be nil, got %s", err)
		}

		if reflect.DeepEqual(a, expectedAction) == false {
			t.Fatalf("Expected action to be %#v, got %#v", expectedAction, a)
		}
	})
}

func TestDummyReader(t *testing.T) {
	t.Run("Read calls ReadFunc", func(t *testing.T) {

		expectedBuf := []byte("abcd")
		expectedAction := &Announce{}
		expectedError := errors.New("readfunc-err")

		d := &DummyReader{
			ReadFunc: func(buf []byte) (Action, error) {
				if reflect.DeepEqual(buf, expectedBuf) == false {
					t.Fatalf("Expected buf to be %v, got %v", expectedBuf, buf)
				}
				return expectedAction, expectedError
			},
		}

		a, err := d.Read(expectedBuf)
		if reflect.DeepEqual(a, expectedAction) == false {
			t.Fatalf("Expected action to be %#v, got %#v", expectedAction, a)
		}

		if err != expectedError {
			t.Fatalf("Expected err to be %s, got %s", expectedError, err)
		}
	})
}
