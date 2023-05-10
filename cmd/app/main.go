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
	token string
	d     string
	csv   *utils.CSV
	s     []string
)

type IP2Location struct {
	Token string `json:"token"`
}

func main() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		log.Fatalf("incorrect port: %s\n", port)
	}

	http.HandleFunc("/", search)
	http.HandleFunc("/search", search)
	http.HandleFunc("/split", split)
	http.HandleFunc("/unzip", unzip)
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
	log.Println("Search...")

	a := r.URL.Query().Get("ip")
	log.Println("a=", a)

	var loc *iploc.Loc
	var err error

	if len(a) == 0 {
		a, _, err = utils.UserIP(r)
		if err != nil {
			log.Println("err=", err)
			fmt.Fprintln(w, err)
			return
		}

		loc, err = iploc.Search(a, s)
	} else {
		loc, err = iploc.Search(a, s)
	}

	if err != nil {
		log.Println("err=", err)
		fmt.Fprintln(w, err)
		return
	}

	log.Println("loc=", loc)
	log.Println("Search completed")
	fmt.Fprintln(w, loc)
}

func split(w http.ResponseWriter, r *http.Request) {
	log.Println("Split...")

	var err error
	s, err = utils.SplitCSV(csv.File, csv.Size/200)
	if err != nil {
		log.Println("err=", err)
		fmt.Fprintln(w, err)
		return
	}

	log.Println("s=", s)
	log.Println("Split completed")
	fmt.Fprintln(w, s)
}

func unzip(w http.ResponseWriter, r *http.Request) {
	log.Println("Unzip...")

	// debug
	if len(d) == 0 {
		d = "/tmp/DB11LITEIPV6.zip"
	}

	var err error
	csv, err = utils.Unzip(d)
	if err != nil {
		log.Println("err=", err)
		fmt.Fprintln(w, err)
		return
	}

	log.Println(csv)
	log.Println("Unzip completed")
	fmt.Fprintln(w, csv)
}

func download(w http.ResponseWriter, r *http.Request) {
	log.Println("Download...")

	// debug
	if len(d) == 0 {
		d = "/tmp/DB11LITEIPV6.zip"
	}

	decoder := json.NewDecoder(r.Body)

	var ip2location *IP2Location
	err := decoder.Decode(&ip2location)
	if err != nil {
		panic(err)
	}

	token = ip2location.Token

	fileCh := make(chan string)
	errCh := make(chan error)
	go func(p, t string) {
		d, _, err = utils.Download(token, ".")
		fileCh <- d
		errCh <- err
	}(".", token)

	log.Println("file=", <-fileCh, "err=", <-errCh)
	log.Println("Download completed")
	fmt.Fprintln(w, d)
}
