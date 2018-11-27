package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/daeMOn63/gorrent/buffer"
	"github.com/daeMOn63/gorrent/fs"
	"github.com/daeMOn63/gorrent/gorrent"
	"github.com/daeMOn63/gorrent/peer"
)

// PublicServer defines a gorrent peer public server, used to handle connections from other peers.
type PublicServer struct {
	peer  gorrent.Peer
	fs    fs.FileSystem
	store peer.GorrentStore
}

// NewPublicServer creates a new peer public server
func NewPublicServer(peer gorrent.Peer, fs fs.FileSystem, store peer.GorrentStore) *PublicServer {
	return &PublicServer{
		fs:    fs,
		store: store,
	}
}

// Listen start listening for peer requests
func (s *PublicServer) Listen() error {
	pieceReader := buffer.NewPieceReader(s.fs)

	addr := fmt.Sprintf("%s:%d", s.peer.PeerAddr, s.peer.Port)
	link, err := net.ListenPacket("udp", addr)
	if err != nil {
		return err
	}
	defer link.Close()

	log.Printf("Peer server listening on %s", addr)

	for {
		buf := make([]byte, peer.MaxUDPPacketSize)
		n, client, err := link.ReadFrom(buf)
		if err != nil {
			log.Println("Error while reading from link: ", err)

			continue
		}

		r := bytes.NewReader(buf[:n])
		chunkRequest, err := readChunkRequest(r)
		if err != nil {
			log.Println(err)

			continue
		}

		log.Printf("%s requested chunk %d from %s", client, chunkRequest.ChunkID, chunkRequest.InfoHash.HexString())

		entry, err := s.store.Get(chunkRequest.InfoHash)
		if err != nil {
			log.Println(err)

			continue
		}

		data, err := pieceReader.ReadPiece(entry.Path, entry.Gorrent.Files, chunkRequest.ChunkID, entry.Gorrent.PieceLength)
		if err != nil {
			log.Println(err)

			continue
		}

		log.Printf("Sending %s (%s) chunk %d to %s", chunkRequest.InfoHash.HexString(), entry.Name, chunkRequest.ChunkID, client)

		data = data[:entry.Gorrent.PieceLength]

		var current int
		for current < len(data) {
			packet := data[current:(current + peer.MaxUDPPacketSize)]
			_, err = link.WriteTo(packet, client)
			if err != nil {
				log.Printf("write error: %s", err)

				break
			}

			current += peer.MaxUDPPacketSize
		}

	}
}

// ReadChunkRequest reads a ChunkRequest struct from given Reader
func readChunkRequest(r io.Reader) (*peer.ChunkRequest, error) {
	cr := &peer.ChunkRequest{}

	if err := binary.Read(r, binary.BigEndian, cr); err != nil {
		return nil, err
	}

	return cr, nil
}
