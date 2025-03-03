package provider

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	// Test creation of a new Location
	firstIP := big.NewInt(0)
	lastIP := big.NewInt(100)
	code := "US"
	country := "United States"
	region := "California"
	city := "San Francisco"
	latitude := "37.7749"
	longitude := "-122.4194"
	zipCode := "94105"
	timeZone := "-07:00"

	// Create new location
	location := New(firstIP, lastIP, code, country, region, city, latitude, longitude, zipCode, timeZone)

	// Check that location is not nil
	require.NotNil(t, location, "Location should not be nil")

	// Check that FirstIP and LastIP are set correctly
	assert.Equal(t, firstIP, location.FirstIP, "FirstIP should match")
	assert.Equal(t, lastIP, location.LastIP, "LastIP should match")

	// Check that Properties map is initialized
	require.NotNil(t, location.Properties, "Properties map should not be nil")

	// Check that all properties are set correctly
	assert.Equal(t, code, location.Properties[Code], "Code property should match")
	assert.Equal(t, country, location.Properties[Country], "Country property should match")
	assert.Equal(t, region, location.Properties[Region], "Region property should match")
	assert.Equal(t, city, location.Properties[City], "City property should match")
	assert.Equal(t, latitude, location.Properties[Latitude], "Latitude property should match")
	assert.Equal(t, longitude, location.Properties[Longitude], "Longitude property should match")
	assert.Equal(t, zipCode, location.Properties[ZipCode], "ZipCode property should match")
	assert.Equal(t, timeZone, location.Properties[TimeZone], "TimeZone property should match")
}

func TestLocationString(t *testing.T) {
	// Test cases
	testCases := []struct {
		name     string
		location *Location
		expected string
	}{
		{
			name: "Complete location data",
			location: New(
				big.NewInt(0),
				big.NewInt(100),
				"US",
				"United States",
				"California",
				"San Francisco",
				"37.7749",
				"-122.4194",
				"94105",
				"-07:00",
			),
			expected: `{"Code":"US","Country":"United States","Region":"California","City":"San Francisco","Latitude":"37.7749","Longitude":"-122.4194","ZipCode":"94105","TimeZone":"-07:00"}`,
		},
		{
			name: "Empty location data",
			location: New(
				big.NewInt(0),
				big.NewInt(100),
				"",
				"",
				"",
				"",
				"",
				"",
				"",
				"",
			),
			expected: `{"Code":"","Country":"","Region":"","City":"","Latitude":"","Longitude":"","ZipCode":"","TimeZone":""}`,
		},
		{
			name: "Partial location data",
			location: New(
				big.NewInt(0),
				big.NewInt(100),
				"JP",
				"Japan",
				"",
				"Tokyo",
				"35.6762",
				"139.6503",
				"",
				"+09:00",
			),
			expected: `{"Code":"JP","Country":"Japan","Region":"","City":"Tokyo","Latitude":"35.6762","Longitude":"139.6503","ZipCode":"","TimeZone":"+09:00"}`,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Get string representation
			result := tc.location.String()

			// Assert that the result matches expected
			assert.Equal(t, tc.expected, result, "String representation should match expected")
		})
	}
}

func TestConstants(t *testing.T) {
	// Test that constants are defined with expected values
	assert.Equal(t, "https://www.ip2location.com/download", DefaultURL, "DefaultURL should match expected value")
	assert.Equal(t, "DB11LITEIPV6", DefaultCode, "DefaultCode should match expected value")
	assert.Equal(t, "test/data/", ZipPath, "ZipPath should match expected value")

	// Test property constants
	assert.Equal(t, Properties("Code"), Code, "Code property should match expected value")
	assert.Equal(t, Properties("Country"), Country, "Country property should match expected value")
	assert.Equal(t, Properties("Region"), Region, "Region property should match expected value")
	assert.Equal(t, Properties("City"), City, "City property should match expected value")
	assert.Equal(t, Properties("Latitude"), Latitude, "Latitude property should match expected value")
	assert.Equal(t, Properties("Longitude"), Longitude, "Longitude property should match expected value")
	assert.Equal(t, Properties("ZipCode"), ZipCode, "ZipCode property should match expected value")
	assert.Equal(t, Properties("TimeZone"), TimeZone, "TimeZone property should match expected value")
}

func TestLocationProperties(t *testing.T) {
	// Create a location
	location := New(
		big.NewInt(0),
		big.NewInt(100),
		"RU",
		"Russia",
		"Moscow",
		"Moscow",
		"55.7558",
		"37.6173",
		"101000",
		"+03:00",
	)

	// Test accessing properties
	tests := []struct {
		property Properties
		expected string
	}{
		{Code, "RU"},
		{Country, "Russia"},
		{Region, "Moscow"},
		{City, "Moscow"},
		{Latitude, "55.7558"},
		{Longitude, "37.6173"},
		{ZipCode, "101000"},
		{TimeZone, "+03:00"},
	}

	for _, test := range tests {
		t.Run(string(test.property), func(t *testing.T) {
			value := location.Properties[test.property]
			assert.Equal(t, test.expected, value, "%s property should match expected value", test.property)
		})
	}
}

func TestBigIntHandling(t *testing.T) {
	// Test with different big.Int values

	// Max uint32 value
	maxUint32 := new(big.Int).SetUint64(4294967295)

	// IPv6 max value (2^128 - 1)
	maxIPv6 := new(big.Int)
	maxIPv6.Exp(big.NewInt(2), big.NewInt(128), nil)
	maxIPv6.Sub(maxIPv6, big.NewInt(1))

	testCases := []struct {
		name    string
		firstIP *big.Int
		lastIP  *big.Int
	}{
		{"Zero values", big.NewInt(0), big.NewInt(0)},
		{"Max uint32", big.NewInt(0), maxUint32},
		{"Large IPv6 range", big.NewInt(0), maxIPv6},
		{"Negative first IP", big.NewInt(-1), big.NewInt(100)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			location := New(
				tc.firstIP,
				tc.lastIP,
				"XX",
				"Test Country",
				"Test Region",
				"Test City",
				"0.0000",
				"0.0000",
				"00000",
				"+00:00",
			)

			// Verify that the big.Int values are stored correctly
			assert.True(t, tc.firstIP.Cmp(location.FirstIP) == 0,
				"FirstIP should be equal to input value")
			assert.True(t, tc.lastIP.Cmp(location.LastIP) == 0,
				"LastIP should be equal to input value")
		})
	}
}
