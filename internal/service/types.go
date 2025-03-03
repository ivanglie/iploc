package service

import (
	"fmt"
	"net/http"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Service struct {
	httpClient httpClient

	zip        string
	zipSize    int64
	csv        string
	CSVSize    int64
	chunks     []string
	BufferSize int64
}

func New() *Service {
	return &Service{
		httpClient: &http.Client{},
	}
}

// String returns a string representation of the DB struct.
func (s *Service) String() string {
	return fmt.Sprintf("DB{zip: %s, zipSize: %d, csv: %s, csvSize: %d, chunks: %v, ChunksCount: %d}",
		s.zip, s.zipSize, s.csv, s.CSVSize, s.chunks, s.BufferSize)
}
