package angels

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sort"
)

type Location struct {
	Lat float64
	Lon float64
}

type Station struct {
	Name   string
	Points int
	Loc    Location
}

// http://en.wikipedia.org/wiki/Haversine_formula
func Distance(lat1, lon1, lat2, lon2 float64) float64 {
	var la1, lo1, la2, lo2 float64
	la1 = lat1 * math.Pi / 180
	lo1 = lon1 * math.Pi / 180
	la2 = lat2 * math.Pi / 180
	lo2 = lon2 * math.Pi / 180

	const earth_radius = 6378100.

	d_lat := la2 - la1
	d_lon := lo2 - lo1
	a := (math.Pow(math.Sin(d_lat/2), 2) +
		math.Cos(la1)*math.Cos(la2)*math.Pow(math.Sin(d_lon/2), 2))
	return 2 * earth_radius * math.Asin(math.Sqrt(a))
}

func (l Location) String() string {
	return fmt.Sprintf("(%0.03f, %0.03f)", l.Lat, l.Lon)
}

func (s Station) String() string {
	return fmt.Sprintf("%v %45s %2d", s.Loc, s.Name, s.Points)
}

func Run(atMost int, r io.Reader, out io.Writer) {
	var features struct {
		Features []struct {
			Geometry struct {
				Location []float64 `json:"coordinates"`
			} `json:"geometry"`
			Properties struct {
				Name   string
				Action string `json:"bike_angels_action"`
				Points int    `json:"bike_angels_points"`
			}
		}
	}

	if err := json.NewDecoder(r).Decode(&features); err != nil {
		panic(err)
	}

	var stations []Station

	for _, station := range features.Features {
		points := station.Properties.Points
		if station.Properties.Action == "give" {
			points = -points
		}
		stations = append(stations, Station{
			Name: station.Properties.Name,
			Loc: Location{
				Lat: station.Geometry.Location[1],
				Lon: station.Geometry.Location[0],
			},
			Points: points,
		})
	}

	type X struct {
		from, to    int
		dist        float64
		ptsPerMeter float64
	}
	var vals []X

	for i, x := range stations {
		for j, y := range stations {
			ptsDiff := float64(x.Points - y.Points)
			if ptsDiff <= 0 ||
				// No trips between stations of the same sign.
				(x.Points*y.Points) > 0 {
				continue
			}
			dist := Distance(x.Loc.Lat, x.Loc.Lon, y.Loc.Lat, y.Loc.Lon)
			ptsPerMeter := ptsDiff / dist
			vals = append(vals, X{from: i, to: j, ptsPerMeter: ptsPerMeter, dist: dist})
		}
	}

	sort.Slice(vals, func(x, y int) bool { return vals[x].ptsPerMeter > vals[y].ptsPerMeter })

	fmt.Fprintf(out, "<html><body><pre>\n")
	fmt.Fprintf(out, "<b>Best Bike Angels opportunities</b>\n\n")
	for i := 0; i < len(vals) && i < atMost; i++ {
		fmt.Fprintf(out, "Dist (m): %.0f /// Points per mile: %.1f\n", vals[i].dist, vals[i].ptsPerMeter*1609)
		url := fmt.Sprintf("https://www.google.com/maps/dir/?api=1&origin=%f,%f&destination=%f,%f&travelmode=bicycling",
			stations[vals[i].from].Loc.Lat,
			stations[vals[i].from].Loc.Lon,
			stations[vals[i].to].Loc.Lat,
			stations[vals[i].to].Loc.Lon)
		fmt.Fprintf(out, "<a href=\"%s\">%s <b>(%d)</b> to %s <b>(%d)</b></a>\n", url, stations[vals[i].from].Name, stations[vals[i].from].Points, stations[vals[i].to].Name, stations[vals[i].to].Points)
		fmt.Fprintf(out, "\n")
	}
	fmt.Fprintf(out, "</pre></body></html>\n")
}
