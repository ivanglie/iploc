package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	log "github.com/rs/zerolog/log"

	"github.com/ivanglie/iploc/internal/iploc"
	"github.com/ivanglie/iploc/internal/utils"
	"github.com/jessevdk/go-flags"
)

type IP2Location struct {
	Token string `json:"token"`
}

var (
	opts struct {
		Port string `long:"p" env:"PORT" default:"18001" description:"port"`
		Dbg  bool   `long:"dbg" env:"DEBUG" description:"debug mode"`
	}

	version = "unknown"

	token string
	d     string
	csv   *utils.CSV
	s     []string
)

func main() {
	fmt.Printf("iploc %s\n", version)

	p := flags.NewParser(&opts, flags.PrintErrors|flags.PassDoubleDash|flags.HelpFlag)
	if _, err := p.Parse(); err != nil {
		if err.(*flags.Error).Type != flags.ErrHelp {
			log.Printf("[ERROR] iploc error: %v", err)
		}
		os.Exit(2)
	}

	setupLog(false)

	port := opts.Port
	if len(port) == 0 {
		log.Fatal().Msgf("incorrect port: %s\n", port)
	}

	http.HandleFunc("/", index)
	http.HandleFunc("/search", search)
	http.HandleFunc("/split", split)
	http.HandleFunc("/unzip", unzip)
	http.HandleFunc("/download", download)

	log.Info().Msgf("Listening on port: %s", port)
	err := http.ListenAndServe(":"+port, nil)

	if errors.Is(err, http.ErrServerClosed) {
		log.Panic().Msg("server closed")
	} else if err != nil {
		log.Fatal().Msgf("error starting server: %s", err)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Index")

	a, _, err := utils.UserIP(r)
	log.Info().Msgf("user ip: %s", a)

	if err != nil {
		log.Error().Msgf("err %v", err)
		fmt.Fprintln(w, err)
		return
	}

	http.Redirect(w, r, "/search?ip="+a, http.StatusSeeOther)
}

func search(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Search...")

	a := r.URL.Query().Get("ip")
	log.Info().Msgf("user ip: %s", a)

	loc, err := iploc.Search(a, s)
	if err != nil {
		log.Error().Msgf("err %v", err)
		fmt.Fprintln(w, err)
		return
	}

	log.Debug().Msgf("loc: %v", loc)
	log.Info().Msg("Search completed")
	fmt.Fprintln(w, loc)
}

func split(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Split...")

	var err error
	s, err = utils.SplitCSV(csv.File, csv.Size/200)
	if err != nil {
		log.Error().Msgf("err %v", err)
		fmt.Fprintln(w, err)
		return
	}

	log.Debug().Msgf("s %s", s)
	log.Info().Msg("Split completed")
	fmt.Fprintln(w, s)
}

func unzip(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Unzip...")

	// debug
	if len(d) == 0 {
		d = "/tmp/DB11LITEIPV6.zip"
	}

	var err error
	csv, err = utils.Unzip(d)
	if err != nil {
		log.Error().Msgf("err %v", err)
		fmt.Fprintln(w, err)
		return
	}

	log.Debug().Msgf("csv %v", csv)
	log.Info().Msg("Unzip completed")
	fmt.Fprintln(w, csv)
}

func download(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Download...")

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

	log.Debug().Msgf("file: %v, err: %v", <-fileCh, <-errCh)
	log.Info().Msg("Download completed")
	fmt.Fprintln(w, d)
}

func setupLog(dbg bool) {
	if dbg {
		log.Level(zerolog.DebugLevel)
		return
	}

	log.Level(zerolog.InfoLevel)
}
