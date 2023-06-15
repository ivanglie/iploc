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
		StatusCode: 200,
		Body:       respBody,
	}, errors.New("something went wrong")
}

func TestDB_Prepare(t *testing.T) {
	customClient = &MockClient{}

	db := &DB{}
	err := db.Prepare("", "../../test/data/", 2)
	assert.NoError(t, err)

	// Empty path
	err = db.Prepare("", "", 2)
	assert.Error(t, err)
}

func TestDB_Search(t *testing.T) {
	db := &DB{}
	db.Chunks = []string{"../../test/data/DB_0001.CSV", "../../test/data/DB_0002.CSV", "../../test/data/DB_0003.CSV"}
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

func Test_split(t *testing.T) {
	// Test splitting a CSV file
	s, err := split("../../test/data/DB.CSV", 1024)
	assert.NoError(t, err)
	assert.Len(t, s, 3)

	// Test incorrect filename
	s, err = split("", 1024)
	assert.Nil(t, s)
	assert.EqualError(t, err, "incorrect path")

	// Test file not found
	s, err = split("not_found.csv", 1024)
	assert.Nil(t, s)
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func Test_unzip(t *testing.T) {
	// Test absolute file path
	csv, _, err := unzip("../../test/data/DB.zip")
	assert.NoError(t, err)
	assert.Equal(t, "DB.CSV", filepath.Base(csv))

	// Test file not found
	_, _, err = unzip("not_found.zip")
	assert.ErrorIs(t, err, os.ErrNotExist)

	// Test invalid file path
	_, _, err = unzip("../test.zip")
	assert.Error(t, err)

	// Test that unzip handles empty input file path
	csv, _, err = unzip("")
	assert.Error(t, err)
	assert.Empty(t, csv)

	// Test that unzip handles invalid zip file path
	csv, _, err = unzip("invalid.zip")
	assert.Error(t, err)
	assert.Empty(t, csv)
}

func Test_download(t *testing.T) {
	customClient = &MockClient{}

	n, s, err := download("", ".")
	assert.NoError(t, err)
	assert.Equal(t, "DB11LITEIPV6.zip", filepath.Base(n))
	assert.Equal(t, int64(1254), s)

	// Bad status
	customClient = &MockClientBadStatus{}

	_, _, err = download("", ".")
	assert.Error(t, err)

	// Error
	customClient = &MockClientError{}

	_, _, err = download("", ".")
	assert.Error(t, err)
}
