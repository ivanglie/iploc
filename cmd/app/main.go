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
	log.Println("Searching...")

	a := r.URL.Query().Get("ip")
	log.Println("a=", a)

	loc, _ := iploc.Search(a, s)
	log.Println("loc=", loc)

	log.Println("Search completed")

	fmt.Fprintln(w, loc)
}

func prepare(w http.ResponseWriter, r *http.Request) {
	log.Println("Preparing...")

	csv, err := utils.UnzipCSV(d)
	if err != nil {
		log.Println("err=", err)
		fmt.Fprintln(w, err)
		return
	}
	log.Println("csv=", csv)

	s, _ = utils.SplitCSV(csv.File, csv.Size/200)
	if err != nil {
		log.Println("err=", err)
		fmt.Fprintln(w, err)
		return
	}
	log.Println("s=", s)

	log.Println("Prepare completed")
	fmt.Fprintln(w, csv, s)
}

func download(w http.ResponseWriter, r *http.Request) {
	log.Println("Downloading...")
	// d = "/Users/alexivnv/Documents/code/go/iploc/cmd/app/DB11LITEIPV6.zip"

	decoder := json.NewDecoder(r.Body)

	var ip2location *IP2Location
	err := decoder.Decode(&ip2location)
	if err != nil {
		panic(err)
	}

	token = ip2location.Token
	log.Println("token=", ip2location.Token)

	ch := make(chan string)
	go func(p, t string) {
		d, _, err = utils.Download(".", token)
		ch <- d
	}(".", token)

	log.Println("<-ch=", <-ch)
	log.Println("Download completed")

	// d, _, _ = utils.Download(".", token)

	fmt.Fprintln(w, d)
}
