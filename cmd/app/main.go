package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ivanglie/iploc/internal"
)

var state *internal.State

func init() {
	s, err := internal.ReadState("state.json")
	if s != nil && err == nil {
		if len(s.IPv4Chunks) > 0 && len(s.IPv6Chunks) > 0 {
			log.Println("You're up to date")
		}
		state = s
		return
	}

	log.Println(err)

	state = internal.NewState()
	log.Println("Updating...")
	state.Update()
	err = internal.WriteState(state, "state.json")
	if err != nil {
		log.Panic(err)
	}
}

func main() {
	http.HandleFunc("/find", find)

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

func find(res http.ResponseWriter, req *http.Request) {
	log.Println(state)

	var item *internal.Item
	var err error

	address := req.URL.Query().Get("ip")
	if len(address) == 0 {
		log.Panic(errors.New(address + " is incorrect IP."))
	}

	switch ver := strings.Count(address, ":"); {
	case ver < 2:
		item, err = internal.Find(address, state.IPv4Chunks...)
	case ver >= 2:
		item, err = internal.Find(address, state.IPv6Chunks...)
	default:
		log.Panic(errors.New(address + " is incorrect IP."))
	}

	fmt.Fprintf(res, "Find page\n%s is %v (err: %v)", address, item, err)
}
