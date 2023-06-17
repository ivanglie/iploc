package httputils

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/acme/autocert"
)

func Test_newHTTPSServer(t *testing.T) {
	handler := http.NewServeMux()
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := newHTTPSServer(handler, nil)
	assert.NotNil(t, server)
	assert.Equal(t, ":https", server.Addr)
	assert.Equal(t, handler, server.Handler)
	assert.Nil(t, server.TLSConfig)
	assert.Equal(t, 5*time.Second, server.ReadHeaderTimeout)
	assert.Equal(t, 30*time.Second, server.IdleTimeout)
}

func Test_newHTTPServer(t *testing.T) {
	handler := http.NewServeMux()
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := newHTTPServer(handler)
	assert.NotNil(t, server)
	assert.Equal(t, ":http", server.Addr)
	assert.Equal(t, handler, server.Handler)
	assert.Equal(t, 5*time.Second, server.ReadHeaderTimeout)
	assert.Equal(t, 30*time.Second, server.IdleTimeout)
}

func Test_newAutocertManager(t *testing.T) {
	autocertManager := newAutocertManager("example.com")
	assert.Equal(t, autocert.DirCache(dir), autocertManager.Cache)
	assert.NotNil(t, autocertManager.Prompt)
	assert.NoError(t, autocertManager.HostPolicy(context.TODO(), "example.com"))
}
