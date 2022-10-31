package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ivanglie/iploc/internal"
)

var db *internal.DB

func init() {
	log.Println("Updating...")

	db = internal.NewDB()
	db.Update()
}

func main() {
	http.HandleFunc("/", search)

	port := os.Getenv("PORT")
	if len(port) == 0 {
		log.Fatalf("incorrect port: %s\n", port)
	}

	log.Printf("Listening on port: %s", port)
	err := http.ListenAndServe(":"+port, nil)

	if errors.Is(err, http.ErrServerClosed) {
		log.Panic("server closed")
	} else if err != nil {
		log.Fatalf("error starting server: %s", err)
	}
}

func search(res http.ResponseWriter, req *http.Request) {
	var ip *internal.IP
	var err error

	if db == nil {
		err = errors.New("db is empty")
		log.Printf("err: %v\n", err)
		return
	}

	a := req.URL.Query().Get("ip")

	t := time.Now()

	ip, err = db.Search(a)
	if err != nil {
		log.Printf("err: %v\n", err)
		return
	}

	d := time.Since(t)
	log.Printf("%s is %v, elapsed time: %v\n", a, ip, d)

	b, err := json.MarshalIndent(ip, "", " ")
	if err != nil {
		log.Printf("err: %v\n", err)
		return
	}

	fmt.Fprintln(res, string(b))
}
