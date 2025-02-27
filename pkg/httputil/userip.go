package httputil

import (
	"net"
	"net/http"
	"strings"
)

// UserIP returns the IP address of the user from the request.
// If the request is coming from a proxy, it will return the IP address of the
// first proxy in the chain. If the request is coming from a direct connection,
// it will return the IP address of the user.
func UserIP(r *http.Request) (string, string, error) {
	remoteIP := r.Header.Get("X-Forwarded-For")

	if len(remoteIP) != 0 {
		ips := strings.Split(remoteIP, ",")
		ip := strings.TrimSpace(ips[0])

		if strings.Contains(ip, ":") {
			return net.SplitHostPort(ip)
		}

		return ip, "", nil
	}

	return net.SplitHostPort(r.RemoteAddr)
}
