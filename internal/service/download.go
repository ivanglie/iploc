package service

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ivanglie/iploc/internal/provider"
)

// download IP2Location database (specified by token) to path.
func (s *Service) Download(token, path string) (err error) {
	if len(path) == 0 {
		err = fmt.Errorf("empty path")
		return
	}

	var req *http.Request
	if req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s?token=%s&file=%s", provider.DefaultURL, token, provider.DefaultCode), nil); err != nil {
		return
	}

	var resp *http.Response
	if resp, err = s.httpClient.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("error %d %s", resp.StatusCode, resp.Status)
		return
	}

	if s.zip, err = filepath.Abs(filepath.Join(filepath.Dir(path), provider.DefaultCode+".zip")); err != nil {
		return
	}

	var file *os.File
	if file, err = os.OpenFile(s.zip, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fs.ModeAppend); err != nil {
		return
	}
	defer file.Close()

	if _, err = io.Copy(file, resp.Body); err != nil {
		return
	}

	if s.zipSize, err = Size(s.zip); err != nil {
		return err
	}

	return
}
