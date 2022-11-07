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
		log.Printf("err: %v", err)
		return
	}

	if !db.IsUpdated() {
		log.Printf("service is updating now")

		s := &internal.Resp{Status: http.StatusOK, Message: "The service is updating now. Please try again later."}
		b, err := json.MarshalIndent(s, "", " ")
		if err != nil {
			log.Printf("err: %v", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(res, string(b))
		return
	}

	a := req.URL.Query().Get("ip")

	t := time.Now()

	ip, err = db.Search(a)
	if err != nil {
		log.Printf("err: %v", err)

		error := &internal.Resp{Status: http.StatusNotFound, Message: err.Error()}
		b, err := json.MarshalIndent(error, "", " ")
		if err != nil {
			log.Printf("err: %v", err)
			return
		}

		http.Error(res, string(b), http.StatusNotFound)
		return
	}

	d := time.Since(t)
	log.Printf("%s is %v, elapsed time: %v\n", a, ip, d)

	b, err := json.MarshalIndent(ip, "", " ")
	if err != nil {
		log.Printf("err: %v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(res, string(b))
}
