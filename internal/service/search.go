package service

import (
	"encoding/csv"
	"errors"
	"fmt"
	"math/big"
	"os"

	"github.com/ivanglie/iploc/internal/provider"
	"github.com/ivanglie/iploc/pkg/netutil"
)

// Search for a given IP address and return a Loc struct.
func (s *Service) Search(address string) (loc *provider.Loc, err error) {
	num, err := netutil.ConvertIP(address)
	if err != nil {
		return
	}

	rec, err := searchChunk(num, s.chunks)
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
func searchByNum(num *big.Int, rec [][]string) (loc *provider.Loc, err error) {
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
		loc = provider.NewLoc(first, last, s[2], s[3], s[4], s[5], s[6], s[7], s[8], s[9])
		return
	}

	return
}
