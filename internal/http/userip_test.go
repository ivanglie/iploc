package http

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserIP(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/foo", nil)

	// Test case 1: X-Forwarded-For header is set
	req.Header.Set("X-Forwarded-For", "192.0.2.1:12345")
	ip, _, err := UserIP(req)
	assert.NoError(t, err)
	assert.Equal(t, "192.0.2.1", ip)

	// Test case 2: X-Forwarded-For header is not set
	req.Header.Del("X-Forwarded-For")
	req.RemoteAddr = "192.0.2.1:12345"
	ip, _, err = UserIP(req)
	assert.NoError(t, err)
	assert.Equal(t, "192.0.2.1", ip)
}
