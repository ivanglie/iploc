package iploc

import (
	"encoding/csv"
	"math/big"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var data [][]string

func setupTest(t *testing.T) {
	t.Log("Setup test")

	f, err := os.Open("../../test/data/DATA.CSV")
	if err != nil {
		t.Logf("Unable to read input file:\n%v", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	data, err = r.ReadAll()
	if err != nil {
		t.Logf("Unable to parse file as CSV:\n%v", err)
	}
}

func teardownTest(t *testing.T) {
	t.Log("Teardown test")
}

func TestSearch(t *testing.T) {
	loc, err := Search("8.8.8.8", []string{"../../test/data/DB_0001.CSV", "../../test/data/DB_0002.CSV", "../../test/data/DB_0003.CSV"})
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

	// Errors
	loc, err = Search("8.8.8.", []string{"../../test/data/DB_0001.CSV", "../../test/data/DB_0002.CSV", "../../test/data/DB_0003.CSV"})
	assert.Nil(t, loc)
	assert.Equal(t, err.Error(), "address ::ffff:8.8.8. is incorrect IP")

	loc, err = Search("8.8.8.8", []string{})
	assert.Nil(t, loc)
	assert.Equal(t, err.Error(), "chunks is empty or not found")

	loc, err = Search("9.9.9.9", []string{"../../test/data/DB_0001.CSV", "../../test/data/DB_0002.CSV", "../../test/data/DB_0003.CSV"})
	assert.Nil(t, loc)
	assert.Equal(t, err.Error(), "281470833330441 not found")
}

func Test_searchChunk(t *testing.T) {
	c, err := searchChunk(big.NewInt(281470816487424),
		[]string{"../../test/data/DB_0001.CSV", "../../test/data/DB_0002.CSV", "../../test/data/DB_0003.CSV"})
	assert.Nil(t, err)
	assert.NotNil(t, c)

	c, err = searchChunk(big.NewInt(281470816486912),
		[]string{"../../test/data/DB_0001.CSV", "../../test/data/DB_0002.CSV", "../../test/data/DB_0003.CSV"})
	assert.Nil(t, err)
	assert.NotNil(t, c)

	n, _ := new(big.Int).SetString("42541956123769884654463883030277652480", 0)
	c, err = searchChunk(n, []string{"../../test/data/DB_0001.CSV", "../../test/data/DB_0002.CSV", "../../test/data/DB_0003.CSV"})
	assert.Nil(t, err)
	assert.NotNil(t, c)

	// Errors
	// Not found
	c, err = searchChunk(big.NewInt(281470816482303),
		[]string{"../../test/data/DB_0001.CSV", "../../test/data/DB_0002.CSV", "../../test/data/DB_0003.CSV"})
	assert.Nil(t, c)
	assert.Equal(t, err.Error(), "chunks is empty or not found")

	// Chunks is empty
	c, err = searchChunk(big.NewInt(281470816487424), []string{})
	assert.Nil(t, c)
	assert.Equal(t, err.Error(), "chunks is empty or not found")

	// Incorrect format
	c, err = searchChunk(big.NewInt(281470816487424), []string{"../../test/data/DBincorrect.CSV"})
	assert.Nil(t, c)
	assert.Equal(t, err.Error(), "record on line 5: wrong number of fields")

	// No such file or directory
	c, err = searchChunk(big.NewInt(281470816487424),
		[]string{"../../test/data/DB_0001.CSV", "../../test/data/DB_0002csv", "../../test/data/DB_0003.CSV"})
	assert.Nil(t, c)
	assert.Equal(t, err.Error(), "open ../../test/data/DB_0002csv: no such file or directory")
}

func Test_searchByNum_IPv4(t *testing.T) {
	setupTest(t)
	defer teardownTest(t)

	n, _ := convertIP("8.8.8.8")
	loc, _ := searchByNum(n, data)
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

func Test_searchByNum_IPv6(t *testing.T) {
	setupTest(t)
	defer teardownTest(t)

	n, _ := convertIP("2001:4860:4860:0:0:0:0:8888")
	loc, _ := searchByNum(n, data)
	assert.NotNil(t, loc)
	assert.NotNil(t, loc.Properties)
	assert.Equal(t, "GB", loc.Properties[Code])
	assert.Equal(t, "United Kingdom of Great Britain and Northern Ireland", loc.Properties[Country])
	assert.Equal(t, "England", loc.Properties[Region])
	assert.Equal(t, "Upper Clapton", loc.Properties[City])
	assert.Equal(t, "51.564000", loc.Properties[Latitude])
	assert.Equal(t, "-0.058080", loc.Properties[Longitude])
	assert.Equal(t, "E5", loc.Properties[ZipCode])
	assert.Equal(t, "+01:00", loc.Properties[TimeZone])
}

func Test_searchByNum_Errors(t *testing.T) {
	setupTest(t)
	defer teardownTest(t)

	// Not found
	n, _ := convertIP("9.9.9.9")
	ip, err := searchByNum(n, data)
	assert.Nil(t, ip)
	assert.Equal(t, err.Error(), "281470833330441 not found")
}

func Test_convertIP_IPv4(t *testing.T) {
	expectedNum, _ := new(big.Int).SetString("281473391529217", 0)
	num, err := convertIP("161.132.13.1")
	assert.Nil(t, err)
	assert.Equal(t, expectedNum, num)
}

func Test_convertIP_IPv6(t *testing.T) {
	expectedNum, _ := new(big.Int).SetString("42540766411282594074389245746715063092", 0)
	num, err := convertIP("2001:0db8:0000:0042:0000:8a2e:0370:7334")
	assert.Nil(t, err)
	assert.Equal(t, expectedNum, num)
}

func Test_convertIP_Errors(t *testing.T) {
	// Empty address
	num, err := convertIP("")
	assert.Nil(t, num)
	assert.Equal(t, err.Error(), "empty address")

	// Incorrect IP
	num, err = convertIP("8.8.8.")
	assert.Nil(t, num)
	assert.Equal(t, err.Error(), "address ::ffff:8.8.8. is incorrect IP")
}

func TestLocString(t *testing.T) {
	loc := newLoc(
		big.NewInt(281470816487424),
		big.NewInt(281470816487679),
		"US",
		"United States of America",
		"California",
		"Mountain View",
		"37.405992",
		"-122.078515",
		"94043",
		"-07:00")

	assert.Equal(t, `{`+
		`"Code":"US",`+
		`"Country":"United States of America",`+
		`"Region":"California",`+
		`"City":"Mountain View",`+
		`"Latitude":"37.405992",`+
		`"Longitude":"-122.078515",`+
		`"ZipCode":"94043",`+
		`"TimeZone":"-07:00"`+
		`}`, loc.String())
}

func TestLocString_Errors(t *testing.T) {
	// Empty loc
	loc := &Loc{}

	assert.Equal(t, `{`+
		`"Code":"",`+
		`"Country":"",`+
		`"Region":"",`+
		`"City":"",`+
		`"Latitude":"",`+
		`"Longitude":"",`+
		`"ZipCode":"",`+
		`"TimeZone":""`+
		`}`, loc.String())
}
