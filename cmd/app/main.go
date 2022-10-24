package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

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
	switch ver := strings.Count(address, ":"); {
	case ver < 2:
		ip, err = internal.Search(address, state.IPv4Chunks...)
	case ver >= 2:
		ip, err = internal.Search(address, state.IPv6Chunks...)
	default:
		fmt.Fprintf(res, "'%s' is incorrect IP", address)
	}

	log.Printf("%s is %v (err: %v, elapsed time: %v)\n", address, ip, err, time.Since(t))
	fmt.Fprintf(res, "%s is %v (err: %v, elapsed time: %v)", address, ip, err, time.Since(t))
}
