package httputils

import (
	"net/http"
	"time"
)

type Server struct {
	Handler http.Handler
}

// ListenAndServe starts a http server.
func (s *Server) ListenAndServe() error {
	httpServer := &http.Server{
		Addr:         ":http",
		Handler:      s.Handler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return httpServer.ListenAndServe()
}
