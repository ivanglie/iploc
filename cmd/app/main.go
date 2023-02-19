package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ivanglie/iploc/internal/iploc"
	"github.com/ivanglie/iploc/internal/utils"
)

var (
	port  string
	token string
	d     string
	csv   *utils.CSV
	s     []string
)

type IP2Location struct {
	Token string `json:"token"`
}

func init() {
	port = os.Getenv("PORT")
	if len(port) == 0 {
		log.Fatalf("incorrect port: %s\n", port)
	}

	token = os.Getenv("IP2LOCATION_TOKEN")
	if len(port) == 0 {
		log.Fatalf("incorrect token: %s\n", token)
	}
}

func main() {
	http.HandleFunc("/search", search)
	http.HandleFunc("/prepare", prepare)
	http.HandleFunc("/download", download)

	log.Printf("Listening on port: %s", port)
	err := http.ListenAndServe(":"+port, nil)

	if errors.Is(err, http.ErrServerClosed) {
		log.Panic("server closed")
	} else if err != nil {
		log.Fatalf("error starting server: %s", err)
	}
}

func search(w http.ResponseWriter, r *http.Request) {
	a := r.URL.Query().Get("ip")
	loc, _ := iploc.Search(a, s)
	fmt.Fprintln(w, loc)
}

func prepare(w http.ResponseWriter, r *http.Request) {
	csv, _ = utils.UnzipCSV(d)
	s, _ = utils.SplitCSV(csv.File, csv.Size/200)
	fmt.Fprintln(w, csv, s)
}

func download(w http.ResponseWriter, r *http.Request) {
	// d = "/Users/alexivnv/Documents/code/go/iploc/cmd/app/DB11LITEIPV6.zip"

	decoder := json.NewDecoder(r.Body)

	var ip2location *IP2Location
	err := decoder.Decode(&ip2location)
	if err != nil {
		panic(err)
	}

	token = ip2location.Token
	log.Println(ip2location.Token)

	d, _, _ = utils.Download(".", token)
	fmt.Fprintln(w, d)
}
