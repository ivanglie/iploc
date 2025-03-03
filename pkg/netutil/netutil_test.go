package netutil

import (
	"math/big"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserIP(t *testing.T) {
	tests := []struct {
		name          string
		remoteAddr    string
		xForwardedFor string
		expectedIP    string
		expectedPort  string
		expectedError bool
	}{
		{
			name:          "Direct connection with port",
			remoteAddr:    "192.168.1.1:8080",
			xForwardedFor: "",
			expectedIP:    "192.168.1.1",
			expectedPort:  "8080",
			expectedError: false,
		},
		{
			name:          "Direct connection with IPv6 and port",
			remoteAddr:    "[2001:db8::1]:8080",
			xForwardedFor: "",
			expectedIP:    "2001:db8::1",
			expectedPort:  "8080",
			expectedError: false,
		},
		{
			name:          "Through proxy with single IP",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.195",
			expectedIP:    "203.0.113.195",
			expectedPort:  "",
			expectedError: false,
		},
		{
			name:          "Through proxy with multiple IPs",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.195, 198.51.100.17, 192.0.2.60",
			expectedIP:    "203.0.113.195",
			expectedPort:  "",
			expectedError: false,
		},
		{
			name:          "Through proxy with IPv6",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "2001:db8::1",
			expectedIP:    "2001:db8::1",
			expectedPort:  "",
			expectedError: true, // SplitHostPort fails with IPv6 without brackets
		},
		{
			name:          "Through proxy with IPv6 and port",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "[2001:db8::1]:8080",
			expectedIP:    "2001:db8::1",
			expectedPort:  "8080",
			expectedError: false,
		},
		{
			name:          "Invalid remote address",
			remoteAddr:    "192.168.1.1", // Missing port
			xForwardedFor: "",
			expectedIP:    "",
			expectedPort:  "",
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock request
			req := &http.Request{
				RemoteAddr: tc.remoteAddr,
				Header:     make(http.Header),
			}

			if tc.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tc.xForwardedFor)
			}

			// Call the function
			ip, port, err := UserIP(req)

			// Check results
			if tc.expectedError {
				assert.Error(t, err, "Expected an error")
			} else {
				assert.NoError(t, err, "Did not expect an error")
				assert.Equal(t, tc.expectedIP, ip, "IP addresses should match")
				assert.Equal(t, tc.expectedPort, port, "Ports should match")
			}
		})
	}
}

func TestConvertIP(t *testing.T) {
	tests := []struct {
		name          string
		address       string
		expectedInt   *big.Int
		expectedError bool
	}{
		{
			name:          "Valid IPv4 address",
			address:       "192.168.1.1",
			expectedInt:   big.NewInt(0).SetBytes(net.ParseIP("::ffff:192.168.1.1").To16()),
			expectedError: false,
		},
		{
			name:          "Valid IPv6 address",
			address:       "2001:db8::1",
			expectedInt:   big.NewInt(0).SetBytes(net.ParseIP("2001:db8::1").To16()),
			expectedError: false,
		},
		{
			name:          "Empty address",
			address:       "",
			expectedInt:   nil,
			expectedError: true,
		},
		{
			name:          "Invalid IP address",
			address:       "not.an.ip.address",
			expectedInt:   nil,
			expectedError: true,
		},
		{
			name:          "IPv4 address with invalid format",
			address:       "192.168.1.256", // Invalid octet
			expectedInt:   nil,
			expectedError: true,
		},
		{
			name:          "IPv6 loopback",
			address:       "::1",
			expectedInt:   big.NewInt(1), // The numeric value of ::1 is 1
			expectedError: false,
		},
		{
			name:          "IPv4 loopback as IPv6",
			address:       "127.0.0.1",
			expectedInt:   big.NewInt(0).SetBytes(net.ParseIP("::ffff:127.0.0.1").To16()),
			expectedError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Call the function
			result, err := ConvertIP(tc.address)

			// Check results
			if tc.expectedError {
				assert.Error(t, err, "Expected an error")
				assert.Nil(t, result, "Result should be nil when an error occurs")
			} else {
				assert.NoError(t, err, "Did not expect an error")
				require.NotNil(t, result, "Result should not be nil")
				assert.Equal(t, 0, tc.expectedInt.Cmp(result),
					"Expected %s but got %s", tc.expectedInt.String(), result.String())
			}
		})
	}
}

func TestConvertIP_SpecificValues(t *testing.T) {
	// Test specific numeric values
	testCases := []struct {
		ipAddress string
		expected  string // Expected numeric value as a string
	}{
		{"0.0.0.0", "281470681743360"},         // ::ffff:0.0.0.0
		{"127.0.0.1", "281472812449793"},       // ::ffff:127.0.0.1
		{"255.255.255.255", "281474976710655"}, // ::ffff:255.255.255.255
		{"::1", "1"},
		{"192.0.2.128", "281473902969472"}, // Converted to ::ffff:192.0.2.128
		{"2001:db8::1", "42540766411282592856903984951653826561"},
	}

	for _, tc := range testCases {
		t.Run(tc.ipAddress, func(t *testing.T) {
			result, err := ConvertIP(tc.ipAddress)

			assert.NoError(t, err)
			require.NotNil(t, result)

			expected := new(big.Int)
			expected.SetString(tc.expected, 10)

			assert.Equal(t, 0, expected.Cmp(result),
				"IP %s should convert to %s, got %s", tc.ipAddress, expected.String(), result.String())
		})
	}
}

func TestUserIP_EdgeCases(t *testing.T) {
	// Test edge cases for the UserIP function
	testCases := []struct {
		name          string
		setup         func() *http.Request
		expectedIP    string
		expectedPort  string
		expectedError bool
	}{
		{
			name: "Empty X-Forwarded-For but with space",
			setup: func() *http.Request {
				req := &http.Request{
					RemoteAddr: "192.168.1.1:8080",
					Header:     make(http.Header),
				}
				req.Header.Set("X-Forwarded-For", " ")
				return req
			},
			expectedIP:    "",
			expectedPort:  "",
			expectedError: false,
		},
		{
			name: "X-Forwarded-For with empty first IP",
			setup: func() *http.Request {
				req := &http.Request{
					RemoteAddr: "192.168.1.1:8080",
					Header:     make(http.Header),
				}
				req.Header.Set("X-Forwarded-For", ", 198.51.100.17")
				return req
			},
			expectedIP:    "",
			expectedPort:  "",
			expectedError: false,
		},
		{
			name: "X-Forwarded-For with IPv6 and no brackets but with port",
			setup: func() *http.Request {
				req := &http.Request{
					RemoteAddr: "192.168.1.1:8080",
					Header:     make(http.Header),
				}
				req.Header.Set("X-Forwarded-For", "2001:db8::1:8080") // This is ambiguous
				return req
			},
			expectedIP:    "",
			expectedPort:  "",
			expectedError: true, // SplitHostPort will fail
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.setup()
			ip, port, err := UserIP(req)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedIP, ip)
				assert.Equal(t, tc.expectedPort, port)
			}
		})
	}
}
