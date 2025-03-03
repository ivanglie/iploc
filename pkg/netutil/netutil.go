package netutil

import (
	"errors"
	"fmt"
	"math/big"
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

// ConvertIP converts an IP address to a big.Int.
// If the address is an IPv4 address, it will be converted to an IPv6 address.
// If the address is not a valid IP address, an error will be returned.
// The big.Int representation of the IP address will be returned.
func ConvertIP(address string) (num *big.Int, err error) {
	if len(address) == 0 {
		err = errors.New("empty address")
		return
	}

	if strings.Contains(address, ".") {
		address = "::ffff:" + address
	}

	ip := net.ParseIP(address)
	if ip == nil {
		err = fmt.Errorf("address %s is incorrect IP", address)
		return
	}

	// from http://golang.org/pkg/net/#pkg-constants
	// IPv6len = 16
	num = big.NewInt(0).SetBytes(ip.To16())
	return
}
