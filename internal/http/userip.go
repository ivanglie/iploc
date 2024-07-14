package http

import (
	"net"
	"net/http"
)

// UserIP returns the IP address of the user from the request.
func UserIP(r *http.Request) (string, string, error) {
	remoteIP := r.Header.Get("X-Forwarded-For")
	if remoteIP == "" {
		remoteIP = r.RemoteAddr
	}

	return net.SplitHostPort(remoteIP)
}
