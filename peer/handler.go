package peer

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/daeMOn63/gorrent/gorrent"
)

const maxUploadSize = 1000 * 1024 // 1 MB

// HTTPHandler hold the handlers available on the peerd server
type HTTPHandler struct {
	gorrentStore GorrentStore
	readWriter   gorrent.ReadWriter
}

// NewHTTPHandler returns a new HTTPHandler
func NewHTTPHandler(gorrentStore GorrentStore, rw gorrent.ReadWriter) *HTTPHandler {
	return &HTTPHandler{
		gorrentStore: gorrentStore,
		readWriter:   rw,
	}
}

// Response describe the generic handler response format
type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Add allow to add a new gorrent to the local server
func (h *HTTPHandler) Add(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeError(w, err, http.StatusBadRequest)

		return
	}

	file, _, err := r.FormFile("gorrent")
	log.Print(file)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)

		return
	}
	defer file.Close()

	g, err := h.readWriter.Read(file)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	if err := h.gorrentStore.Save(g); err != nil {
		writeError(w, err, http.StatusInternalServerError)

		return
	}

	writeSuccess(w, string(g.InfoHash().HexString()))
}

// Remove allow to remove an existing gorrent from the server
func (h *HTTPHandler) Remove(w http.ResponseWriter, r *http.Request) {

}

// Info returns information about a given gorrent
func (h *HTTPHandler) Info(w http.ResponseWriter, r *http.Request) {

}

type listEntry struct {
	InfoHash string `json:"infoHash"`
	Size     int64  `json:"size"`
}

// List returns the list of gorrent
func (h *HTTPHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.gorrentStore.All()
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)

		return
	}

	var entries []listEntry

	for _, g := range list {

		e := listEntry{
			InfoHash: g.InfoHash().HexString(),
			Size:     g.TotalFileSize(),
		}

		entries = append(entries, e)
	}

	writeSuccess(w, entries)
}

func writeSuccess(w http.ResponseWriter, data interface{}) {
	r := &Response{
		Status: http.StatusOK,
		Data:   data,
	}

	enc := json.NewEncoder(w)
	if err := enc.Encode(r); err != nil {
		log.Printf("writeSuccess error: %s", err)
	}
}

func writeError(w http.ResponseWriter, e error, status int) {
	r := &Response{
		Status:  status,
		Message: e.Error(),
	}

	enc := json.NewEncoder(w)
	if err := enc.Encode(r); err != nil {
		log.Printf("writeError error: %s", err)
	}
}
