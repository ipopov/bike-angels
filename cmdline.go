package main

import (
	"flag"
	//"fmt"
	//"io"
	"angels"
	"os"
)

func main() {
	fileName := flag.String("station-info", "", "")
	atMost := flag.Int("at-most", 10, "")
	flag.Parse()
	file, err := os.Open(*fileName)
	if err != nil {
		panic(err)
	}
	angels.Run(*atMost, file, os.Stdout)
}
