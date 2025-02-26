package http

import (
	"context"
	"net/http"
)

type Interface interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

type Server struct {
	httpServer Interface
}

// NewServer creates a new Server.
func NewServer(addr string, handler http.Handler) *Server {
	s := &Server{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
	}

	return s
}

// ListenAndServe starts a https server.
func (s *Server) ListenAndServe() error {
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

// Shutdown gracefully shuts down the server without interrupting any active connections.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.Shutdown(ctx)
}
