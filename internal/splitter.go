package internal

import (
	"bytes"
	"encoding/csv"
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

	linesInChunck := lines / chunks
	i := 0
	recs := [][]string{}

	var f *os.File
	f, err = os.Open(p)
	if err != nil {
		return
	}
	defer f.Close()

	reader := csv.NewReader(f)
	for {
		rec := []string{}
		rec, err = reader.Read()
		if err != nil {
			if err == io.EOF {
				err = nil
				return
			}

			log.Println("err:", err)
			return
		}

		if len(recs) < linesInChunck {
			recs = append(recs, rec)
		}

		if len(recs) == linesInChunck {
			i++
			np := fmt.Sprintf("%s_%04d.CSV", strings.TrimSuffix(p, ".CSV"), i)

			var nf *os.File
			nf, err = os.Create(np)
			if err != nil {
				log.Println("failed creating file:", err)
			}

			writer := csv.NewWriter(nf)
			err = writer.WriteAll(recs)
			if err != nil {
				log.Println("failed writing in file:", err)
				return
			}
			nf.Close()

			recs = [][]string{}
			s = append(s, np)
		}
	}
}
