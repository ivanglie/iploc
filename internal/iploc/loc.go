package iploc

import (
	"encoding/csv"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"strings"
)

const (
	Code      Properties = "Code"      // Two-character country code based on ISO 3166.
	Country   Properties = "Country"   // Country name based on ISO 3166.
	Region    Properties = "Region"    // Region or state name.
	City      Properties = "City"      // City name.
	Latitude  Properties = "Latitude"  // City latitude. Default to capital city latitude if city is unknown.
	Longitude Properties = "Longitude" // City longitude. Default to capital city longitude if city is unknown.
	ZipCode   Properties = "ZipCode"   // ZIP/Postal code.
	TimeZone  Properties = "TimeZone"  // UTC time zone (with DST supported).
)

type Properties string

type Loc struct {
	FirstIP    *big.Int `json:"-"` // First IP address show netblock.
	LastIP     *big.Int `json:"-"` // Last IP address show netblock.
	Properties map[Properties]string
}

func newLoc(firstIP, lastIP *big.Int, code, country, region, city, latitude, longitude, zipCode, timeZone string) *Loc {
	loc := &Loc{FirstIP: firstIP, LastIP: lastIP}

	loc.Properties = make(map[Properties]string)
	loc.Properties[Code] = code
	loc.Properties[Country] = country
	loc.Properties[Region] = region
	loc.Properties[City] = city
	loc.Properties[Latitude] = latitude
	loc.Properties[Longitude] = longitude
	loc.Properties[ZipCode] = zipCode
	loc.Properties[TimeZone] = timeZone

	return loc
}

// String representation of *IP.
func (loc *Loc) String() string {
	p := loc.Properties
	return fmt.Sprintf(
		`{"Code":"%s","Country":"%s","Region":"%s","City":"%s","Latitude":"%s","Longitude":"%s","ZipCode":"%s","TimeZone":"%s"}`,
		p[Code], p[Country], p[Region], p[City], p[Latitude], p[Longitude], p[ZipCode], p[TimeZone])
}

// Search location by address in file paths.
func Search(address string, paths []string) (loc *Loc, err error) {
	num, err := convertIP(address)
	if err != nil {
		return
	}

	rec, err := searchChunk(num, paths)
	if err != nil {
		return
	}

	loc, err = searchByNum(num, rec)
	if err != nil {
		return
	}

	return
}

// searchChunk where num is contained in file paths.
func searchChunk(num *big.Int, paths []string) (r [][]string, err error) {
	if len(paths) == 0 {
		err = errors.New("chunks is empty or not found")
		return
	}

	var f *os.File
	mid := len(paths) / 2
	f, err = os.Open(paths[mid])
	if err != nil {
		return
	}

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = 10
	rec, err := reader.ReadAll()
	if err != nil {
		return
	}

	first, _ := new(big.Int).SetString(rec[0][0], 0)
	last, _ := new(big.Int).SetString(rec[len(rec)-1][1], 0)

	switch {
	case num.Cmp(last) > 0:
		r, err = searchChunk(num, paths[mid:])
	case num.Cmp(first) < 0:
		r, err = searchChunk(num, paths[:mid])
	case num.Cmp(first)+num.Cmp(last) == 0:
		r = rec
		return
	}

	return
}

// searchByNum search location by num into rec using binary search algorithm.
func searchByNum(num *big.Int, rec [][]string) (loc *Loc, err error) {
	mid := len(rec) / 2
	first, _ := new(big.Int).SetString(rec[mid][0], 0)
	last, _ := new(big.Int).SetString(rec[mid][1], 0)

	switch {
	case mid == 0:
		err = fmt.Errorf("%v not found", num)
		return
	case num.Cmp(first)+num.Cmp(last) > 0:
		loc, err = searchByNum(num, rec[mid:])
	case num.Cmp(first)+num.Cmp(last) < 0:
		loc, err = searchByNum(num, rec[:mid])
	default:
		s := rec[mid]
		loc = newLoc(first, last, s[2], s[3], s[4], s[5], s[6], s[7], s[8], s[9])
		return
	}

	return
}

// convertIP address to num.
func convertIP(address string) (num *big.Int, err error) {
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
