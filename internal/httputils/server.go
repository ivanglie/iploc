package httputils

import (
	"crypto/tls"
	"net/http"
	"time"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

const (
	letsEncryptStagingURL = "https://acme-staging-v02.api.letsencrypt.org/directory"
	dir                   = "certs"
)

type Interface interface {
	ListenAndServe() error
	ListenAndServeTLS(certFile string, keyFile string) error
}

type Server struct {
	httpServer  Interface
	httpsServer Interface
	UseSSL      bool
	Host        string
	UseDebug    bool
	Handler     http.Handler
}

// ListenAndServe starts a http server.
func (s *Server) ListenAndServe() error {
	if s.UseSSL {
		return s.listenAndServeTLS()
	}

	return s.listenAndServe()
}

// ListenAndServe starts a http server without TLS.
func (s *Server) listenAndServe() error {
	s.httpServer = newHTTPServer(s.Handler)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

// ListenAndServeTLS starts a http server with TLS.
// If UseDebug is true, it uses the staging server.
func (s *Server) listenAndServeTLS() error {
	autocertManager := newAutocertManager(s.Host)

	if s.UseDebug {
		autocertManager.Client = &acme.Client{DirectoryURL: letsEncryptStagingURL}
	}

	errCh := make(chan error)
	go func() {
		s.httpServer = newHTTPServer(autocertManager.HTTPHandler(nil))
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	s.httpsServer = newHTTPSServer(s.Handler, autocertManager.TLSConfig())
	if err := s.httpsServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
		return err
	}

	return <-errCh
}

// newHTTPSServer creates a HTTPS server.
func newHTTPSServer(h http.Handler, tlsConfig *tls.Config) *http.Server {
	return &http.Server{
		Addr:              ":https",
		Handler:           h,
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
}

// newHTTPServer creates a HTTP server.
func newHTTPServer(h http.Handler) *http.Server {
	return &http.Server{
		Addr:              ":http",
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
}

// newAutocertManager creates a new ACME autocert manager.
func newAutocertManager(hosts ...string) autocert.Manager {
	return autocert.Manager{
		Cache:      autocert.DirCache(dir),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(hosts...),
	}
}
