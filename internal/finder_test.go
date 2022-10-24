package internal

import (
	"math/big"
	"testing"

	"github.com/spf13/afero"

	"github.com/stretchr/testify/assert"
)

const (
	dataIPv4 = `"134736896","134737407","US","United States of America","Texas","Dallas","32.783060","-96.806670","75201","-05:00"
"134737408","134737663","US","United States of America","Louisiana","Monroe","32.548330","-92.045238","71203","-05:00"
"134737664","134737919","US","United States of America","New Jersey","Newark","40.732119","-74.173605","07101","-04:00"
"134737920","134738943","US","United States of America","Louisiana","Monroe","32.548330","-92.045238","71203","-05:00"
"134738944","134739199","CA","Canada","Quebec","Montreal","45.508840","-73.587810","H1A 0A1","-04:00"
"134739200","134739455","US","United States of America","Georgia","Atlanta","33.749000","-84.387980","30301","-04:00"
"134739456","134739711","US","United States of America","New Jersey","Newark","40.733400","-74.173500","07102","-04:00"
"134739712","134743039","US","United States of America","Louisiana","Monroe","32.548330","-92.045238","71203","-05:00"
"134743040","134743295","US","United States of America","California","Mountain View","37.405992","-122.078515","94043","-07:00"
"134743296","134743551","US","United States of America","Pennsylvania","Philadelphia","39.952340","-75.163790","19019","-04:00"
"134743552","134743807","US","United States of America","Pennsylvania","Newtown Square","39.983484","-75.414006","19073","-04:00"
"134743808","134744063","US","United States of America","Texas","Dallas","32.783060","-96.806670","75201","-05:00"
"134744064","134744319","US","United States of America","California","Mountain View","37.405992","-122.078515","94043","-07:00"
"134744320","134744575","US","United States of America","Florida","Fort Lauderdale","26.183096","-80.173925","33309","-04:00"
"134744576","134744831","US","United States of America","District of Columbia","Washington","38.895110","-77.036370","20001","-04:00"
"134744832","134750207","US","United States of America","Louisiana","Monroe","32.548330","-92.045238","71203","-05:00"
"134750208","134750463","US","United States of America","Minnesota","Plymouth","45.047700","-93.425940","55442","-05:00"
"134750464","134750719","US","United States of America","Louisiana","Monroe","32.548330","-92.045238","71203","-05:00"
"134750720","134750975","US","United States of America","Maryland","Baltimore","39.290380","-76.612190","21201","-04:00"
"134750976","134751999","US","United States of America","Louisiana","Monroe","32.548330","-92.045238","71203","-05:00"`

	dataIPv6 = `"42541956101379402827307645057575682048","42541956101379402845754389131285233663","UY","Uruguay","Montevideo","Montevideo","-34.833460","-56.167350","12300","-03:00"
"42541956101379402845754389131285233664","42541956101379402864201133204994785279","US","United States of America","California","Mountain View","37.386050","-122.083850","94041","-07:00"
"42541956101379402864201133204994785280","42541956101379402882647877278704336895","VE","Venezuela (Bolivarian Republic of)","Distrito Capital","Caracas","10.488010","-66.879190","1050","-04:00"
"42541956101379402882647877278704336896","42541956123769884636017138956568100863","US","United States of America","California","Mountain View","37.386050","-122.083850","94041","-07:00"
"42541956123769884636017138956568100864","42541956123769884654463883030277652479","GB","United Kingdom of Great Britain and Northern Ireland","England","Upper Clapton","51.564000","-0.058080","E5","+01:00"
"42541956123769884654463883030277652480","42541956180599069564461627201156022271","US","United States of America","California","Mountain View","37.386050","-122.083850","94041","-07:00"
"42541956180599069564461627201156022272","42541956735196207164311990355963674623","-","-","-","-","0.000000","0.000000","-","-"
"42541956735196207164311990355963674624","42541956814424369678576327949507624959","US","United States of America","Virginia","Ashburn","39.039474","-77.491809","20147","-04:00"
"42541956814424369678576327949507624960","42541957369021507278426691104315277311","-","-","-","-","0.000000","0.000000","-","-"
"42541957369021507278426691104315277312","42541957369031178684983608137712926719","US","United States of America","Louisiana","Monroe","32.524505","-92.128516","71207","-05:00"
"42541957369031178684983608137712926720","42541957369031232992198161138632884227","US","United States of America","Georgia","Alpharetta","34.075380","-84.294090","30239","-04:00"
`
)

var (
	fileIPv4 afero.File
	fileIPv6 afero.File
)

func setupTest(t *testing.T) {
	t.Log("Setup test")

	fs := afero.NewOsFs()
	fileIPv4, _ = fs.Create("testIPv4.csv")
	fileIPv4.WriteString(dataIPv4)

	fileIPv6, _ = fs.Create("testIPv6.csv")
	fileIPv6.WriteString(dataIPv6)
}

func teardownTest(t *testing.T) {
	t.Log("Teardown test")

	afero.NewOsFs().Remove("testIPv4.csv")
	afero.NewOsFs().Remove("testIPv6.csv")
}

func TestSearch(t *testing.T) {
	setupTest(t)
	defer teardownTest(t)

	ipv4, _ := Search("8.8.8.8", []string{fileIPv4.Name()}...)
	assert.Equal(t, "US", ipv4.Code)
	assert.Equal(t, "United States of America", ipv4.Country)
	assert.Equal(t, "California", ipv4.Region)
	assert.Equal(t, "Mountain View", ipv4.City)
	assert.Equal(t, "37.405992", ipv4.Latitude)
	assert.Equal(t, "-122.078515", ipv4.Longitude)
	assert.Equal(t, "94043", ipv4.ZipCode)
	assert.Equal(t, "-07:00", ipv4.TimeZone)

	ipv6, _ := Search("2001:4860:4860:0:0:0:0:8888", []string{fileIPv6.Name()}...)
	assert.Equal(t, "GB", ipv6.Code)
	assert.Equal(t, "United Kingdom of Great Britain and Northern Ireland", ipv6.Country)
	assert.Equal(t, "England", ipv6.Region)
	assert.Equal(t, "Upper Clapton", ipv6.City)
	assert.Equal(t, "51.564000", ipv6.Latitude)
	assert.Equal(t, "-0.058080", ipv6.Longitude)
	assert.Equal(t, "E5", ipv6.ZipCode)
	assert.Equal(t, "+01:00", ipv6.TimeZone)
}

func Test_convIP(t *testing.T) {
	num, err := convIP("161.132.13.1")
	assert.Nil(t, err)
	assert.Equal(t, new(big.Int).SetInt64(2709785857), num)

	expectedNum, _ := new(big.Int).SetString("42540766411282594074389245746715063092", 0)
	num, err = convIP("2001:0db8:0000:0042:0000:8a2e:0370:7334")
	assert.Nil(t, err)
	assert.Equal(t, expectedNum, num)
}

func Test_openCSV(t *testing.T) {
	setupTest(t)
	defer teardownTest(t)

	file, err := openCSV("../cmd/app/" + fileIPv6.Name())
	assert.NotNil(t, err)
	assert.Nil(t, file)

	file, err = openCSV(fileIPv4.Name())
	assert.Nil(t, err)
	assert.NotNil(t, file)
}
