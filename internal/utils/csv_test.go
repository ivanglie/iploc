package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCSV_String(t *testing.T) {
	csv := &CSV{File: "file1.CSV", Size: 1024}
	assert.Equal(t, `{"CSV":{"File":"file1.CSV","Size":1024}}`, csv.String())
}

func TestUnzipCSV(t *testing.T) {
	// Test absolute file path
	csv, err := UnzipCSV("../../test/data/DB.zip")
	assert.NoError(t, err)
	// assert.Equal(t, filepath.Join("testdata", "test.CSV"), csv.File)
	// assert.EqualValues(t, 28, csv.Size)
	assert.Equal(t, "DB.CSV", filepath.Base(csv.File))

	// Test file not found
	_, err = UnzipCSV("not_found.zip")
	assert.ErrorIs(t, err, os.ErrNotExist)

	// Test invalid file path
	_, err = UnzipCSV("../test.zip")
	assert.Error(t, err)

	// Test that UnzipCSV handles invalid input file path
	csv, err = UnzipCSV("")
	assert.NoError(t, err)
	assert.Nil(t, csv)

	// Test that UnzipCSV handles invalid zip file path
	csv, err = UnzipCSV("invalid.zip")
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
	assert.EqualError(t, err, "incorrect filename")

	// Test file not found
	s, err = SplitCSV("not_found.csv", 1024)
	assert.Nil(t, s)
	assert.ErrorIs(t, err, os.ErrNotExist)
}
