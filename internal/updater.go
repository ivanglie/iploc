package internal

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	scsv "github.com/tolik505/split-csv"
)

const (
	baseUrl  = "https://www.ip2location.com/download" // IP2Location API Download Link
	codeIPv4 = "DB11LITE"                             // IP2Location IPv4 Database Code
	codeIPv6 = "DB11LITEIPV6"                         // IP2Location IPv6 Database Code
)

// Keep a state
type State struct {
	// TODO: mutex
	IPv4Chunks []string `json:"IPv4"` // csv file paths of IPv4 (chunks)
	IPv6Chunks []string `json:"IPv6"` // csv file paths of IPv6 (chunks)
	wd         string   // Working directory path
}

// Create a new state
func NewState() (s *State) {
	wd, err := os.Getwd()
	if err != nil {
		log.Panic(err)
		return
	}

	s = &State{wd: wd}
	return
}

// Write *State to JSON file specified by name
func WriteState(s *State, name string) (err error) {
	f, err := json.MarshalIndent(s, "", " ")
	err = ioutil.WriteFile(name, f, 0644)

	return
}

// Read state from JSON file specified by name
func ReadState(name string) (s *State, err error) {
	f, err := os.Open(name)
	if err != nil {
		return
	}
	defer f.Close()

	bytes, err := ioutil.ReadAll(f)
	err = json.Unmarshal(bytes, &s)
	return
}

// Update data: download zip file, unzip it to csv file, and split large csv file on smaller chunks.
func (s *State) Update() {
	update := func(state *State, code string) {
		log.Println(code, "downloading...")
		err := s.download(code)
		if err != nil {
			log.Panic(err)
		}
		log.Println(code, "was downloaded.")

		log.Println(code, "unzipping...")
		csv, err := s.extract(code)
		if err != nil {
			log.Panic(err)
		}
		log.Println(code, "was unzipped.")

		log.Println(csv, "splitting...")
		if strings.Contains(csv, "IPV6") {
			state.IPv6Chunks, err = s.split(csv)
			if err != nil {
				log.Panic(err)
			}
		} else {
			state.IPv4Chunks, err = s.split(csv)
			if err != nil {
				log.Panic(err)
			}
		}
		log.Println(csv, "was splitted.")
	}

	wg := &sync.WaitGroup{}

	for _, code := range []string{codeIPv4, codeIPv6} {
		wg.Add(1)

		c := code
		go func() {
			defer wg.Done()
			update(s, c)
		}()

	}

	wg.Wait()
}

// Split csv file specified by name on smaller chunks and return a filepaths of chunks.
func (s *State) split(name string) (chunks []string, err error) {
	splitter := scsv.New()
	splitter.FileChunkSize = 100_000_000 //in bytes (100MB)
	splitter.WithHeader = false          //copying of header in chunks is disabled

	chunks, err = splitter.Split(s.wd+string(os.PathSeparator)+name, s.wd)
	return
}

// Unzip file specified by zipName and return a extracted fileName.
func (s *State) extract(zipName string) (fileName string, err error) {
	srcFilePath := s.wd + string(os.PathSeparator) + zipName + ".zip"
	r, err := zip.OpenReader(srcFilePath)
	if err != nil {
		return
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(s.wd, 0755)

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
		name, err := extractAndWriteFile(s.wd, f)
		if err != nil {
			return "", err
		}
		if strings.Contains(name, ".CSV") {
			fileName = name
		}
	}

	return fileName, nil
}

// Download IP2Location database specified by code.
func (s *State) download(code string) (err error) {
	token := os.Getenv("IP2LOCATION_TOKEN")
	url := fmt.Sprintf("%s?token=%s&file=%s", baseUrl, token, code)

	destFilePath := s.wd + string(os.PathSeparator) + code + ".zip"
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
