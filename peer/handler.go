package peer

import (
	"net/http"
)

// HTTPHandler hold the handlers available on the peerd server
type HTTPHandler struct{}

// NewHTTPHandler returns a new HTTPHandler
func NewHTTPHandler() *HTTPHandler {
	return &HTTPHandler{}
}

// Add allow to add a new gorrent to the local server
func (h *HTTPHandler) Add(w http.ResponseWriter, r *http.Request) {
}

// Remove allow to remove an existing gorrent from the server
func (h *HTTPHandler) Remove(w http.ResponseWriter, r *http.Request) {

}

// Info returns information about a given gorrent
func (h *HTTPHandler) Info(w http.ResponseWriter, r *http.Request) {

}

// List returns the list of gorrent
func (h *HTTPHandler) List(w http.ResponseWriter, r *http.Request) {

}
