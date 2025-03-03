package service

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ivanglie/iploc/internal/provider"
	"github.com/ivanglie/iploc/pkg/netutil"
)

const (
	DefaultBufferDivisor int64 = 200
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Service struct.
type Service struct {
	httpClient httpClient

	zip           string
	zipSize       int64
	csv           string
	csvSize       int64
	bufferDivisor int64
	bufferSize    int64
	chunks        []string
}

// New creates a new Service.
func New() *Service {
	return &Service{
		bufferDivisor: DefaultBufferDivisor,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// Search for a given IP address and return a Loc struct.
func (s *Service) Search(address string) (loc *provider.Location, err error) {
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
func searchByNum(num *big.Int, rec [][]string) (loc *provider.Location, err error) {
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
		loc = provider.New(first, last, s[2], s[3], s[4], s[5], s[6], s[7], s[8], s[9])
		return
	}

	return
}

// Download IP2Location database (specified by token) to path.
func (s *Service) Download(token, path string) (err error) {
	if len(path) == 0 {
		err = fmt.Errorf("empty path")
		return
	}

	var req *http.Request
	if req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s?token=%s&file=%s", provider.DefaultURL, token, provider.DefaultCode), nil); err != nil {
		return
	}

	var resp *http.Response
	if resp, err = s.httpClient.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("error %d %s", resp.StatusCode, resp.Status)
		return
	}

	if s.zip, err = filepath.Abs(filepath.Join(filepath.Dir(path), provider.DefaultCode+".zip")); err != nil {
		return
	}

	var file *os.File
	if file, err = os.OpenFile(s.zip, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fs.ModeAppend); err != nil {
		return
	}
	defer file.Close()

	if _, err = io.Copy(file, resp.Body); err != nil {
		return
	}

	if s.zipSize, err = s.size(s.zip); err != nil {
		return
	}

	return
}

// Split splits the CSV file into chunks.
func (s *Service) Split() error {
	if len(s.csv) == 0 {
		return fmt.Errorf("empty %s.csv", provider.DefaultCode)
	}

	if s.csvSize == 0 {
		return errors.New("csvSize is 0")
	}

	s.bufferSize = s.csvSize / s.bufferDivisor
	if s.bufferSize == 0 {
		return errors.New("bufferSize is 0")
	}

	file, err := os.Open(s.csv)
	if err != nil {
		return err
	}
	defer file.Close()

	buffer := make([]byte, s.bufferSize)
	head := []byte{}
	i := 0
	for {
		count, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		chunk := append(head, buffer[:count]...)
		count = len(chunk)

		if index := bytes.LastIndex(chunk, []byte{'\n'}); index > -1 {
			head = chunk[index+1 : count]
			chunk = chunk[:index]
		} else {
			head = nil
		}

		i++
		np, _ := filepath.Abs(fmt.Sprintf("%s_%04d.CSV", strings.TrimSuffix(s.csv, ".CSV"), i))
		err = os.WriteFile(np, chunk, 0777)
		if err != nil {
			return err
		}

		s.chunks = append(s.chunks, np)
	}

	return nil
}

// Unzip extracts the CSV file from the zip archive.
func (s *Service) Unzip() error {
	csvFilePath := ""

	if len(s.ZIP()) == 0 {
		return fmt.Errorf("empty zip path")
	}

	zr, err := zip.OpenReader(s.ZIP())
	if err != nil {
		return err
	}
	defer zr.Close()

	for _, f := range zr.File {
		if f.Name[len(f.Name)-4:] != ".CSV" {
			continue
		}

		var in io.ReadCloser
		if in, err = f.Open(); err != nil {
			return err
		}
		defer in.Close()

		csvFilePath = filepath.Join(filepath.Dir(s.ZIP()), f.Name)

		var out *os.File
		if out, err = os.Create(csvFilePath); err != nil {
			return err
		}
		defer out.Close()

		r := bufio.NewReader(in)
		for {
			var line []byte
			if line, _, err = r.ReadLine(); err == io.EOF {
				break
			}
			if err != nil {
				return err
			}

			if _, err = fmt.Fprintln(out, string(line)); err != nil {
				return err
			}
		}
	}

	if len(csvFilePath) == 0 {
		return fmt.Errorf("no CSV file found in the zip archive")
	}

	s.csv = csvFilePath

	if s.csvSize, err = s.size(s.csv); err != nil {
		return err
	}

	return nil
}

// Copy copies the file from src to dst.
func Copy(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(d, s)
	return err
}

// SetZIP sets the zip path.
func (s *Service) SetZIP(zip string) {
	s.zip = zip
}

// ZIP returns the zip path.
func (s *Service) ZIP() string {
	return s.zip
}

// SetBufferDivisor sets the buffer divisor.
func (s *Service) SetBufferDivisor(divisor int64) {
	s.bufferDivisor = divisor
}

// String returns a string representation of the DB struct.
func (s *Service) String() string {
	return fmt.Sprintf("{zip: %s, zipSize: %d, csv: %s, csvSize: %d, chunks: %v, count: %d}",
		s.zip, s.zipSize, s.csv, s.csvSize, s.chunks, s.bufferSize)
}

// size returns the size of the file in bytes.
func (s Service) size(filePath string) (int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}

	return fileInfo.Size(), nil
}
