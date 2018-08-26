package main

import (
	"fmt"
	"gorrent/fs"
	"gorrent/gorrent"
)

func main() {

	src := "test/sample1/data"
	out := "test/sample1/expected.gorrent"

	pb := gorrent.NewMemoryPieceBuffer(1024)
	filesystem := fs.NewFileSystem()
	creator := gorrent.NewCreator(pb, filesystem)

	g, err := creator.Create(src, 10)
	if err != nil {
		panic(err)
	}

	var totalSize int64
	for _, f := range g.Files {
		totalSize += f.Length
	}

	fmt.Println("totalSize: ", totalSize)

	expectedPieces := int(totalSize / int64(g.PieceLength))
	if totalSize%int64(g.PieceLength) != 0 {
		expectedPieces++
	}

	fmt.Println("Pieces:", len(g.Pieces), "ExpectedPieces:", expectedPieces)

	if err := creator.Save(out, g); err != nil {
		panic(err)
	}
	fmt.Println("saved")

	newG, err := creator.Open(out)
	if err != nil {
		panic(err)
	}
	fmt.Println(newG.PieceLength, len(newG.Files), len(newG.Pieces), err)
}
