package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
)

type Location struct {
	Lat float32
	Lon float32
}

type Station struct {
	Name   string
	Points int
	Loc    Location
}

func (l Location) String() string {
  return fmt.Sprintf("(%0.03f, %0.03f)", l.Lat, l.Lon)
}

func (s Station) String() string {
  return fmt.Sprintf("%v %45s %2d", s.Loc, s.Name, s.Points)
}

func main() {
	fileName := flag.String("station-info", "", "")
	flag.Parse()

	file, err := os.Open(*fileName)
	if err != nil {
		panic(err)
	}

	var features struct {
		Features []struct {
			Geometry struct {
				Location []float32 `json:"coordinates"`
			} `json:"geometry"`
			Properties struct {
				Name   string
				Action string `json:"bike_angels_action"`
				Points int    `json:"bike_angels_points"`
			}
		}
	}

	if err = json.NewDecoder(file).Decode(&features); err != nil {
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
				Lat: station.Geometry.Location[0],
				Lon: station.Geometry.Location[1],
			},
			Points: points,
		})
	}

	sort.Slice(stations, func(x, y int) bool { return stations[x].Points < stations[y].Points })

	for _, station := range stations {
		if station.Points == 0 {
		  continue
		}
		fmt.Printf("%+v\n", station)
	}
}
