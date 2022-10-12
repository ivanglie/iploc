package internal

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"os"
	"strings"
)

// Item
type Item struct {
	IPfrom    *big.Int // First IP address show netblock.
	IPto      *big.Int // Last IP address show netblock.
	Code      string   // Two-character country code based on ISO 3166.
	Country   string   // Country name based on ISO 3166.
	Region    string   // Region or state name.
	City      string   // City name.
	Latitude  string   // City latitude. Default to capital city latitude if city is unknown.
	Longitude string   // City longitude. Default to capital city longitude if city is unknown.
	ZipCode   string   // ZIP/Postal code.
	TimeZone  string   // UTC time zone (with DST supported).
}

// String representation of *Item
func (i *Item) String() string {
	return fmt.Sprintf("Code: %s, Country: %s, Region: %s, City: %s, Latitude: %s, Longitude: %s, ZipCode: %s, TimeZone: %s",
		i.Code, i.Country, i.Region, i.City, i.Latitude, i.Longitude, i.ZipCode, i.TimeZone)
}

// Find
func Find(s string, paths ...string) (item *Item, err error) {
	if len(paths) == 0 {
		err = errors.New("empty path")
		return
	}

	if len(s) == 0 {
		err = errors.New("empty s")
		return
	}

	ip, err := convIP(s)
	log.Println(ip)
	if err != nil {
		return
	}

	itemCh := make(chan *Item)
	errCh := make(chan error, len(paths))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, p := range paths {
		log.Println(p)
		go find(ctx, p, ip, itemCh, errCh)
	}

	numEOF := 0

loop:
	for {
		select {
		case item = <-itemCh:
			cancel()
			break loop
		case <-errCh:
			log.Println(err, numEOF, len(paths))
			numEOF++
			log.Println("numEOF:", numEOF)
			if numEOF == len(paths) {
				err = errors.New("not found")
				cancel()
				break loop
			}
		}
	}

	return
}

// Find
func find(ctx context.Context, path string, ip *big.Int, itemCh chan<- *Item, errCh chan<- error) {
	file, err := openCSV(path)
	if err != nil {
		return
	}

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = 10

	n := 0
	for {
		n++
		select {
		case <-ctx.Done():
			return
		default:
			rec, err := reader.Read()
			if err != nil {
				log.Println("error:", err, "rec:", rec, "line:", n, "file:", file.Name())

				if err == io.EOF {
					errCh <- err
					return
				}

				log.Println("error occurs when reading: ", err)
				continue
			}

			fromIP, b := new(big.Int).SetString(rec[0], 0)
			if !b {
				log.Printf("field fromIP is not a int (%v)", err)
				continue
			}
			toIP, b := new(big.Int).SetString(rec[1], 0)
			if !b {
				log.Printf("field toIP is not a int (%v)", err)
				continue
			}

			it := &Item{fromIP, toIP, rec[2], rec[3], rec[4], rec[5], rec[6], rec[7], rec[8], rec[9]}
			if ip.Cmp(it.IPfrom)+ip.Cmp(it.IPto) == 0 {
				itemCh <- it
				log.Println("success! line:", n, "file:", file.Name())
				return
			}
		}
	}
}

// Convert IP address to num (*big.Int)
func convIP(address string) (num *big.Int, err error) {
	ip := net.ParseIP(address)
	if ip == nil {
		err = errors.New(address + " is incorrect IP")
		return
	}

	switch ver := strings.Count(address, ":"); {
	case ver < 2 && ip.To4() != nil:
		num, err = convIPv4(ip.To4())
		return
	case ver >= 2 && ip.To16() != nil:
		num, err = convIPv6(ip.To16())
		return
	default:
		err = errors.New(address + " is incorrect IP")
		return
	}
}

// Convert netIP (v6) address to num (*big.Int)
func convIPv4(netIP net.IP) (num *big.Int, err error) {
	var long uint32
	err = binary.Read(bytes.NewBuffer(netIP), binary.BigEndian, &long)
	num = big.NewInt(0).SetBytes(netIP)
	return
}

// Convert netIP (v6) address to num (*big.Int)
func convIPv6(netIP net.IP) (num *big.Int, err error) {
	// from http://golang.org/pkg/net/#pkg-constants
	// IPv6len = 16
	num = big.NewInt(0).SetBytes(netIP)
	return
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
