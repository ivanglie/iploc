package internal

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// splitCSV file specified by p on smaller chunks and return a filepaths of chunks.
func splitCSV(p string, chunks int) (s []string, err error) {
	if len(p) == 0 {
		err = errors.New("incorrect filename")
		return
	}

	var lines int
	lines, err = func(s string) (count int, err error) {
		var f *os.File
		f, err = os.Open(s)
		if err != nil {
			return
		}
		defer f.Close()

		buf := make([]byte, 32*1024)
		sep := []byte{'\n'}

		var n int
		for {
			n, err = f.Read(buf)
			count += bytes.Count(buf[:n], sep)

			switch {
			case err == io.EOF:
				err = nil
				return
			case err != nil:
				return
			}
		}
	}(p)
	if err != nil {
		log.Println(err)
		return
	}

	i := 0
	linesInChunck := lines / chunks
	chunk := []string{}
	createChunk := func(filepath string, data []string) (err error) {
		var f *os.File
		f, err = os.Create(filepath)
		if err != nil {
			return
		}

		writer := bufio.NewWriter(f)
		_, err = writer.WriteString(strings.Join(data, "\n"))
		if err != nil {
			return
		}

		f.Close()

		return
	}

	var f *os.File
	f, err = os.Open(p)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		chunk = append(chunk, scanner.Text())

		if len(chunk) == linesInChunck {
			i++
			np := fmt.Sprintf("%s_%04d.CSV", strings.TrimSuffix(p, ".CSV"), i)
			createChunk(np, chunk)

			s = append(s, np)
			chunk = []string{}
		}
	}

	if len(chunk) > 0 {
		createChunk(fmt.Sprintf("%s_%04d.CSV", strings.TrimSuffix(p, ".CSV"), i), chunk)
	}

	return
}
