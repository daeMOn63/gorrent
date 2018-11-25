package fs

import (
	"errors"
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
		expectedFileCount := 6
		if err != nil {
			t.Fatalf("unexpected err: %s", err)
		}

		if len(files) != expectedFileCount {
			t.Fatalf("expected %d files, got %d", expectedFileCount, len(files))
		}

		expectedFiles := []string{"/lorem.txt", "/pic", "/pic/kittens", "/pic/kittens/kittens.jpg", "/pic/kittens/kittens_files", "/pic/kittens/kittens_files/kittens2.jpeg"}
		if reflect.DeepEqual(files, expectedFiles) == false {
			t.Fatalf("Expected files %v, got %v", expectedFiles, files)
		}
	})
}

func TestDummyFS(t *testing.T) {
	t.Run("FindFiles should call FindFilesFunc", func(t *testing.T) {
		expectedFiles := []string{"a", "b"}
		expectedErr := errors.New("findfileserr-string")
		expectedPath := "some/path"
		expectedWorkers := 1

		fs := &DummyFS{
			FindFilesFunc: func(path string, workers int) ([]string, error) {
				if path != expectedPath {
					t.Fatalf("Expected path to be %s, got %s", expectedPath, path)
				}

				if workers != expectedWorkers {
					t.Fatalf("Expected workers to be %d, got %d", expectedWorkers, workers)
				}

				return expectedFiles, expectedErr
			},
		}

		files, err := fs.FindFiles(expectedPath, expectedWorkers)

		if reflect.DeepEqual(files, expectedFiles) == false {
			t.Fatalf("Expected files to be %v, got %v", expectedFiles, files)
		}

		if err != expectedErr {
			t.Fatalf("Expected err to be %s, got %s", expectedErr, err)
		}
	})

	t.Run("Open should call OpenFunc", func(t *testing.T) {
		expectedPath := "some/path"
		expectedErr := errors.New("openerr-string")
		expectedFile := &DummyFile{}

		fs := &DummyFS{
			OpenFunc: func(path string) (File, error) {
				if path != expectedPath {
					t.Fatalf("Expected path to be %s, got %s", expectedPath, path)
				}

				return expectedFile, expectedErr
			},
		}

		file, err := fs.Open(expectedPath)

		if reflect.DeepEqual(file, expectedFile) == false {
			t.Fatalf("Expected file to be %#v, got %#v", expectedFile, file)
		}

		if err != expectedErr {
			t.Fatalf("Expected err to be %s, got %s", expectedErr, err)
		}
	})

	t.Run("Create should call CreateFunc", func(t *testing.T) {
		expectedPath := "some/path"
		expectedErr := errors.New("createerr-string")
		expectedFile := &DummyFile{}

		fs := &DummyFS{
			CreateFunc: func(path string) (File, error) {
				if path != expectedPath {
					t.Fatalf("Expected path to be %s, got %s", expectedPath, path)
				}

				return expectedFile, expectedErr
			},
		}

		file, err := fs.Create(expectedPath)

		if reflect.DeepEqual(file, expectedFile) == false {
			t.Fatalf("Expected file to be %#v, got %#v", expectedFile, file)
		}

		if err != expectedErr {
			t.Fatalf("Expected err to be %s, got %s", expectedErr, err)
		}
	})
}
