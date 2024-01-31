# iploc


[![Test](https://github.com/ivanglie/iploc/actions/workflows/test.yml/badge.svg)](https://github.com/ivanglie/iploc/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/ivanglie/iploc/branch/master/graph/badge.svg?token=sLJxFoa5EC)](https://codecov.io/gh/ivanglie/iploc)

Web service that identifies the country, region or state, city, latitude and longitude, ZIP/Postal code, and timezone based on an IP address (IPv4 or IPv6).

## Features

* Reduces CPU and RAM utilization by using a binary search algorithm on smaller chunks of a large raw data file, which has been pre-split

* Auto downloading a database and preparing it for use, using a IP2Location Download Token

## Example

Search location:

```code
curl http://localhost/search?ip=8.8.8.8
```

Output:
```json
{
  "Code": "US",
  "Country": "United States of America",
  "Region": "California",
  "City": "Mountain View",
  "Latitude": "37.405992",
  "Longitude": "-122.078515",
  "ZipCode": "94043",
  "TimeZone": "-08:00"
}
```
See [requests.rest](./test/requests.rest).

## Acknowledgment

This site or product includes IP2Location LITE data available from <a href="https://lite.ip2location.com">https://lite.ip2location.com</a>.