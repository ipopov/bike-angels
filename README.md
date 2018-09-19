# bike-angels

~~~~
curl 'https://layer.bicyclesharing.net/map/v1/nyc/stations' > station-info
go build deserialize.go
./deserialize --station-info=./station-info
~~~~
