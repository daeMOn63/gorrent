package fs

import (
	"errors"
	"io"
	"reflect"
	"testing"
)

func TestDummyFile(t *testing.T) {
	t.Run("Name should return NameVal", func(t *testing.T) {
		expectedName := "some-name"

		d := &DummyFile{
			NameVal: expectedName,
		}

		if d.Name() != expectedName {
			t.Fatalf("Expected name to be %s, got %s", expectedName, d.Name())
		}
	})

	t.Run("Close should return CloseErr", func(t *testing.T) {
		expectedErr := errors.New("closerr-string")

		d := &DummyFile{
			CloseErr: expectedErr,
		}

		if d.Close() != expectedErr {
			t.Fatalf("Expected err to be %s, got %s", expectedErr, d.Close())
		}
	})

	t.Run("Stat should return StatVal and StatErr", func(t *testing.T) {
		expectedVal := &DummyFileInfo{}
		expectedErr := errors.New("staterr-string")

		d := &DummyFile{
			StatErr: expectedErr,
			StatVal: expectedVal,
		}

		info, err := d.Stat()

		if reflect.DeepEqual(info, expectedVal) == false {
			t.Fatalf("Expected info to be %#v, got %#v", expectedVal, info)
		}

		if err != expectedErr {
			t.Fatalf("Expected err to be %s, got %s", expectedErr, err)
		}
	})

	t.Run("Read should return content and eof when done", func(t *testing.T) {
		d := &DummyFile{
			Content: []byte("abcdef"),
		}

		expectedLength := 2
		buf := make([]byte, expectedLength)
		n, err := d.Read(buf)

		if err != nil {
			t.Fatalf("Expected err to be nil, got %s", err)
		}

		if n != expectedLength {
			t.Fatalf("Expected n to be %d, got %d", expectedLength, n)
		}

		expectedBuf := []byte("ab")
		if reflect.DeepEqual(buf, expectedBuf) == false {
			t.Fatalf("Expected buf to be %v, got %v", expectedBuf, buf)
		}

		expectedLength = len(d.Content) - expectedLength
		buf = make([]byte, 6)
		n, err = d.Read(buf)

		if err != nil {
			t.Fatalf("Expected err to be nil, got %s", err)
		}

		if n != expectedLength {
			t.Fatalf("Expected n to be %d, got %d", expectedLength, n)
		}

		expectedBuf = []byte("cdef")
		if reflect.DeepEqual(buf[:len(expectedBuf)], expectedBuf) == false {
			t.Fatalf("Expected buf to be %v, got %v", expectedBuf, buf)
		}

		buf = make([]byte, 2)
		n, err = d.Read(buf)
		if err != io.EOF {
			t.Fatalf("Expected err to be EOF, got %s", err)
		}
		if n != 0 {
			t.Fatalf("Expected n to be 0, got %d", n)
		}
		expectedBuf = make([]byte, 2)
		if reflect.DeepEqual(buf, expectedBuf) == false {
			t.Fatalf("Expected buf to be %v, got %v", expectedBuf, buf)
		}
	})

	t.Run("Read should return ReadErr when set", func(t *testing.T) {

		expectedErr := errors.New("readerr-string")
		d := DummyFile{
			ReadErr: expectedErr,
		}

		buf := make([]byte, 1)
		n, err := d.Read(buf)
		if n != 0 {
			t.Fatalf("Expected n to be 0, got %d", n)
		}

		expectedBuf := make([]byte, 1)
		if reflect.DeepEqual(buf, expectedBuf) == false {
			t.Fatalf("Expected buf to be %v, got %v", expectedBuf, buf)
		}

		if err != expectedErr {
			t.Fatalf("Expected err to be %s, got %s", expectedErr, err)
		}

	})
}
