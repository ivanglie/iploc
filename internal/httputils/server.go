package httputils

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

const (
	// The ACME URL for our ACME v2 staging environment.
	// See https://letsencrypt.org/docs/staging-environment/ for more details.
	letsEncryptStagingURL = "https://acme-staging-v02.api.letsencrypt.org/directory"

	// The directory where the certificates are stored.
	dir = "certs"
)

type Interface interface {
	ListenAndServe() error
	ListenAndServeTLS(certFile string, keyFile string) error
	Shutdown(ctx context.Context) error
}

type Server struct {
	sync.RWMutex
	httpServer      Interface
	httpsServer     Interface
	defaultHandler  http.Handler
	useSSL          bool
	host            string
	autocertManager autocert.Manager
	useStaging      bool
}

// NewServer creates a new Server.
func NewServer(handler http.Handler, useSSL bool, host string, useStaging bool) *Server {
	s := &Server{
		defaultHandler: handler,
		useSSL:         useSSL,
		host:           host,
		useStaging:     useStaging,
	}

	if !s.useSSL {
		s.useStaging = false
		s.createHTTP(nil)
		return s
	}

	if len(s.host) == 0 {
		return nil
	}

	s.autocertManager = autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(host),
		Cache:      autocert.DirCache(dir),
	}

	s.createHTTP(s.autocertManager.HTTPHandler(nil)) // Redirects all http requests to https

	if !s.useStaging {
		return s
	}

	s.autocertManager.Client = &acme.Client{DirectoryURL: letsEncryptStagingURL}
	s.createHTTPS()

	return s
}

// ListenAndServe starts a https server.
func (s *Server) ListenAndServe() error {
	s.Lock()
	defer s.Unlock()

	if s.httpServer != nil {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}

		return nil
	}

	if s.httpsServer == nil {
		return errors.New("no server to start")
	}

	errCh := make(chan error)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	if err := s.httpsServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
		return err
	}

	return <-errCh
}

// Shutdown gracefully shuts down the server without interrupting any active connections.
func (s *Server) Shutdown(ctx context.Context) error {
	s.Lock()
	defer s.Unlock()

	for _, server := range []Interface{s.httpServer, s.httpsServer} {
		if server == nil {
			continue
		}

		if err := server.Shutdown(ctx); err != nil {
			return err
		}
	}

	return nil
}

// createHTTPS creates a HTTPS server.
func (s *Server) createHTTPS() error {
	s.Lock()
	defer s.Unlock()

	s.httpsServer = &http.Server{
		Addr:              ":https",
		Handler:           s.defaultHandler,
		TLSConfig:         s.autocertManager.TLSConfig(),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	return nil
}

// createHTTP creates a HTTP server.
// If h is nil, it uses the s.DefaultHandler.
func (s *Server) createHTTP(h http.Handler) {
	s.Lock()
	defer s.Unlock()

	handler := s.defaultHandler
	if h != nil {
		handler = h
	}

	s.httpServer = &http.Server{
		Addr:              ":http",
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
}

// String is the string representation of the server.
func (s *Server) String() string {
	s.RLock()
	defer s.RUnlock()

	var httpInterface interface{} = s.httpServer
	httpValue := httpInterface.(*http.Server)

	if s.httpsServer != nil && s.httpServer != nil {
		var httpsInterface interface{} = s.httpsServer
		httpsValue := httpsInterface.(*http.Server)

		return fmt.Sprintf("Server{%s, SSL: %v, %s, CacheDir: %v, Staging: %v}",
			httpValue.Addr, s.useSSL, httpsValue.Addr, s.autocertManager.Cache, s.useStaging)
	}

	return fmt.Sprintf("Server{%s, SSL: %v, Staging: %v}", httpValue.Addr, s.useSSL, s.useStaging)
}
