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

	"github.com/ivanglie/iploc/pkg/log"
)

const (
	baseUrl = "https://www.ip2location.com/download" // IP2Location API Download Link
	code    = "DB11LITEIPV6"                         // IP2Location IPv4 and IPv6 Database Code
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type downloaderFunc func(token, path string) error
type unzipperFunc func() error
type splitterFunc func() error

type DB struct {
	sync.RWMutex

	downloadFunc downloaderFunc
	unzipFunc    unzipperFunc
	splitFunc    splitterFunc

	httpClient httpClient

	zip        string
	zipSize    int64
	csv        string
	CSVSize    int64
	chunks     []string
	BufferSize int64
}

func NewDB() *DB {
	db := &DB{httpClient: &http.Client{}}
	db.downloadFunc = db.download
	db.unzipFunc = db.unzip
	db.splitFunc = db.split

	return db
}

func (db *DB) Init(token, path string) error {
	log.Info("Download...")
	if err := db.downloadFunc(token, path); err != nil {
		return fmt.Errorf("downloading: %v", err)
	}
	log.Info("Download completed")

	log.Info("Unzip...")
	if err := db.unzipFunc(); err != nil {
		return fmt.Errorf("unzipping: %v", err)
	}
	log.Info("Unzip completed")

	log.Info("Split...")
	db.BufferSize = db.CSVSize / 200
	if err := db.splitFunc(); err != nil {
		return fmt.Errorf("splitting: %v", err)
	}
	log.Info("Split completed")

	return nil
}

// Search for a given IP address and return a Loc struct.
func (db *DB) Search(address string) (*Loc, error) {
	db.RLock()
	defer db.RUnlock()

	return search(address, db.chunks)
}

// split CSV file.
func (db *DB) split() (err error) {
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

// unzip file.
func (db *DB) unzip() (err error) {
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

// download IP2Location database (specified by token) to path.
func (db *DB) download(token, path string) (err error) {
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
	if resp, err = db.httpClient.Do(req); err != nil {
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
