package gorrent

import "testing"

func TestGorrent(t *testing.T) {
	t.Run("TotalFileSize calculate the correct size", func(t *testing.T) {
		g := &Gorrent{}

		if g.TotalFileSize() != 0 {
			t.Fatalf("Expected total file size to be 0, got %d", g.TotalFileSize())
		}

		g.Files = append(g.Files, File{
			Length: 1234,
		})

		if g.TotalFileSize() != 1234 {
			t.Fatalf("Expected total file size to be 1234, got %d", g.TotalFileSize())
		}

		g.Files = append(g.Files, File{
			Length: 4321,
		})

		if g.TotalFileSize() != 5555 {
			t.Fatalf("Expected total file size to be 5555, got %d", g.TotalFileSize())
		}
	})
}
