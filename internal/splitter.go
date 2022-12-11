package internal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// splitCSV file specified by p on smaller chunks and return a filepaths of chunks.
func splitCSV(p string, bufferSize int64) (s []string, err error) {
	if len(p) == 0 {
		err = errors.New("incorrect filename")
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
	chunk := make([]byte, bufferSize)
	head := make([]byte, bufferSize)
	i := 0
	for {
		count, err := file.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Println("err:", err)
			}

			break
		}

		chunk = append(head, buffer[:count]...)
		count = len(chunk)

		if index := bytes.LastIndex(chunk, []byte{'\n'}); index > -1 {
			chunk = chunk[:index]
			head = chunk[index+1 : count]
		}

		i++
		np := fmt.Sprintf("%s_%04d.CSV", strings.TrimSuffix(p, ".CSV"), i)
		writeToFile(np, chunk)
		s = append(s, np)
	}

	return
}
