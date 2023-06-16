package httputils

import (
	"net/http"
	"time"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

type Server struct {
	UseSSL   bool
	Host     string
	UseDebug bool
	Handler  http.Handler
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
	httpServer := &http.Server{
		Addr:         ":http",
		Handler:      s.Handler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return httpServer.ListenAndServe()
}

// ListenAndServeTLS starts a http server with TLS.
// If UseDebug is true, it uses the staging server.
func (s *Server) listenAndServeTLS() error {
	autocertManager := autocert.Manager{
		Cache:      autocert.DirCache("certs"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(s.Host),
	}

	if s.UseDebug {
		autocertManager.Client = &acme.Client{
			DirectoryURL: "https://acme-staging-v02.api.letsencrypt.org/directory",
		}
	}

	go func() {
		httpServer := &http.Server{
			Addr:         ":http",
			Handler:      autocertManager.HTTPHandler(nil),
			IdleTimeout:  time.Minute,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		}

		httpServer.ListenAndServe()
	}()

	httpServer := &http.Server{
		Addr:         ":https",
		Handler:      s.Handler,
		TLSConfig:    autocertManager.TLSConfig(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	err := httpServer.ListenAndServeTLS("", "")

	return err
}
