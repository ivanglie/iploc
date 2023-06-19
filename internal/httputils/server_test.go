package httputils

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockHTTPServer struct{}

func (m *mockHTTPServer) ListenAndServe() error {
	return nil
}

func (m *mockHTTPServer) ListenAndServeTLS(certFile string, keyFile string) error {
	return nil
}

func (m *mockHTTPServer) Shutdown(ctx context.Context) error {
	return nil
}

type errorHTTPServer struct{}

func (m *errorHTTPServer) ListenAndServe() error {
	return assert.AnError
}

func (m *errorHTTPServer) ListenAndServeTLS(certFile string, keyFile string) error {
	return assert.AnError
}

func (m *errorHTTPServer) Shutdown(ctx context.Context) error {
	return assert.AnError
}

var (
	mockHandler *http.ServeMux
)

func setUp() {
	mockHandler = http.NewServeMux()
	mockHandler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func tearDown() {
	mockHandler = nil
}

func TestMain(m *testing.M) {
	setUp()

	exitCode := m.Run()

	tearDown()

	os.Exit(exitCode)
}

func TestServer_ListerAndServe_HTTP(t *testing.T) {
	s := NewServer(mockHandler, false, "", false)
	s.httpServer = &mockHTTPServer{}
	assert.NoError(t, s.ListenAndServe())

	// Error
	s = NewServer(mockHandler, false, "", false)
	s.httpServer = &errorHTTPServer{}
	assert.Error(t, s.ListenAndServe())
}

func TestServer_ListerAndServe_HTTPS(t *testing.T) {
	// Error
	s := NewServer(mockHandler, true, "example.com", true)
	s.httpServer = &errorHTTPServer{}
	s.httpsServer = &errorHTTPServer{}

	assert.Error(t, s.ListenAndServe())
}

func TestServer_Shutdown_HTTP(t *testing.T) {
	s := NewServer(mockHandler, true, "example.com", true)

	s.httpServer = &mockHTTPServer{}
	s.httpsServer = &mockHTTPServer{}

	assert.NoError(t, s.Shutdown(context.TODO()))

	// Error
	s = NewServer(mockHandler, true, "example.com", true)
	s.httpServer = &errorHTTPServer{}
	s.httpsServer = &errorHTTPServer{}

	assert.Error(t, s.Shutdown(context.TODO()))

	// Have no any servers
	s = &Server{}
	assert.NoError(t, s.Shutdown(context.TODO()))
}

func TestServer_createHTTPS(t *testing.T) {
	s := NewServer(mockHandler, true, "example.com", true)
	assert.NotNil(t, s.httpServer)
	assert.NotNil(t, s.httpsServer)
	assert.Equal(t, mockHandler, s.defaultHandler)

	var httpInterface interface{} = s.httpServer
	v, ok := httpInterface.(*http.Server)
	assert.True(t, ok)
	assert.NotNil(t, s.httpServer)
	assert.Equal(t, ":http", v.Addr)
	assert.Equal(t, 5*time.Second, v.ReadHeaderTimeout)
	assert.Equal(t, 30*time.Second, v.IdleTimeout)

	var httpsInterface interface{} = s.httpsServer
	v, ok = httpsInterface.(*http.Server)
	assert.True(t, ok)
	assert.NotNil(t, s.httpsServer)
	assert.Equal(t, ":https", v.Addr)
	assert.Equal(t, 5*time.Second, v.ReadHeaderTimeout)
	assert.Equal(t, 30*time.Second, v.IdleTimeout)

	// Empty host error
	s = NewServer(mockHandler, true, "", true)
	assert.Nil(t, s)
}

func TestServer_createHTTP(t *testing.T) {
	s := NewServer(mockHandler, false, "", false)
	assert.NotNil(t, s.httpServer)
	assert.False(t, s.useStaging)

	var i interface{} = s.httpServer
	v, ok := i.(*http.Server)
	assert.True(t, ok)
	assert.Equal(t, ":http", v.Addr)
	assert.Equal(t, mockHandler, v.Handler)
	assert.Equal(t, 5*time.Second, v.ReadHeaderTimeout)
	assert.Equal(t, 30*time.Second, v.IdleTimeout)
}

func TestServer_String(t *testing.T) {
	// HTTP
	s := NewServer(mockHandler, false, "", false)
	assert.Equal(t, "Server{:http, SSL: false, Staging: false}", s.String())

	// HTTPS
	s = NewServer(mockHandler, true, "example.com", true)
	assert.Equal(t, "Server{:http, SSL: true, :https, CacheDir: certs, Staging: true}", s.String())
}
