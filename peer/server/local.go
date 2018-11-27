package server

import (
	"log"
	"net"
	"net/http"
	"path/filepath"

	"github.com/daeMOn63/gorrent/fs"
	"github.com/daeMOn63/gorrent/gorrent"
	"github.com/daeMOn63/gorrent/peer"
	"github.com/daeMOn63/gorrent/peer/handlers"
	"github.com/gorilla/mux"
)

// LocalServer defines a peer local server, used for adding new gorrents, or getting their status.
type LocalServer struct {
	sockPath string
	fs       fs.FileSystem
	store    peer.GorrentStore
}

// NewLocalServer creates a new peer local server
func NewLocalServer(sockPath string, fs fs.FileSystem, store peer.GorrentStore) *LocalServer {
	return &LocalServer{
		sockPath: sockPath,
		fs:       fs,
		store:    store,
	}
}

// Listen start listening for peer requests
func (s *LocalServer) Listen() error {
	if err := s.fs.MkdirAll(filepath.Dir(s.sockPath), 0700); err != nil {
		return err
	}

	gorrentReadWriter := gorrent.NewReadWriter()
	handler := handlers.NewLocalHTTP(s.store, gorrentReadWriter)

	router := mux.NewRouter()
	router.HandleFunc("/add", handler.Add).Methods("POST")
	router.HandleFunc("/remove/{hash}", handler.Remove).Methods("GET")
	router.HandleFunc("/info/{hash}", handler.Info).Methods("GET")
	router.HandleFunc("/", handler.List).Methods("GET")

	logger := peer.NewLoggerMiddleware()
	router.Use(logger.Handle)

	localServer := http.Server{
		Handler: router,
	}

	s.fs.Remove(s.sockPath)
	conn, err := net.Listen("unix", s.sockPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Start local peer server
	log.Printf("local server listening on unix://%s", s.sockPath)

	return localServer.Serve(conn)
}
