package internal

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/csv"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"strings"
)

type IP struct {
	FirstIP   *big.Int // First IP address show netblock.
	LastIP    *big.Int // Last IP address show netblock.
	Code      string   // Two-character country code based on ISO 3166.
	Country   string   // Country name based on ISO 3166.
	Region    string   // Region or state name.
	City      string   // City name.
	Latitude  string   // City latitude. Default to capital city latitude if city is unknown.
	Longitude string   // City longitude. Default to capital city longitude if city is unknown.
	ZipCode   string   // ZIP/Postal code.
	TimeZone  string   // UTC time zone (with DST supported).
}

// String representation of *IP
func (ip *IP) String() string {
	return fmt.Sprintf("Code: %s, Country: %s, Region: %s, City: %s, Latitude: %s, Longitude: %s, ZipCode: %s, TimeZone: %s",
		ip.Code, ip.Country, ip.Region, ip.City, ip.Latitude, ip.Longitude, ip.ZipCode, ip.TimeZone)
}

func Search(address string, paths ...string) (ip *IP, err error) {
	if len(paths) == 0 {
		err = errors.New("empty path")
		return
	}

	if len(address) == 0 {
		err = errors.New("empty address")
		return
	}

	num, err := convIP(address)
	if err != nil {
		return
	}

	ipCh := make(chan *IP)
	errCh := make(chan error, len(paths))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, p := range paths {
		go search(ctx, p, num, ipCh, errCh)
	}

	numEOF := 0

loop:
	for {
		select {
		case ip = <-ipCh:
			cancel()
			break loop
		case <-errCh:
			numEOF++
			if numEOF == len(paths) {
				err = errors.New("not found")
				cancel()
				break loop
			}
		}
	}

	return
}

// search
func search(ctx context.Context, path string, num *big.Int, ipCh chan<- *IP, errCh chan<- error) {
	file, err := openCSV(path)
	if err != nil {
		return
	}

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = 10

	select {
	case <-ctx.Done():
		return
	default:
		rec, err := reader.ReadAll()
		if err != nil {
			errCh <- err
			return
		}

		s, err := binarySearch(rec, num)
		if err != nil {
			errCh <- err
			return
		}

		f, b := new(big.Int).SetString(s[0][0], 0)
		if !b {
			err = errors.New("not big.Int")
			errCh <- err
			return
		}

		l, b := new(big.Int).SetString(s[0][1], 0)
		if !b {
			err = errors.New("not big.Int")
			errCh <- err
			return
		}

		ip := &IP{
			FirstIP:   f,
			LastIP:    l,
			Code:      s[0][2],
			Country:   s[0][3],
			Region:    s[0][4],
			City:      s[0][5],
			Latitude:  s[0][6],
			Longitude: s[0][7],
			ZipCode:   s[0][8],
			TimeZone:  s[0][9],
		}
		ipCh <- ip
		return
	}
}

// binarySearch
func binarySearch(rec [][]string, num *big.Int) (s [][]string, err error) {
	mid := len(rec) / 2
	first, _ := new(big.Int).SetString(rec[mid][0], 0)
	last, _ := new(big.Int).SetString(rec[mid][1], 0)

	switch {
	case mid == 0:
		err = errors.New("not found")
		return
	case num.Cmp(first)+num.Cmp(last) > 0:
		s, err = binarySearch(rec[mid:], num)
	case num.Cmp(first)+num.Cmp(last) < 0:
		s, err = binarySearch(rec[:mid], num)
	default: // case bi.Cmp(first)+bi.Cmp(last) == 0:
		s = rec[mid:]
	}
	return
}

// Convert IP address to num (*big.Int)
func convIP(address string) (num *big.Int, err error) {
	ip := net.ParseIP(address)
	if ip == nil {
		err = errors.New("'" + address + "' is incorrect IP")
		return
	}

	switch ver := strings.Count(address, ":"); {
	case ver < 2 && ip.To4() != nil:
		num, err = func(netIP net.IP) (num *big.Int, err error) {
			var long uint32
			err = binary.Read(bytes.NewBuffer(netIP), binary.BigEndian, &long)
			num = big.NewInt(0).SetBytes(netIP)
			return
		}(ip.To4())
		return
	case ver >= 2 && ip.To16() != nil:
		num, err = func(netIP net.IP) (num *big.Int, err error) {
			// from http://golang.org/pkg/net/#pkg-constants
			// IPv6len = 16
			num = big.NewInt(0).SetBytes(netIP)
			return
		}(ip.To16())
		return
	default:
		err = errors.New("'" + address + "' is incorrect IP")
		return
	}
}

// Open csv file specified by path
func openCSV(path string) (file *os.File, err error) {
	if len(path) == 0 {
		err = errors.New("path is empty")
		return
	}

	file, err = os.Open(path)
	return
}
