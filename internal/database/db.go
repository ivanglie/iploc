package database

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/ivanglie/iploc/internal/csv"
	"github.com/ivanglie/iploc/pkg/log"
)

const (
	baseUrl = "https://www.ip2location.com/download" // IP2Location API Download Link
	code    = "DB11LITEIPV6"                         // IP2Location IPv4 and IPv6 Database Code

	zipPath     = "test/data/"
	zipFileName = "DB.zip"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type downloaderFunc func(token, path string) error

type DB struct {
	sync.RWMutex

	downloadFunc downloaderFunc

	httpClient httpClient

	zip        string
	zipSize    int64
	csv        string
	CSVSize    int64
	chunks     []string
	BufferSize int64
}

func New() *DB {
	db := &DB{httpClient: &http.Client{}}
	db.downloadFunc = db.download

	return db
}

func (db *DB) Init(local bool, token, path string) (err error) {
	db.Lock()
	defer db.Unlock()

	if local {
		log.Info("Copy...")
		db.zip = filepath.Join(path, zipFileName)
		if err := csv.Copy(filepath.Join(zipPath, zipFileName), db.zip); err != nil {
			return fmt.Errorf("copying: %v", err)
		}

		if db.zipSize, err = csv.Size(db.zip); err != nil {
			return err
		}

		db.CSVSize = db.zipSize
		log.Info(fmt.Sprintf("Copying completed %v", db))
	} else {
		log.Info("Download...")
		if err := db.downloadFunc(token, path); err != nil {
			return fmt.Errorf("downloading: %v", err)
		}
		log.Info("Download completed")
	}

	log.Info("Unzip...")
	if len(db.zip) == 0 {
		return fmt.Errorf("empty db.zip")
	}

	if db.csv, err = csv.Unzip(db.zip); err != nil {
		return err
	}

	if db.CSVSize, err = csv.Size(db.csv); err != nil {
		return err
	}
	log.Info("Unzip completed")

	log.Info("Split...")
	k := int64(200)
	if local {
		k = 2
	}

	db.chunks, err = csv.Split(db.csv, db.CSVSize, db.CSVSize/k)
	if err != nil {
		return fmt.Errorf("splitting: %v", err)
	}

	log.Info("Split completed")

	return err
}

// Search for a given IP address and return a Loc struct.
func (db *DB) Search(address string) (*Loc, error) {
	db.RLock()
	defer db.RUnlock()

	return search(address, db.chunks)
}

// download IP2Location database (specified by token) to path.
func (db *DB) download(token, path string) (err error) {
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

	if db.zipSize, err = csv.Size(db.zip); err != nil {
		return err
	}

	return
}

// String returns a string representation of the DB struct.
func (db *DB) String() string {
	return fmt.Sprintf("DB{zip: %s, zipSize: %d, csv: %s, csvSize: %d, chunks: %v, ChunksCount: %d}",
		db.zip, db.zipSize, db.csv, db.CSVSize, db.chunks, db.BufferSize)
}
