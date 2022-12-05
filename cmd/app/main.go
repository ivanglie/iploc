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

var (
	port string
	db   *internal.DB
)

func init() {
	port = os.Getenv("PORT")
	if len(port) == 0 {
		log.Fatalf("incorrect port: %s\n", port)
	}

	db = internal.NewDB()
}

func main() {
	http.HandleFunc("/search", search)
	http.HandleFunc("/update", update)

	log.Printf("Listening on port: %s", port)
	err := http.ListenAndServe(":"+port, nil)

	if errors.Is(err, http.ErrServerClosed) {
		log.Panic("server closed")
	} else if err != nil {
		log.Fatalf("error starting server: %s", err)
	}
}

func search(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	var ip *internal.IP
	var err error

	if db == nil {
		err = errors.New("db is empty")
		log.Printf("err: %v", err)
		return
	}

	if !db.IsUpdated() {
		if db.IsUpdating() {
			log.Printf("service is updating now")

			s := &internal.Resp{Status: http.StatusOK, Message: "The service is updating now. Please try again later."}
			b, err := json.MarshalIndent(s, "", " ")
			if err != nil {
				log.Printf("err: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			fmt.Fprintln(w, string(b))
			return
		} else {
			log.Printf("service is unavailable")

			s := &internal.Resp{Status: http.StatusServiceUnavailable, Message: "The service is unavailable. Please update service."}
			b, err := json.MarshalIndent(s, "", " ")
			if err != nil {
				log.Printf("err: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			http.Error(w, string(b), http.StatusServiceUnavailable)
			return
		}
	}

	a := r.URL.Query().Get("ip")

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

		http.Error(w, string(b), http.StatusNotFound)
		return
	}

	d := time.Since(t)
	log.Printf("%s is %v, elapsed time: %v\n", a, ip, d)

	b, err := json.MarshalIndent(ip, "", " ")
	if err != nil {
		log.Printf("err: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, string(b))
}

func update(w http.ResponseWriter, r *http.Request) {
	if db == nil {
		err := errors.New("db is empty")
		log.Printf("err: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if db.IsUpdating() {
		err := errors.New("service is updating now")
		log.Printf("err: %v", err)

		s := &internal.Resp{Status: http.StatusOK, Message: "The service is updating now. Please wait."}
		b, err := json.MarshalIndent(s, "", " ")
		if err != nil {
			log.Printf("err: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Error(w, string(b), http.StatusInternalServerError)
		return
	}

	log.Println("updating")
	fmt.Fprintln(w, "Updating...")

	db.Update()
}
