package internal

import (
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const data = `"281470816482304","281470816482559","CA","Canada","Quebec","Montreal","45.508840","-73.587810","H1A 0A1","-04:00"
"281470816482560","281470816482815","US","United States of America","Georgia","Atlanta","33.749000","-84.387980","30301","-04:00"
"281470816482816","281470816483071","US","United States of America","New Jersey","Newark","40.733400","-74.173500","07102","-04:00"
"281470816483072","281470816486399","US","United States of America","Louisiana","Monroe","32.548330","-92.045238","71203","-05:00"
"281470816486400","281470816486655","US","United States of America","California","Mountain View","37.405992","-122.078515","94043","-07:00"
"281470816486656","281470816486911","US","United States of America","Pennsylvania","Philadelphia","39.952340","-75.163790","19019","-04:00"
"281470816486912","281470816487167","US","United States of America","Pennsylvania","Newtown Square","39.983484","-75.414006","19073","-04:00"
"281470816487168","281470816487423","US","United States of America","Texas","Dallas","32.783060","-96.806670","75201","-05:00"
"281470816487424","281470816487679","US","United States of America","California","Mountain View","37.405992","-122.078515","94043","-07:00"
"281470816487680","281470816487935","US","United States of America","Florida","Fort Lauderdale","26.183096","-80.173925","33309","-04:00"
"42541956101379402827307645057575682048","42541956101379402845754389131285233663","UY","Uruguay","Montevideo","Montevideo","-34.833460","-56.167350","12300","-03:00"
"42541956101379402845754389131285233664","42541956101379402864201133204994785279","US","United States of America","California","Mountain View","37.386050","-122.083850","94041","-07:00"
"42541956101379402864201133204994785280","42541956101379402882647877278704336895","VE","Venezuela (Bolivarian Republic of)","Distrito Capital","Caracas","10.488010","-66.879190","1050","-04:00"
"42541956101379402882647877278704336896","42541956123769884636017138956568100863","US","United States of America","California","Mountain View","37.386050","-122.083850","94041","-07:00"
"42541956123769884636017138956568100864","42541956123769884654463883030277652479","GB","United Kingdom of Great Britain and Northern Ireland","England","Upper Clapton","51.564000","-0.058080","E5","+01:00"
"42541956123769884654463883030277652480","42541956180599069564461627201156022271","US","United States of America","California","Mountain View","37.386050","-122.083850","94041","-07:00"
"42541956180599069564461627201156022272","42541956735196207164311990355963674623","-","-","-","-","0.000000","0.000000","-","-"
"42541956735196207164311990355963674624","42541956814424369678576327949507624959","US","United States of America","Virginia","Ashburn","39.039474","-77.491809","20147","-04:00"
"42541956814424369678576327949507624960","42541957369021507278426691104315277311","-","-","-","-","0.000000","0.000000","-","-"
"42541957369021507278426691104315277312","42541957369031178684983608137712926719","US","United States of America","Louisiana","Monroe","32.524505","-92.128516","71207","-05:00"
"42541957369031178684983608137712926720","42541957369031232992198161138632884227","US","United States of America","Georgia","Alpharetta","34.075380","-84.294090","30239","-04:00"`

var (
	db *DB
)

func setupTest(t *testing.T) {
	t.Log("Setup test")

	db = NewDB()
	d := strings.Replace(data, "\"", "", -1)
	a := strings.Split(d, "\n")
	for _, v := range a {
		b := strings.Split(v, ",")
		db.rec = append(db.rec, b)
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
