package provider

import (
	"fmt"
	"math/big"
)

const (
	DefaultURL  = "https://www.ip2location.com/download" // IP2Location API Download Link
	DefaultCode = "DB11LITEIPV6"                         // IP2Location IPv4 and IPv6 Database Code

	ZipPath     = "test/data/"
	ZipFileName = "DB.zip"

	Code      Properties = "Code"      // Two-character country code based on ISO 3166.
	Country   Properties = "Country"   // Country name based on ISO 3166.
	Region    Properties = "Region"    // Region or state name.
	City      Properties = "City"      // City name.
	Latitude  Properties = "Latitude"  // City latitude. Default to capital city latitude if city is unknown.
	Longitude Properties = "Longitude" // City longitude. Default to capital city longitude if city is unknown.
	ZipCode   Properties = "ZipCode"   // ZIP/Postal code.
	TimeZone  Properties = "TimeZone"  // UTC time zone (with DST supported).
)

type Properties string

type Loc struct {
	FirstIP    *big.Int `json:"-"` // First IP address show netblock.
	LastIP     *big.Int `json:"-"` // Last IP address show netblock.
	Properties map[Properties]string
}

func NewLoc(firstIP, lastIP *big.Int, code, country, region, city, latitude, longitude, zipCode, timeZone string) *Loc {
	loc := &Loc{FirstIP: firstIP, LastIP: lastIP}

	loc.Properties = make(map[Properties]string)
	loc.Properties[Code] = code
	loc.Properties[Country] = country
	loc.Properties[Region] = region
	loc.Properties[City] = city
	loc.Properties[Latitude] = latitude
	loc.Properties[Longitude] = longitude
	loc.Properties[ZipCode] = zipCode
	loc.Properties[TimeZone] = timeZone

	return loc
}

// String representation of *IP.
func (loc *Loc) String() string {
	p := loc.Properties
	return fmt.Sprintf(
		`{"Code":"%s","Country":"%s","Region":"%s","City":"%s","Latitude":"%s","Longitude":"%s","ZipCode":"%s","TimeZone":"%s"}`,
		p[Code], p[Country], p[Region], p[City], p[Latitude], p[Longitude], p[ZipCode], p[TimeZone])
}
