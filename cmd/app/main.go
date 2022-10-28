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

var chunks *internal.Chunks

func init() {
	s, err := internal.ReadChunks("chunks.json")
	if s != nil && err == nil {
		if len(s.Paths) > 0 {
			log.Println("You're up to date")
		}
		chunks = s
		return
	}

	log.Println(err)

	chunks = internal.NewChunks()
	log.Println("Updating...")
	chunks.Update()
	err = internal.WriteChunks(chunks, "chunks.json")
	if err != nil {
		log.Panic(err)
	}
}

func main() {
	http.HandleFunc("/", search)

	port := os.Getenv("PORT")
	if len(port) == 0 {
		log.Fatalf("Incorrect port: %s\n", port)
	}

	log.Printf("Listening on port: %s", port)
	err := http.ListenAndServe(":"+port, nil)

	if errors.Is(err, http.ErrServerClosed) {
		log.Panic("Server closed.")
	} else if err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}

func search(res http.ResponseWriter, req *http.Request) {
	var ip *internal.IP
	var err error

	address := req.URL.Query().Get("ip")
	if len(address) == 0 {
		log.Panic(errors.New("'" + address + "' is incorrect IP"))
	}

	t := time.Now()
	ip, err = internal.Search(address, chunks.Paths...)
	if err != nil {
		log.Printf("err: %v\n", err)
		return
	}

	d := time.Since(t)
	bytes, err := json.MarshalIndent(ip, "", " ")
	if err != nil {
		log.Printf("err: %v\n", err)
		return
	}

	log.Printf("%s is %v\nerr: %v\nelapsed time: %v\n", address, string(bytes), err, d)
	fmt.Fprintln(res, string(bytes))
}
