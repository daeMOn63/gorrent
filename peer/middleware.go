package peer

import (
	"log"
	"net/http"
)

// LoggerMiddleware hold the logger handler
type LoggerMiddleware struct{}

// NewLoggerMiddleware creates a new Logger middleware
func NewLoggerMiddleware() *LoggerMiddleware {
	return &LoggerMiddleware{}
}

// Handle log the request and call next
func (l *LoggerMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

		next.ServeHTTP(w, r)
	})
}
