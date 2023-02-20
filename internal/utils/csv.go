package utils

import (
	"archive/zip"
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

// String representation of *IP.
func (csv *CSV) String() string {
	return fmt.Sprintf(
		`{"CSV":{"File":"%s","Size":%d}}`, csv.File, csv.Size)
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

// UnzipCSV file specified by p and return an extracted csv filename, size.
func UnzipCSV(p string) (csv *CSV, err error) {
	if len(p) == 0 {
		err = errors.New("incorrect path")
		return
	}

	p, err = filepath.Abs(p)
	if err != nil {
		return
	}

	log.Println("p=", p)

	var r *zip.ReadCloser
	r, err = zip.OpenReader(p)
	if err != nil {
		return
	}
	defer r.Close()

	log.Println("r=", r)

	for _, f := range r.File {
		log.Println("f=", f)
		log.Println("f.Name=", f.Name)
		if !strings.Contains(f.Name, ".CSV") {
			continue
		}

		csv = &CSV{}
		d := filepath.Dir(p)
		log.Println("d=", d)
		csv.File = filepath.Join(d, f.Name)
		if !strings.HasPrefix(csv.File, filepath.Clean(d)) {
			err = fmt.Errorf("invalid path: %s", csv.File)
			return
		}

		if f.FileInfo().IsDir() {
			continue
		}

		var out *os.File
		if out, err = os.OpenFile(csv.File, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode()); err != nil {
			return
		}
		defer out.Close()

		log.Println("out=", out)

		var r io.ReadCloser
		if r, err = f.Open(); err != nil {
			return
		}
		defer r.Close()

		log.Println("reader=", r)

		var w int64
		w, err = io.Copy(out, r)
		log.Println("w=", w, "err=", err)
		if err != nil {
			log.Println("err=", err)
			return
		}

		var info os.FileInfo
		info, err = out.Stat()
		log.Println("info=", info, "err=", err)
		if err != nil {
			return
		}

		csv.Size = info.Size()
	}

	return
}

// Download IP2Location database specified by code and return a zip filename, size.
func Download(p, token string) (d string, size int64, err error) {
	if len(p) == 0 {
		err = errors.New("incorrect path")
		return
	}

	url := fmt.Sprintf("%s?token=%s&file=%s", baseUrl, token, "DB11LITEIPV6")
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = errors.New("bad status: " + resp.Status)
		return
	}

	p, err = filepath.Abs(filepath.Dir(p))
	if err != nil {
		return
	}

	d = filepath.Join(p, code+".zip")
	out, err := os.OpenFile(d, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fs.ModeAppend)
	if err != nil {
		return
	}
	defer out.Close()

	// if _, err = io.Copy(out, io.NopCloser(bytes.NewReader([]byte("foo")))); err != nil {
	if _, err = io.Copy(out, resp.Body); err != nil {
		return
	}

	var info os.FileInfo
	if info, err = out.Stat(); err != nil {
		return
	}

	size = info.Size()

	return
}
