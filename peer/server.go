package peer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"time"

	"github.com/daeMOn63/gorrent/tracker"

	"github.com/daeMOn63/gorrent/buffer"
	"github.com/daeMOn63/gorrent/fs"
	"github.com/daeMOn63/gorrent/gorrent"

	"github.com/gorilla/mux"
)

const (
	// MaxUDPPacketSize defines the maximum size of udp packets
	MaxUDPPacketSize = 1024
)

// Server defines methods for a peer Server
type Server interface {
	Listen() error
}

type server struct {
	config *Config
	fs     fs.FileSystem
	store  GorrentStore
}

var _ Server = &server{}

// NewServer creates a new peer server
func NewServer(config *Config, fs fs.FileSystem, store GorrentStore) Server {
	return &server{
		config: config,
		fs:     fs,
		store:  store,
	}
}

// Listen start listening on the unix socket for requests
func (s *server) Listen() error {

	if err := s.fs.MkdirAll(filepath.Dir(s.config.SockPath), 0700); err != nil {
		return err
	}

	fileBuffer := buffer.NewFile(s.fs, s.config.TmpPath)
	gorrentReadWriter := gorrent.NewReadWriter()

	var peerID gorrent.PeerID
	peerID.SetString(s.config.ID)

	peer := gorrent.NewPeer(s.config.ID, s.config.PublicIP, s.config.PublicPort)
	tracker := tracker.NewClient(*peer, s.config.TrackerProtocol)
	peerClient := NewClient(2 * time.Second)

	handler := NewHTTPHandler(s.store, gorrentReadWriter)

	router := mux.NewRouter()
	router.HandleFunc("/add", handler.Add).Methods("POST")
	router.HandleFunc("/remove/{hash}", handler.Remove).Methods("GET")
	router.HandleFunc("/info/{hash}", handler.Info).Methods("GET")
	router.HandleFunc("/", handler.List).Methods("GET")

	logger := NewLoggerMiddleware()
	router.Use(logger.Handle)

	localServer := http.Server{
		Handler: router,
	}

	s.fs.Remove(s.config.SockPath)
	conn, err := net.Listen("unix", s.config.SockPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	announcer := NewAnnouncer(s.store, tracker, time.Duration(s.config.AnnounceDelay)*time.Millisecond)
	go announcer.AnnounceForever()

	// Start watcher
	watcher := NewWatcher(s.store, s.fs, fileBuffer, tracker, peerClient)
	go func() {
		if err := watcher.Watch(); err != nil {
			log.Println("watcher error: ", err)
		}
	}()

	// Start public peer server
	go func() {
		if err := s.listenPeer(); err != nil {
			log.Println("listenPeer error: ", err)
		}
	}()

	// Start local peer server
	return localServer.Serve(conn)
}

// ChunkRequest defines the data transfered on a chunkRequest
type ChunkRequest struct {
	InfoHash gorrent.Sha1Hash
	ChunkID  int64
}

// ReadChunkRequest reads a ChunkRequest struct from given Reader
func ReadChunkRequest(r io.Reader) (*ChunkRequest, error) {
	cr := &ChunkRequest{}

	if err := binary.Read(r, binary.BigEndian, cr); err != nil {
		return nil, err
	}

	return cr, nil
}

func (s *server) listenPeer() error {
	pieceReader := buffer.NewPieceReader(s.fs)

	addr := fmt.Sprintf("%s:%d", s.config.PublicIP, s.config.PublicPort)
	link, err := net.ListenPacket("udp", addr)
	if err != nil {
		return err
	}
	defer link.Close()

	log.Printf("Peer server listening on %s", addr)

	for {
		buf := make([]byte, MaxUDPPacketSize)
		n, client, err := link.ReadFrom(buf)
		if err != nil {
			log.Println("Error while reading from link: ", err)

			continue
		}

		r := bytes.NewReader(buf[:n])
		chunkRequest, err := ReadChunkRequest(r)
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
			packet := data[current:(current + MaxUDPPacketSize)]
			_, err = link.WriteTo(packet, client)
			if err != nil {
				log.Printf("write error: %s", err)

				break
			}

			current += MaxUDPPacketSize
		}

	}
}
