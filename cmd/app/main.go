package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	log "github.com/rs/zerolog/log"

	"github.com/ivanglie/iploc/internal/database"
	"github.com/ivanglie/iploc/internal/utils"
	"github.com/jessevdk/go-flags"
)

type IP2Location struct {
	Token string `json:"token"`
}

var (
	opts struct {
		Token string `long:"token" env:"TOKEN" description:"IP2Location token"`
		Dbg   bool   `long:"dbg" env:"DEBUG" description:"debug mode"`
	}

	version = "unknown"
	db      = &database.DB{}
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

	setupLog(opts.Dbg)

	go func() {
		log.Info().Msg("Prepare...")
		db.Prepare(opts.Token, ".")
		log.Info().Msg("Prepare completed")
	}()

	listenAndServe()
}

func listenAndServe() {
	handler := http.NewServeMux()
	handler.HandleFunc("/", index)
	handler.HandleFunc("/search", search)

	httpServer := &http.Server{
		Addr:    ":18001",
		Handler: handler,
	}

	log.Info().Msgf("Listening on port %s", httpServer.Addr)
	err := httpServer.ListenAndServe()

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

	loc, err := db.Search(a)
	if err != nil {
		log.Error().Msgf("err %v", err)
		fmt.Fprintln(w, err)
		return
	}

	log.Debug().Msgf("loc: %v", loc)
	log.Info().Msg("Search completed")
	fmt.Fprintln(w, loc)
}

func setupLog(dbg bool) {
	if dbg {
		log.Level(zerolog.DebugLevel)
		return
	}

	log.Level(zerolog.InfoLevel)
}
