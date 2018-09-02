package fs

import (
	"errors"
	"io"
	"os"
	"reflect"
	"testing"
	"time"
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

	t.Run("Write should append content", func(t *testing.T) {
		d := &DummyFile{}

		buf := []byte("abcd")
		n, err := d.Write(buf)

		if err != nil {
			t.Fatalf("Expected err to be nil, got %s", err)
		}

		if n != len(buf) {
			t.Fatalf("Expected n to be %d, got %d", len(buf), n)
		}

		if reflect.DeepEqual(d.Content, buf) == false {
			t.Fatalf("Expected content to be %v, got %v", buf, d.Content)
		}
	})

	t.Run("Write should return WriteErr when set", func(t *testing.T) {
		expectedErr := errors.New("writeerr-string")
		d := &DummyFile{
			WriteErr: expectedErr,
		}

		buf := []byte("abcd")
		n, err := d.Write(buf)

		if n != 0 {
			t.Fatalf("Expected n to be 0, got %d", n)
		}

		if err != expectedErr {
			t.Fatalf("Expected err to be %s, got %s", expectedErr, err)
		}

		if len(d.Content) != 0 {
			t.Fatalf("Expected content to be empty, got %#v", d.Content)
		}
	})
}

func TestDummyFileInfo(t *testing.T) {
	t.Run("DummyFileInfo should return configured data", func(t *testing.T) {

		expectedIsDir := true
		expectedMode := os.FileMode(0777)
		expectedName := "foobar"
		expectedTime := time.Now()
		expectedSize := int64(42)
		expectedSys := make(map[string]interface{})

		d := DummyFileInfo{
			IsDirVal:   expectedIsDir,
			ModeVal:    expectedMode,
			NameVal:    expectedName,
			ModTimeVal: expectedTime,
			SizeVal:    expectedSize,
			SysVal:     expectedSys,
		}

		if d.IsDir() != expectedIsDir {
			t.Fatalf("Expected IsDir to be %v, got %v", expectedIsDir, d.IsDir())
		}

		if d.Mode() != expectedMode {
			t.Fatalf("Expected Mode to be %v, got %v", expectedMode, d.Mode())
		}

		if d.Name() != expectedName {
			t.Fatalf("Expected Name to be %v, got %v", expectedName, d.Name())
		}

		if d.ModTime() != expectedTime {
			t.Fatalf("Expected ModTime to be %v, got %v", expectedTime, d.ModTime())
		}

		if d.Size() != expectedSize {
			t.Fatalf("Expected size to be %v, got %v", expectedSize, d.Size())
		}

		if reflect.DeepEqual(d.Sys(), expectedSys) == false {
			t.Fatalf("Expected Sys to be %v, got %v", expectedSys, d.Sys())
		}
	})
}
