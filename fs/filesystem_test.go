package fs

import (
	"reflect"
	"testing"
)

func TestFileSystemFindFiles(t *testing.T) {
	filesystem := NewFileSystem()

	t.Run("FindFiles must fail with 0 workers", func(t *testing.T) {
		var err error
		_, err = filesystem.FindFiles("../test/sample1/data", 0)
		if err == nil {
			t.Fatalf("an error was expected")
		}

		_, err = filesystem.FindFiles("../test/sample1/data", -1)
		if err == nil {
			t.Fatalf("an error was expected")
		}
	})

	t.Run("FindFiles must fail with a invalid path", func(t *testing.T) {
		_, err := filesystem.FindFiles("./unknow", 1)
		if err == nil {
			t.Fatalf("an error was expected")
		}
	})

	t.Run("FindFiles must return list of filepaths", func(t *testing.T) {
		files, err := filesystem.FindFiles("../test/sample1/data", 4)
		expectedFileCount := 3
		if err != nil {
			t.Fatalf("unexpected err: %s", err)
		}

		if len(files) != expectedFileCount {
			t.Fatalf("expected %d files, got %d", expectedFileCount, len(files))
		}

		expectedFiles := []string{"/lorem.txt", "/pic/kittens/kittens.jpg", "/pic/kittens/kittens_files/kittens2.jpeg"}
		if reflect.DeepEqual(files, expectedFiles) == false {
			t.Fatalf("Expected files %v, got %v", expectedFiles, files)
		}
	})
}
