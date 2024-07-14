package database

import (
	"bufio"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockClient struct{}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	f, _ := os.Open("../../test/data/DB.zip")
	r := bufio.NewReader(f)
	respBody := io.NopCloser(r)

	return &http.Response{
		StatusCode: 200,
		Body:       respBody,
	}, nil
}

type badStatusClient struct{}

func (m *badStatusClient) Do(req *http.Request) (*http.Response, error) {
	f, _ := os.Open("../../test/data/DB.zip")
	r := bufio.NewReader(f)
	respBody := io.NopCloser(r)

	return &http.Response{
		StatusCode: 503,
		Status:     "Service Unavailable",
		Body:       respBody,
	}, nil
}

type errorClient struct{}

func (m *errorClient) Do(req *http.Request) (*http.Response, error) {
	f, _ := os.Open("../../test/data/DB.zip")
	r := bufio.NewReader(f)
	respBody := io.NopCloser(r)

	return &http.Response{
		StatusCode: 503,
		Status:     "Service Unavailable",
		Body:       respBody,
	}, errors.New("something went wrong")
}

func TestDB_Init(t *testing.T) {
	db := NewDB()
	db.downloadFunc = func(url, path string) error { return nil }

	// assert.NoError(t, db.Init(true, "token", "path"))

	// Download error
	db.downloadFunc = func(url, path string) error { return errors.New("download error") }
	assert.Error(t, db.Init(true, "token", "path"))
}

func TestDB_Search(t *testing.T) {
	db := &DB{chunks: []string{"../../test/data/DB_0001.CSV", "../../test/data/DB_0002.CSV", "../../test/data/DB_0003.CSV"}}
	loc, err := db.Search("8.8.8.8")
	assert.Nil(t, err)
	assert.NotNil(t, loc)
	assert.NotNil(t, loc.Properties)
	assert.Equal(t, "US", loc.Properties[Code])
	assert.Equal(t, "United States of America", loc.Properties[Country])
	assert.Equal(t, "California", loc.Properties[Region])
	assert.Equal(t, "Mountain View", loc.Properties[City])
	assert.Equal(t, "37.405992", loc.Properties[Latitude])
	assert.Equal(t, "-122.078515", loc.Properties[Longitude])
	assert.Equal(t, "94043", loc.Properties[ZipCode])
	assert.Equal(t, "-07:00", loc.Properties[TimeZone])
}

func TestDB_download(t *testing.T) {
	db := NewDB()
	db.httpClient = &mockClient{}
	err := db.download("token", "../../test/data/")
	assert.NoError(t, err)
	assert.Equal(t, code+".zip", filepath.Base(db.zip))
	assert.Equal(t, int64(1254), db.zipSize)

	os.Remove("../../test/data/" + code + ".zip")

	// Bad status error
	db = NewDB()
	db.httpClient = &badStatusClient{}
	err = db.download("token", "../../test/data/")
	assert.Equal(t, "error 503 Service Unavailable", err.Error())

	// Empty path error
	db = NewDB()
	err = db.download("token", "")
	assert.Equal(t, "empty path", err.Error())

	// Something went wrong error
	db = NewDB()
	db.httpClient = &errorClient{}
	err = db.download("token", "../../test/data/")
	assert.Equal(t, "something went wrong", err.Error())
}

func TestDB_String(t *testing.T) {
	db := &DB{BufferSize: 2, chunks: []string{"DB_0001.CSV", "DB_0002.CSV"}}
	assert.Equal(t, "DB{zip: , zipSize: 0, csv: , csvSize: 0, chunks: [DB_0001.CSV DB_0002.CSV], ChunksCount: 2}", db.String())
}
