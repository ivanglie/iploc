package utils

import (
	"bufio"
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

func TestCSV_String(t *testing.T) {
	csv := &CSV{File: "file1.CSV", Size: 1024}
	assert.Equal(t, `{"File":"file1.CSV","Size":1024}`, csv.String())
}

func TestUnzipCSV(t *testing.T) {
	// Test absolute file path
	csv, err := Unzip("../../test/data/DB.zip")
	assert.NoError(t, err)
	// assert.Equal(t, filepath.Join("testdata", "test.CSV"), csv.File)
	// assert.EqualValues(t, 28, csv.Size)
	assert.Equal(t, "DB.CSV", filepath.Base(csv.File))

	// Test file not found
	_, err = Unzip("not_found.zip")
	assert.ErrorIs(t, err, os.ErrNotExist)

	// Test invalid file path
	_, err = Unzip("../test.zip")
	assert.Error(t, err)

	// Test that UnzipCSV handles invalid input file path
	csv, err = Unzip("")
	assert.Error(t, err)
	assert.Nil(t, csv)

	// Test that UnzipCSV handles invalid zip file path
	csv, err = Unzip("invalid.zip")
	assert.Error(t, err)
	assert.Nil(t, csv)
}

func TestSplitCSV(t *testing.T) {
	// Test splitting a CSV file
	s, err := SplitCSV("../../test/data/DB.CSV", 1024)
	assert.NoError(t, err)
	assert.Len(t, s, 3)

	// Test incorrect filename
	s, err = SplitCSV("", 1024)
	assert.Nil(t, s)
	assert.EqualError(t, err, "incorrect path")

	// Test file not found
	s, err = SplitCSV("not_found.csv", 1024)
	assert.Nil(t, s)
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestDownload(t *testing.T) {
	customClient = &MockClient{}

	n, s, err := Download("", ".")
	assert.NoError(t, err)
	assert.Equal(t, "DB11LITEIPV6.zip", filepath.Base(n))
	assert.Equal(t, int64(1254), s)

	os.Remove(n)
}
