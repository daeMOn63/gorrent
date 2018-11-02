package peer

import (
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

// Server defines methods for a peer Server
type Server interface {
	Listen() error
}

type server struct {
	config *Config
	fs     fs.FileSystem
}

var _ Server = &server{}

// NewServer creates a new peer server
func NewServer(config *Config, fs fs.FileSystem) Server {
	return &server{
		config: config,
		fs:     fs,
	}
}

// Listen start listening on the unix socket for requests
func (s *server) Listen() error {

	if err := s.fs.MkdirAll(filepath.Dir(s.config.SockPath), 0700); err != nil {
		return err
	}

	store, err := NewStore(s.config.DbPath, 0600)
	if err != nil {
		return err
	}

	fileBuffer := buffer.NewFile(s.fs, s.config.TmpPath)
	gorrentReadWriter := gorrent.NewReadWriter()

	var peerID gorrent.PeerID
	peerID.SetString(s.config.ID)

	peer := gorrent.NewPeer(s.config.ID, s.config.PublicIP, s.config.PublicPort)
	tracker := tracker.NewClient(*peer, s.config.TrackerProtocol)

	handler := NewHTTPHandler(store, gorrentReadWriter)

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

	watcher := NewWatcher(store, s.fs, fileBuffer, tracker, time.Duration(s.config.AnnounceDelay)*time.Millisecond)
	go watcher.Watch()

	return localServer.Serve(conn)
}
