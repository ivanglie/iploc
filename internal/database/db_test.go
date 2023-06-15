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

type MockClient struct {
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	f, _ := os.Open("../../test/data/DB.zip")
	r := bufio.NewReader(f)
	respBody := io.NopCloser(r)

	return &http.Response{
		StatusCode: 200,
		Body:       respBody,
	}, nil
}

type MockClientBadStatus struct {
}

func (m *MockClientBadStatus) Do(req *http.Request) (*http.Response, error) {
	f, _ := os.Open("../../test/data/DB.zip")
	r := bufio.NewReader(f)
	respBody := io.NopCloser(r)

	return &http.Response{
		StatusCode: 503,
		Status:     "Service Unavailable",
		Body:       respBody,
	}, nil
}

type MockClientError struct {
}

func (m *MockClientError) Do(req *http.Request) (*http.Response, error) {
	f, _ := os.Open("../../test/data/DB.zip")
	r := bufio.NewReader(f)
	respBody := io.NopCloser(r)

	return &http.Response{
		StatusCode: 503,
		Status:     "Service Unavailable",
		Body:       respBody,
	}, errors.New("something went wrong")
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

func TestDB_Split(t *testing.T) {
	f, _ := os.Open("../../test/data/DB.zip")
	info, _ := f.Stat()
	csvSize := info.Size()

	db := &DB{csv: "../../test/data/DB.CSV", CSVSize: csvSize, BufferSize: 1024}
	err := db.Split()
	assert.NoError(t, err)
	assert.Equal(t, "DB_0001.CSV", filepath.Base(db.chunks[0]))
	assert.Equal(t, "DB_0002.CSV", filepath.Base(db.chunks[1]))
	assert.Equal(t, "DB_0003.CSV", filepath.Base(db.chunks[2]))

	// File not found
	db = &DB{csv: "../../test/data/DB1.CSV", CSVSize: csvSize, BufferSize: 1024}
	err = db.Split()
	assert.ErrorIs(t, err, os.ErrNotExist)

	// db.csvSize is 0 error
	db = &DB{csv: "../../test/data/DB.CSV"}
	err = db.Split()
	assert.Equal(t, "db.csvSize is 0", err.Error())

	// Empty db.csv error
	db = &DB{}
	err = db.Split()
	assert.Equal(t, "empty db.csv", err.Error())
}

func TestDB_Unzip(t *testing.T) {
	db := &DB{zip: "../../test/data/DB.zip"}
	err := db.Unzip()
	assert.NoError(t, err)
	assert.Equal(t, "DB.CSV", filepath.Base(db.csv))
	assert.Equal(t, int64(3068), db.CSVSize)

	// Empty db.zip error
	db = &DB{}
	err = db.Unzip()
	assert.Equal(t, "empty db.zip", err.Error())

	// File not found
	db = &DB{zip: "../../test/data/DB1.zip"}
	err = db.Unzip()
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestDB_Download(t *testing.T) {
	customClient = &MockClient{}

	db := &DB{}
	err := db.Download("token", "../../test/data/")
	assert.NoError(t, err)
	assert.Equal(t, code+".zip", filepath.Base(db.zip))
	assert.Equal(t, int64(1254), db.zipSize)

	os.Remove("../../test/data/" + code + ".zip")

	// Bad status error
	customClient = &MockClientBadStatus{}

	db = &DB{}
	err = db.Download("token", "../../test/data/")
	assert.Equal(t, "error 503 Service Unavailable", err.Error())

	// Empty path error
	db = &DB{}
	err = db.Download("token", "")
	assert.Equal(t, "empty path", err.Error())

	// Something went wrong error
	customClient = &MockClientError{}

	db = &DB{}
	err = db.Download("token", "../../test/data/")
	assert.Equal(t, "something went wrong", err.Error())
}

func TestDB_String(t *testing.T) {
	db := &DB{BufferSize: 2, chunks: []string{"DB_0001.CSV", "DB_0002.CSV"}}
	assert.Equal(t, "DB{zip: , zipSize: 0, csv: , csvSize: 0, chunks: [DB_0001.CSV DB_0002.CSV], ChunksCount: 2}", db.String())
}
