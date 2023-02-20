# iploc

Web service that identifies the country, region or state, city, latitude and longitude, ZIP/Postal code, and timezone based on an IP address (IPv4 or IPv6).

## Features

* Reduces CPU and RAM utilization by using a binary search algorithm on smaller chunks of a large raw data file, which has been pre-split

* Procedure for downloading a database

## API

* Download DB
```code
POST http://localhost:9000/download HTTP/1.1
content-type: application/json

{
    "token": "YOUR-TOKEN"
}
```

* Unzip file
```code
GET http://localhost:9000/unzip
```

* Split CSV
```code
GET http://localhost:9000/split
```

* Search location
```code
GET http://localhost:9000/search?ip=8.8.8.8
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

## Acknowledgment

This site or product includes IP2Location LITE data available from <a href="https://lite.ip2location.com">https://lite.ip2location.com</a>.