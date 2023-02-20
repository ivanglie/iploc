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
func UnzipCSV(p string) (c *CSV, err error) {
	if len(p) == 0 {
		err = errors.New("incorrect path")
		return
	}

	p, err = filepath.Abs(p)
	if err != nil {
		return
	}

	zr, err := zip.OpenReader(p)
	if err != nil {
		return
	}
	defer zr.Close()

	for _, f := range zr.File {
		if f.Name[len(f.Name)-4:] != ".CSV" {
			continue
		}

		in, err := f.Open()
		if err != nil {
			log.Println(err)
			continue
		}
		defer in.Close()

		out, err := os.Create(filepath.Join(filepath.Dir(p), f.Name))
		if err != nil {
			log.Println(err)
			continue
		}
		defer out.Close()

		r := bufio.NewReader(in)
		for {
			var line []byte
			line, _, err = r.ReadLine()
			if err == io.EOF {
				break
			}

			_, err = fmt.Fprintln(out, string(line))
			if err != nil {
				log.Println(err)
				continue
			}
		}

		var info os.FileInfo
		info, err = out.Stat()
		if err != nil {
			log.Println(err)
			continue
		}

		c = &CSV{File: info.Name(), Size: info.Size()}
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
