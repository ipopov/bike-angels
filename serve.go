package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"
	//"fmt"
	"angels"
	"log"
	"net/http"
	"strings"
)

type APIFetcher interface {
	Get() (string, error)
}

type FileAPIFetcher struct {
	file string
}

func (f *FileAPIFetcher) Get() (string, error) {
	log.Print("Reading...")
	file, err := os.Open(f.file)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	b.ReadFrom(file)
	return b.String(), nil
}

type WebAPIFetcher struct {
}

func (f *WebAPIFetcher) Get() (string, error) {
	log.Print("Fetching from the API...")
	resp, err := http.Get("https://layer.bicyclesharing.net/map/v1/nyc/stations")
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	b.ReadFrom(resp.Body)
	return b.String(), nil
}

type AngelHandler struct {
	atMost      int
	f           APIFetcher
	m           sync.Mutex
	contents    string
	lastFetched time.Time
}

func (a *AngelHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tact, _ := time.ParseDuration("15m")
	a.m.Lock()
	now := time.Now()
	if now.Truncate(tact).After(a.lastFetched.Truncate(tact)) {
		str, err := a.f.Get()
		if err != nil {
			panic(err)
		}
		var b bytes.Buffer
		angels.Run(a.atMost, strings.NewReader(str), &b)
		a.contents = b.String()
		a.lastFetched = now
	}
	str := a.contents
	a.m.Unlock()

	strings.NewReader(str).WriteTo(w)
}

func main() {
	fileName := flag.String("station-info", "", "")
	atMost := flag.Int("at-most", 10, "")
	local := flag.Bool("local", true, "")
	port := flag.Int("port", 8001, "")
	flag.Parse()
	if *local {
		http.Handle("/", &AngelHandler{atMost: *atMost, f: &FileAPIFetcher{file: *fileName}})
	} else {
		http.Handle("/", &AngelHandler{atMost: *atMost, f: &WebAPIFetcher{}})
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
