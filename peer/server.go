package peer

import (
	"net"
	"net/http"
	"path/filepath"

	"github.com/daeMOn63/gorrent/fs"

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

	hander := NewHTTPHandler()

	router := mux.NewRouter()
	router.HandleFunc("/add", hander.Add).Methods("POST")
	router.HandleFunc("/remove/{hash}", hander.Remove).Methods("GET")
	router.HandleFunc("/info/{hash}", hander.Info).Methods("GET")
	router.HandleFunc("/", hander.List).Methods("GET")

	server := http.Server{
		Handler: router,
	}

	s.fs.Remove(s.config.SockPath)
	conn, err := net.Listen("unix", s.config.SockPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	return server.Serve(conn)
}
