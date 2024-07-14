package utils

import (
	"archive/zip"
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func SplitCSV(filePath string, fileSize, bufferSize int64) ([]string, error) {
	chunks := []string{}

	if len(filePath) == 0 {
		return chunks, errors.New("empty db.csv")
	}

	if fileSize == 0 {
		return chunks, errors.New("db.csvSize is 0")
	}

	if bufferSize == 0 {
		return chunks, errors.New("bufferSize is 0")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return chunks, err
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
		if err != nil {
			return chunks, err
		}

		chunk := append(head, buffer[:count]...)
		count = len(chunk)

		if index := bytes.LastIndex(chunk, []byte{'\n'}); index > -1 {
			head = chunk[index+1 : count]
			chunk = chunk[:index]
		} else {
			head = nil
		}

		i++
		np, _ := filepath.Abs(fmt.Sprintf("%s_%04d.CSV", strings.TrimSuffix(filePath, ".CSV"), i))
		err = os.WriteFile(np, chunk, 0777)
		if err != nil {
			return chunks, err
		}

		chunks = append(chunks, np)
	}

	return chunks, nil
}

func UnzipCSV(filePath string) (string, error) {
	csvFilePath := ""

	if len(filePath) == 0 {
		return csvFilePath, fmt.Errorf("empty filePath")
	}

	zr, err := zip.OpenReader(filePath)
	if err != nil {
		return csvFilePath, err
	}
	defer zr.Close()

	for _, f := range zr.File {
		if f.Name[len(f.Name)-4:] != ".CSV" {
			continue
		}

		var in io.ReadCloser
		if in, err = f.Open(); err != nil {
			return csvFilePath, err
		}
		defer in.Close()

		csvFilePath = filepath.Join(filepath.Dir(filePath), f.Name)

		var out *os.File
		if out, err = os.Create(csvFilePath); err != nil {
			return csvFilePath, err
		}
		defer out.Close()

		r := bufio.NewReader(in)
		for {
			var line []byte
			if line, _, err = r.ReadLine(); err == io.EOF {
				break
			}
			if err != nil {
				return csvFilePath, err
			}

			if _, err = fmt.Fprintln(out, string(line)); err != nil {
				return csvFilePath, err
			}
		}
	}

	if csvFilePath == "" {
		return csvFilePath, fmt.Errorf("no CSV file found in the zip archive")
	}

	return csvFilePath, nil
}

func CopyFile(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(d, s)
	return err
}

func FileSize(filePath string) (int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}

	return fileInfo.Size(), nil
}
