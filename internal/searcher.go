package internal

import (
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
)

type IP struct {
	FirstIP   *big.Int `json:"-"` // First IP address show netblock.
	LastIP    *big.Int `json:"-"` // Last IP address show netblock.
	Code      string   // Two-character country code based on ISO 3166.
	Country   string   // Country name based on ISO 3166.
	Region    string   // Region or state name.
	City      string   // City name.
	Latitude  string   // City latitude. Default to capital city latitude if city is unknown.
	Longitude string   // City longitude. Default to capital city longitude if city is unknown.
	ZipCode   string   // ZIP/Postal code.
	TimeZone  string   // UTC time zone (with DST supported).
}

func newIP(firstIP, lastIP *big.Int, code, country, region, city, latitude, longitude, zipCode, timeZone string) *IP {
	return &IP{FirstIP: firstIP, LastIP: lastIP, Code: code, Country: country, Region: region, City: city,
		Latitude: latitude, Longitude: longitude, ZipCode: zipCode, TimeZone: timeZone}
}

// String representation of *IP.
func (ip *IP) String() string {
	return fmt.Sprintf("Code: %s, Country: %s, Region: %s, City: %s, Latitude: %s, Longitude: %s, ZipCode: %s, TimeZone: %s",
		ip.Code, ip.Country, ip.Region, ip.City, ip.Latitude, ip.Longitude, ip.ZipCode, ip.TimeZone)
}

// Search IP location by address.
func (db *DB) Search(address string) (ip *IP, err error) {
	if len(address) == 0 {
		err = errors.New("address is empty")
		return
	}

	if len(db.rec) == 0 {
		err = errors.New("rec is empty")
		return
	}

	num, err := convertIP(address)
	if err != nil {
		return
	}

	ip, _, err = binarySearch(db.rec, num)
	if err != nil {
		return
	}

	return
}

// binarySearch num into rec.
func binarySearch(rec [][]string, num *big.Int) (ip *IP, s []string, err error) {
	mid := len(rec) / 2
	first, _ := new(big.Int).SetString(rec[mid][0], 0)
	last, _ := new(big.Int).SetString(rec[mid][1], 0)

	log.Println("len(rec)=", len(rec), "mid=", mid)
	switch {
	case mid == 0:
		err = errors.New("not found")
		return
	case num.Cmp(first)+num.Cmp(last) > 0:
		ip, s, err = binarySearch(rec[mid:], num)
	case num.Cmp(first)+num.Cmp(last) < 0:
		ip, s, err = binarySearch(rec[:mid], num)
	default:
		s = rec[mid]
		ip = newIP(first, last, s[2], s[3], s[4], s[5], s[6], s[7], s[8], s[9])
		return
	}

	return
}

// convertIP address to num.
func convertIP(address string) (num *big.Int, err error) {
	for i := 0; i < len(address); i++ {
		if address[i] == '.' {
			// Convert IPV4 to IPV6
			// from https://stackoverflow.com/questions/27398691/ip-database-from-ip2location-and-convert-ipv4-to-ipv6-in-perl
			address = "::ffff:" + address
			break
		}
	}

	ip := net.ParseIP(address)
	if ip == nil {
		err = errors.New("'" + address + "' is incorrect IP")
		return
	}

	// from http://golang.org/pkg/net/#pkg-constants
	// IPv6len = 16
	num = big.NewInt(0).SetBytes(ip.To16())
	return
}
