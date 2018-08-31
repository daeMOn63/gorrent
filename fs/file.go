package fs

import (
	"io"
	"os"
	"time"
)

// File interface describe methods available on a file
type File interface {
	Name() string
	Stat() (os.FileInfo, error)
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	Close() error
}

// DummyFileInfo implements os.FileInfo and allow to configure its behavior
type DummyFileInfo struct {
	SizeVal    int64
	NameVal    string
	ModeVal    os.FileMode
	ModTimeVal time.Time
	IsDirVal   bool
	SysVal     interface{}
}

// Size returns the DummyFileInfo SizeVal
func (dfi *DummyFileInfo) Size() int64 {
	return dfi.SizeVal
}

// Name returns NameVal
func (dfi *DummyFileInfo) Name() string {
	return dfi.NameVal
}

// Mode returns ModeVal
func (dfi *DummyFileInfo) Mode() os.FileMode {
	return dfi.ModeVal
}

// ModTime returns ModTimeVal
func (dfi *DummyFileInfo) ModTime() time.Time {
	return dfi.ModTimeVal
}

// IsDir returns IsDirVal
func (dfi *DummyFileInfo) IsDir() bool {
	return dfi.IsDirVal
}

// Sys returns SysVal
func (dfi *DummyFileInfo) Sys() interface{} {
	return dfi.SysVal
}

// DummyFile implement File and allow to configure its behavior
type DummyFile struct {
	NameVal    string
	Content    []byte
	StatVal    os.FileInfo
	StatErr    error
	CurReadPtr int
	ReadErr    error
	WriteErr   error
	CloseErr   error
}

// Name returns dummyFile NameVal
func (f *DummyFile) Name() string {
	return f.NameVal
}

// Stat returns dummyFile StatVal
func (f *DummyFile) Stat() (os.FileInfo, error) {
	return f.StatVal, f.StatErr
}

// Read reads up to len(p) bytes from DummyFile Content, and advance internal cursor.
// io.EOF is returned when there is nothing more to read
// An error can be returned by setting the ReadErr field of the DummyFile.
func (f *DummyFile) Read(p []byte) (int, error) {

	if f.ReadErr != nil {
		return 0, f.ReadErr
	}

	if f.CurReadPtr >= len(f.Content) {
		return 0, io.EOF
	}

	maxRead := len(p)
	maxAvailable := len(f.Content[f.CurReadPtr:])
	if maxRead > maxAvailable {
		maxRead = maxAvailable
	}

	copy(p, f.Content[f.CurReadPtr:f.CurReadPtr+maxRead])
	f.CurReadPtr += maxRead
	return maxRead, nil
}

// Write append content to the DummyFile Content field
// An error can be returned by setting the WriteErr field
func (f *DummyFile) Write(p []byte) (int, error) {
	if f.WriteErr != nil {
		return 0, f.WriteErr
	}

	f.Content = append(f.Content, p...)
	return len(p), nil
}

// Close returns CloseErr field of DummyFile
func (f *DummyFile) Close() error {
	return f.CloseErr
}
