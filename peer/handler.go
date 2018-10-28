package peer

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/daeMOn63/gorrent/gorrent"

	"github.com/dustin/go-humanize"
)

const maxUploadSize = 1000 * 1024 // 1 MB

var (
	// ErrPathRequired is the error returned when the path parameter is missing on the request
	ErrPathRequired = errors.New("path is required")
)

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

	file, fileHeaders, err := r.FormFile("gorrent")
	if err != nil {
		writeError(w, err, http.StatusBadRequest)

		return
	}
	defer file.Close()

	gorrent, err := h.readWriter.Read(file)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	path := r.Form.Get("path")
	if path == "" {
		writeError(w, ErrPathRequired, http.StatusBadRequest)
		return
	}

	g := &GorrentEntry{
		Name:       fileHeaders.Filename,
		Gorrent:    gorrent,
		CreatedAt:  time.Now(),
		Path:       path,
		Uploaded:   0,
		Downloaded: 0,
		Status:     StatusNew,
	}

	if err := h.gorrentStore.Save(g); err != nil {
		writeError(w, err, http.StatusInternalServerError)

		return
	}

	writeSuccess(w, string(gorrent.InfoHash().HexString()))
}

// Remove allow to remove an existing gorrent from the server
func (h *HTTPHandler) Remove(w http.ResponseWriter, r *http.Request) {

}

// Info returns information about a given gorrent
func (h *HTTPHandler) Info(w http.ResponseWriter, r *http.Request) {

}

type listEntry struct {
	InfoHash  string `json:"infoHash"`
	Name      string `json:"name"`
	Size      string `json:"size"`
	CreatedAt string `json:"createdAt"`
	Completed string `json:"completed"`
	Status    string `json:"status"`
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
			InfoHash:  g.Gorrent.InfoHash().HexString(),
			Name:      g.Name,
			Size:      humanize.Bytes(g.Gorrent.TotalFileSize()),
			CreatedAt: humanize.Time(g.CreatedAt),
			Completed: fmt.Sprintf("%d %%", (g.Downloaded*100)/g.Gorrent.TotalFileSize()),
			Status:    string(g.Status),
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

	jsonEncode(w, r)
}

func writeError(w http.ResponseWriter, e error, status int) {
	r := &Response{
		Status:  status,
		Message: e.Error(),
	}

	jsonEncode(w, r)
}

func jsonEncode(w http.ResponseWriter, response *Response) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")

	if err := enc.Encode(response); err != nil {
		log.Printf("jsonEncode error: %s", err)
	}
}
