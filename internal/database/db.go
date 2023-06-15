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
	Zip     string
	ZipSize int64
	Csv     string
	CsvSize int64
	Chunks  []string
}

type CustomClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var customClient CustomClient

func init() {
	customClient = &http.Client{}
}

// Prepare are downloading, unzipping and splitting the database file.
// token is a token for downloading the database file.
// path is a path to the database directory.
// chunks is a number of chunks to split the database file.
func (db *DB) Prepare(token, path string, chunks int64) (err error) {
	db.Lock()
	defer db.Unlock()

	if db.Zip, db.ZipSize, err = download(token, path); err != nil {
		return
	}

	if db.Csv, db.CsvSize, err = unzip(db.Zip); err != nil {
		return
	}

	if db.Chunks, err = split(db.Csv, db.CsvSize/chunks); err != nil {
		return
	}

	return
}

// Search for a given IP address and return a Loc struct.
func (db *DB) Search(address string) (*Loc, error) {
	db.RLock()
	defer db.RUnlock()

	return search(address, db.Chunks)
}

// split CSV file specified by p on smaller chunks and return a filepaths of chunks.
func split(p string, bufferSize int64) (s []string, err error) {
	if len(p) == 0 {
		err = errors.New("incorrect path")
		return
	}

	file, err := os.Open(p)
	if err != nil {
		return
	}
	defer file.Close()

	buffer := make([]byte, bufferSize)
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
		np, _ := filepath.Abs(fmt.Sprintf("%s_%04d.CSV", strings.TrimSuffix(p, ".CSV"), i))
		os.WriteFile(np, chunk, 0777) //TODO: add error handling

		s = append(s, np)
	}

	return
}

// unzip file specified by p and return an extracted csv filename, size.
func unzip(path string) (name string, size int64, err error) {
	if len(path) == 0 {
		err = fmt.Errorf("empty path")
		return
	}

	var zr *zip.ReadCloser
	if zr, err = zip.OpenReader(path); err != nil {
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

		name = filepath.Join(filepath.Dir(path), f.Name)

		var out *os.File
		if out, err = os.Create(name); err != nil {
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

		size = info.Size()
	}

	return
}

// download IP2Location database specified by token and return a name, size of zip file.
func download(token, path string) (name string, size int64, err error) {
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
		err = fmt.Errorf("bad status: code %d, error %s", resp.StatusCode, resp.Status)
		return
	}

	if name, err = filepath.Abs(filepath.Join(filepath.Dir(path), code+".zip")); err != nil {
		return
	}

	var file *os.File
	if file, err = os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fs.ModeAppend); err != nil {
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

	size = info.Size()

	return
}
