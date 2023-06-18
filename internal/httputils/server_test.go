package httputils

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var handler *http.ServeMux

func setUp() {
	handler = http.NewServeMux()
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func tearDown() {
	handler = nil
}

func TestMain(m *testing.M) {
	setUp()

	exitCode := m.Run()

	tearDown()

	os.Exit(exitCode)
}

func TestServer_ListerAndServe(t *testing.T) {
	// HTTP
	s := NewServer(handler, false, "", false)
	assert.NotNil(t, s.httpServer)
	assert.Nil(t, s.httpsServer)

	go s.ListenAndServe()
	s.Shutdown(context.TODO())

	// HTTPS
	s = NewServer(handler, true, "example.com", true)
	assert.NotNil(t, s.httpServer)
	assert.NotNil(t, s.httpsServer)

	go s.ListenAndServe()
	s.Shutdown(context.TODO())
}

func TestServer_createHTTPS(t *testing.T) {
	s := NewServer(handler, true, "example.com", true)
	assert.NotNil(t, s.httpServer)
	assert.NotNil(t, s.httpsServer)
	assert.Equal(t, handler, s.defaultHandler)

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
	s = NewServer(handler, true, "", true)
	assert.Nil(t, s)

}

func TestServer_createHTTP(t *testing.T) {
	s := NewServer(handler, false, "", false)
	assert.NotNil(t, s.httpServer)
	assert.False(t, s.useStaging)

	var i interface{} = s.httpServer
	v, ok := i.(*http.Server)
	assert.True(t, ok)
	assert.Equal(t, ":http", v.Addr)
	assert.Equal(t, handler, v.Handler)
	assert.Equal(t, 5*time.Second, v.ReadHeaderTimeout)
	assert.Equal(t, 30*time.Second, v.IdleTimeout)
}

func TestServer_String(t *testing.T) {
	// HTTP
	s := NewServer(handler, false, "", false)
	assert.Equal(t, "Server{:http, SSL: false, Staging: false}", s.String())

	// HTTPS
	s = NewServer(handler, true, "example.com", true)
	assert.Equal(t, "Server{:http, SSL: true, :https, CacheDir: certs, Staging: true}", s.String())
}
