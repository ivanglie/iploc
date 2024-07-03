# iploc


[![Test](https://github.com/ivanglie/iploc/actions/workflows/test.yml/badge.svg)](https://github.com/ivanglie/iploc/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/ivanglie/iploc/branch/master/graph/badge.svg?token=sLJxFoa5EC)](https://codecov.io/gh/ivanglie/iploc)

This is a simple IP geolocation service built with Go, HTMX, and Bootstrap. The application can look up geolocation information based on an IP address.

## Features

  * Lookup geolocation information by IP address (IPv4 or IPv6).
  * Reduces CPU and RAM utilization by using a binary search algorithm on smaller chunks of a large raw data file, which has been pre-split
  * Auto downloading a database and preparing it for use, using a IP2Location Download Token
  * Returns the result as JSON or HTML based on the Accept header in the request
  * Simple web interface for entering an IP address and displaying results
  * Logging of search operations and results

## Example

Search location:

```code
curl http://localhost/search?ip=8.8.8.8 -H "Accept: application/json"
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
See [requests.http](./test/requests.http).

## Acknowledgment

This site or product includes IP2Location LITE data available from <a href="https://lite.ip2location.com">https://lite.ip2location.com</a>.

## License

This project is licensed under the MIT License - see the [LICENSE](/LICENSE.md) file for details.