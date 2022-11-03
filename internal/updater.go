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
		var s int64
		db.zip, s, err = download(db.code)
		if err != nil {
			log.Panic(err)
		}
		log.Printf("%s downloaded (%d bytes)", db.zip, s)

		log.Println(db.zip, "unzipping...")
		db.csv, db.size, err = unzipCSV(db.zip)
		if err != nil {
			log.Panic(err)
		}
		log.Printf("%s unzipped (%d bytes)", db.csv, db.size)

		log.Println(db.csv, "splitting...")
		db.chunks, err = splitCSV(db.csv, db.size)
		if err != nil {
			log.Panic(err)
		}
		log.Printf("%s splitted (%d chunks)", db.csv, len(db.chunks))

		log.Printf("db=%v", db)
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

// splitCSV file specified by n on smaller chunks and return a filepaths of chunks.
func splitCSV(n string, s int64) (chunks []string, err error) {
	splitter := scsv.New()
	splitter.FileChunkSize = int(s) / 200
	splitter.WithHeader = false // copying of header in chunks is disabled
	chunks, err = splitter.Split(n, "")

	return
}

// unzipCSV file specified by n and return an extracted csv filename, size.
func unzipCSV(n string) (csv string, size int64, err error) {
	wd, err := os.Getwd()
	if err != nil {
		return
	}

	var archive *zip.ReadCloser
	archive, err = zip.OpenReader(n)
	if err != nil {
		return
	}
	defer archive.Close()

	for _, f := range archive.File {
		if !strings.Contains(f.Name, ".CSV") {
			err = errors.New("not csv file")
			continue
		}

		filePath := filepath.Join(wd, f.Name)

		if !strings.HasPrefix(filePath, filepath.Clean(wd)+string(os.PathSeparator)) {
			err = errors.New("invalid file path " + filePath)
			return
		}

		if f.FileInfo().IsDir() {
			fmt.Println("creating directory...")
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return
		}

		var dstFile *os.File
		dstFile, err = os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return
		}
		defer dstFile.Close()

		var fileInArchive io.ReadCloser
		fileInArchive, err = f.Open()
		if err != nil {
			return
		}
		defer fileInArchive.Close()

		if _, err = io.Copy(dstFile, fileInArchive); err != nil {
			return
		}

		var dstFileInfo os.FileInfo
		if dstFileInfo, err = dstFile.Stat(); err != nil {
			return
		}

		csv = dstFileInfo.Name()
		size = dstFileInfo.Size()
	}

	return
}

// download IP2Location database specified by code and return a zip filename, size.
func download(code string) (z string, size int64, err error) {
	token := os.Getenv("IP2LOCATION_TOKEN")
	url := fmt.Sprintf("%s?token=%s&file=%s", baseUrl, token, code)

	out, err := os.Create(code + ".zip")
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

	if _, err = io.Copy(out, resp.Body); err != nil {
		return
	}

	var outFileInfo os.FileInfo
	if outFileInfo, err = out.Stat(); err != nil {
		return
	}

	z = outFileInfo.Name()
	size = outFileInfo.Size()

	return
}
