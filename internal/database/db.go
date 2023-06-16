package database

import (
	"archive/zip"
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	baseUrl = "https://www.ip2location.com/download" // IP2Location API Download Link
	code    = "DB11LITEIPV6"                         // IP2Location IPv4 and IPv6 Database Code
)

type DB struct {
	sync.RWMutex
	zip        string
	zipSize    int64
	csv        string
	CSVSize    int64
	chunks     []string
	BufferSize int64
}

type CustomClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var customClient CustomClient

func init() {
	customClient = &http.Client{}
}

// Search for a given IP address and return a Loc struct.
func (db *DB) Search(address string) (*Loc, error) {
	db.RLock()
	defer db.RUnlock()

	return search(address, db.chunks)
}

// Split CSV file.
func (db *DB) Split() (err error) {
	db.Lock()
	defer db.Unlock()

	if len(db.csv) == 0 {
		err = errors.New("empty db.csv")
		return
	}

	if db.CSVSize == 0 {
		err = errors.New("db.csvSize is 0")
		return
	}

	file, err := os.Open(db.csv)
	if err != nil {
		return
	}
	defer file.Close()

	buffer := make([]byte, db.BufferSize)
	head := []byte{}
	i := 0
	for {
		count, err := file.Read(buffer)
		if err == io.EOF {
			break
		}

		chunk := append(head, buffer[:count]...)
		count = len(chunk)

		if index := bytes.LastIndex(chunk, []byte{'\n'}); index > -1 {
			chunk = chunk[:index]
			head = chunk[index+1 : count]
		}

		i++
		np, _ := filepath.Abs(fmt.Sprintf("%s_%04d.CSV", strings.TrimSuffix(db.csv, ".CSV"), i))
		os.WriteFile(np, chunk, 0777) //TODO: add error handling

		db.chunks = append(db.chunks, np)
	}

	return
}

// Unzip file.
func (db *DB) Unzip() (err error) {
	db.Lock()
	defer db.Unlock()

	if len(db.zip) == 0 {
		err = fmt.Errorf("empty db.zip")
		return
	}

	var zr *zip.ReadCloser
	if zr, err = zip.OpenReader(db.zip); err != nil {
		return
	}
	defer zr.Close()

	for _, f := range zr.File {
		if f.Name[len(f.Name)-4:] != ".CSV" {
			continue
		}

		var in io.ReadCloser
		if in, err = f.Open(); err != nil {
			continue
		}
		defer in.Close()

		db.csv = filepath.Join(filepath.Dir(db.zip), f.Name)

		var out *os.File
		if out, err = os.Create(db.csv); err != nil {
			continue
		}
		defer out.Close()

		r := bufio.NewReader(in)
		for {
			var line []byte
			if line, _, err = r.ReadLine(); err == io.EOF {
				break
			}

			if _, err = fmt.Fprintln(out, string(line)); err != nil {
				continue
			}
		}

		var info os.FileInfo
		info, err = out.Stat()
		if err != nil {
			continue
		}

		db.CSVSize = info.Size()
	}

	return
}

// Download IP2Location database (specified by token) to path.
func (db *DB) Download(token, path string) (err error) {
	db.Lock()
	defer db.Unlock()

	if len(path) == 0 {
		err = fmt.Errorf("empty path")
		return
	}

	var req *http.Request
	if req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s?token=%s&file=%s", baseUrl, token, code), nil); err != nil {
		return
	}

	var resp *http.Response
	if resp, err = customClient.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("error %d %s", resp.StatusCode, resp.Status)
		return
	}

	if db.zip, err = filepath.Abs(filepath.Join(filepath.Dir(path), code+".zip")); err != nil {
		return
	}

	var file *os.File
	if file, err = os.OpenFile(db.zip, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fs.ModeAppend); err != nil {
		return
	}
	defer file.Close()

	if _, err = io.Copy(file, resp.Body); err != nil {
		return
	}

	var info os.FileInfo
	if info, err = file.Stat(); err != nil {
		return
	}

	db.zipSize = info.Size()

	return
}

// String returns a string representation of the DB struct.
func (db *DB) String() string {
	return fmt.Sprintf("DB{zip: %s, zipSize: %d, csv: %s, csvSize: %d, chunks: %v, ChunksCount: %d}",
		db.zip, db.zipSize, db.csv, db.CSVSize, db.chunks, db.BufferSize)
}
