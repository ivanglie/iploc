package internal

import (
	"archive/zip"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	baseUrl = "https://www.ip2location.com/download" // IP2Location API Download Link
	dbCode  = "DB11LITEIPV6"                         // IP2Location IPv4 and IPv6 Database Code
)

type DB struct {
	path string
	file *os.File
	rec  [][]string
}

func NewDB() *DB {
	return &DB{}
}

// Update database: download zip file, unzip it to csv file, open and read it.
func (db *DB) Update() (err error) {
	update := func(code string) {
		var err error

		log.Println(code, "downloading...")
		err = download(code)
		if err != nil {
			log.Panic(err)
		}
		log.Println(code, "downloaded")

		log.Println(code, "unzipping...")
		db.path, err = extract(code)
		if err != nil {
			log.Panic(err)
		}
		log.Println(code, "unzipped")

		log.Println(code, "opening...")
		db.file, err = openCSV(db.path)
		if err != nil {
			log.Panic(err)
		}
		defer db.file.Close()
		log.Println(code, "opened...")

		log.Println(code, "reading...")
		reader := csv.NewReader(db.file)
		reader.FieldsPerRecord = 10
		db.rec, err = reader.ReadAll()
		if err != nil {
			log.Panic(err)
		}
		log.Println(code, "read")
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		update(dbCode)
	}()
	wg.Wait()

	return
}

// openCSV file specified by path.
func openCSV(path string) (file *os.File, err error) {
	if len(path) == 0 {
		err = errors.New("path is empty")
		return
	}

	file, err = os.Open(path)
	return
}

// extract (unzip) file specified by zipName and return an extracted fileName.
func extract(zipName string) (fileName string, err error) {
	wd, err := os.Getwd()
	if err != nil {
		log.Panic(err)
		return
	}
	srcFilePath := wd + string(os.PathSeparator) + zipName + ".zip"
	r, err := zip.OpenReader(srcFilePath)
	if err != nil {
		return
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(wd, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(dir string, zf *zip.File) (string, error) {
		rc, err := zf.Open()
		if err != nil {
			return "", err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dir, zf.Name)

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(path, filepath.Clean(dir)+string(os.PathSeparator)) {
			return "", fmt.Errorf("illegal file path: %s", path)
		}

		if zf.FileInfo().IsDir() {
			os.MkdirAll(path, zf.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), zf.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zf.Mode())
			if err != nil {
				return "", err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return "", err
			}
		}

		return zf.FileInfo().Name(), nil
	}

	for _, f := range r.File {
		name, err := extractAndWriteFile(wd, f)
		if err != nil {
			return "", err
		}
		if strings.Contains(name, ".CSV") {
			fileName = name
		}
	}

	return fileName, nil
}

// download IP2Location database specified by code.
func download(code string) (err error) {
	token := os.Getenv("IP2LOCATION_TOKEN")
	url := fmt.Sprintf("%s?token=%s&file=%s", baseUrl, token, code)

	wd, err := os.Getwd()
	if err != nil {
		log.Panic(err)
		return
	}

	destFilePath := wd + string(os.PathSeparator) + code + ".zip"
	err = func(url, destFilePath string) error {
		// Create the file
		out, err := os.Create(destFilePath)
		if err != nil {
			return err
		}
		defer out.Close()

		// Get the data
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Check server response
		if resp.StatusCode != http.StatusOK {
			return errors.New("bad status: " + resp.Status)
		}

		// Writer the body to file
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return err
		}

		return nil
	}(url, destFilePath)
	return
}
