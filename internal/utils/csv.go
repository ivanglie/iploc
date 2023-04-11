package utils

import (
	"archive/zip"
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	baseUrl = "https://www.ip2location.com/download" // IP2Location API Download Link
	code    = "DB11LITEIPV6"                         // IP2Location IPv4 and IPv6 Database Code
)

type CSV struct {
	File string
	Size int64
}

type CustomClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var customClient CustomClient

func init() {
	customClient = &http.Client{}
}

// String representation of *IP.
func (csv *CSV) String() string {
	return fmt.Sprintf(`{"File":"%s","Size":%d}`, csv.File, csv.Size)
}

// SplitCSV file specified by p on smaller chunks and return a filepaths of chunks.
func SplitCSV(p string, bufferSize int64) (s []string, err error) {
	if len(p) == 0 {
		err = errors.New("incorrect path")
		return
	}

	file, err := os.Open(p)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	writeToFile := func(np string, b []byte) {
		err := os.WriteFile(np, b, 0777)
		if err != nil {
			log.Println("failed writing in file:", err)
			return
		}
	}

	buffer := make([]byte, bufferSize)
	head := []byte{}
	i := 0
	for {
		count, err := file.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Println("err=", err)
			}

			break
		}

		chunk := append(head, buffer[:count]...)
		count = len(chunk)

		if index := bytes.LastIndex(chunk, []byte{'\n'}); index > -1 {
			chunk = chunk[:index]
			head = chunk[index+1 : count]
		}

		i++
		np := fmt.Sprintf("%s_%04d.CSV", strings.TrimSuffix(p, ".CSV"), i)
		np, _ = filepath.Abs(np)
		writeToFile(np, chunk)
		s = append(s, np)
	}

	return
}

// Unzip file specified by p and return an extracted csv filename, size.
func Unzip(path string) (c *CSV, err error) {
	if len(path) == 0 {
		err = fmt.Errorf("incorrect path %s", path)
		return
	}

	if path, err = filepath.Abs(path); err != nil {
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

		name := filepath.Join(filepath.Dir(path), f.Name)

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

		c = &CSV{File: name, Size: info.Size()}
	}

	return
}

// Download IP2Location database specified by token and return a name, size of zip file.
func Download(token, path string) (name string, size int64, err error) {
	if len(path) == 0 {
		err = fmt.Errorf("incorrect path %s", path)
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
		err = fmt.Errorf("bad status: %s", resp.Status)
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
