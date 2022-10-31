package internal

import (
	"encoding/csv"
	"log"
	"math/big"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var db *DB

func setupTest(t *testing.T) {
	t.Log("Setup test")

	f, _ := os.Open("../test/test.csv")
	defer f.Close()

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = 10

	var err error
	db = NewDB()
	db.rec, err = reader.ReadAll()
	if err != nil {
		log.Panic(err)
	}
}

func teardownTest(t *testing.T) {
	t.Log("Teardown test")

}

func TestSearch(t *testing.T) {
	setupTest(t)
	defer teardownTest(t)

	ip, _ := db.Search("8.8.8.8")
	assert.Equal(t, "US", ip.Code)
	assert.Equal(t, "United States of America", ip.Country)
	assert.Equal(t, "California", ip.Region)
	assert.Equal(t, "Mountain View", ip.City)
	assert.Equal(t, "37.405992", ip.Latitude)
	assert.Equal(t, "-122.078515", ip.Longitude)
	assert.Equal(t, "94043", ip.ZipCode)
	assert.Equal(t, "-07:00", ip.TimeZone)

	ip, _ = db.Search("2001:4860:4860:0:0:0:0:8888")
	assert.Equal(t, "GB", ip.Code)
	assert.Equal(t, "United Kingdom of Great Britain and Northern Ireland", ip.Country)
	assert.Equal(t, "England", ip.Region)
	assert.Equal(t, "Upper Clapton", ip.City)
	assert.Equal(t, "51.564000", ip.Latitude)
	assert.Equal(t, "-0.058080", ip.Longitude)
	assert.Equal(t, "E5", ip.ZipCode)
	assert.Equal(t, "+01:00", ip.TimeZone)
}

func Test_convertIP(t *testing.T) {
	expectedNum, _ := new(big.Int).SetString("281473391529217", 0)
	num, err := convertIP("161.132.13.1")
	assert.Nil(t, err)
	assert.Equal(t, expectedNum, num)

	expectedNum, _ = new(big.Int).SetString("42540766411282594074389245746715063092", 0)
	num, err = convertIP("2001:0db8:0000:0042:0000:8a2e:0370:7334")
	assert.Nil(t, err)
	assert.Equal(t, expectedNum, num)
}
