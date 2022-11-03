package internal

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	scsv "github.com/tolik505/split-csv"
)

const (
	baseUrl = "https://www.ip2location.com/download" // IP2Location API Download Link
	dbCode  = "DB11LITEIPV6"                         // IP2Location IPv4 and IPv6 Database Code
)

type DB struct {
	code   string // database code.
	zip    string // zip file name.
	csv    string // csv file name.
	size   int64  // csv file size.
	chunks []string
	rec    [][]string
}

func NewDB() *DB {
	return &DB{code: dbCode}
}

// String representation of *IP.
func (db *DB) String() string {
	return fmt.Sprintf("code: %s, zip: %s, csv: %s, size: %d (in bytes), chunks: %d, recs: %d\n",
		db.code, db.zip, db.csv, db.size, len(db.chunks), len(db.rec))
}

// Update database: download zip file, unzip it to csv file, open and read it.
func (db *DB) Update() (err error) {
	update := func() {
		var err error

		log.Println(db.code, "downloading...")
		// db.zip = db.code + ".zip" // debug
		db.zip, err = download(db.code)
		if err != nil {
			log.Panic(err)
		}
		log.Println(db.code, "downloaded")

		log.Println(db.zip, "unzipping...")
		db.csv, db.size, err = extract(db.zip)
		if err != nil {
			log.Panic(err)
		}
		log.Println(db.zip, "unzipped")

		log.Println(db.csv, "splitting...")
		db.chunks, err = split(db.csv, db.size)
		if err != nil {
			log.Panic(err)
		}
		log.Printf("%s splitted (%d chunks)", db.csv, len(db.chunks))

		log.Println("db=", db)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		update()
	}()
	wg.Wait()

	return
}

// open csv file specified by path.
func open(path string) (file *os.File, err error) {
	if len(path) == 0 {
		err = errors.New("path is empty")
		return
	}

	file, err = os.Open(path)
	return
}

// Split csv file specified by n on smaller chunks and set a filepaths of chunks into db.chunks.
func split(n string, s int64) (chunks []string, err error) {
	splitter := scsv.New()
	splitter.FileChunkSize = int(s) / 200
	splitter.WithHeader = false // copying of header in chunks is disabled
	chunks, err = splitter.Split(n, "")

	return
}

// extract (unzip) file specified by n and return an extracted csv, size.
func extract(n string) (csv string, size int64, err error) {
	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(d string, z *zip.File) (s string, err error) {
		var rc io.ReadCloser
		rc, err = z.Open()
		if err != nil {
			return
		}
		defer rc.Close()

		p := filepath.Join(d, z.Name)

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(p, filepath.Clean(d)+string(os.PathSeparator)) {
			err = errors.New("illegal file path: " + p)
			return
		}

		if z.FileInfo().IsDir() {
			err = os.MkdirAll(p, z.Mode())
			if err != nil {
				return
			}
		} else {
			err = os.MkdirAll(filepath.Dir(p), z.Mode())
			if err != nil {
				return
			}

			var f *os.File
			f, err = os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, z.Mode())
			if err != nil {
				return
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return
			}
		}

		s = z.FileInfo().Name()
		return
	}

	wd, err := os.Getwd()
	if err != nil {
		return
	}

	r, err := zip.OpenReader(filepath.Join(wd, n))
	if err != nil {
		return
	}
	defer r.Close()

	err = os.MkdirAll(wd, 0755)
	if err != nil {
		return
	}

	for _, f := range r.File {
		var s string
		s, err = extractAndWriteFile(wd, f)
		if err != nil {
			return
		}

		if strings.Contains(s, ".CSV") {
			csv = s
			size = f.FileInfo().Size()
			return
		}
	}

	return
}

// download IP2Location database specified by code.
func download(code string) (s string, err error) {
	z := code + ".zip"
	token := os.Getenv("IP2LOCATION_TOKEN")
	url := fmt.Sprintf("%s?token=%s&file=%s", baseUrl, token, code)

	out, err := os.Create(z)
	if err != nil {
		return
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = errors.New("bad status: " + resp.Status)
		return
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return
	}

	s = z

	return
}
