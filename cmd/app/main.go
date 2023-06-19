package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	log "github.com/rs/zerolog/log"

	"github.com/ivanglie/iploc/internal/database"
	"github.com/ivanglie/iploc/internal/httputils"
	"github.com/jessevdk/go-flags"
)

var (
	opts struct {
		Token string `long:"token" env:"TOKEN" description:"IP2Location token"`
		Ssl   bool   `long:"ssl" env:"SSL" description:"use ssl"`
		Host  string `long:"host" env:"HOST" description:"hostname"`
		Dbg   bool   `long:"dbg" env:"DEBUG" description:"use debug"`
	}

	db      *database.DB
	version = "unknown"
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

	db = database.NewDB()
	go func() {
		if err := db.Init(opts.Token, "."); err != nil {
			log.Error().Msg(err.Error())
		}
	}()

	h := http.NewServeMux()
	h.HandleFunc("/", index)
	h.HandleFunc("/search", search)

	s := httputils.NewServer(h, opts.Ssl, opts.Host, opts.Dbg)
	log.Info().Msg(s.String())

	log.Info().Msg("Listening...")
	if err := s.ListenAndServe(); err != nil {
		log.Fatal().Msg(err.Error())
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Index")

	a, _, err := httputils.UserIP(r)
	log.Info().Msgf("user ip: %s", a)

	if err != nil {
		log.Error().Msgf("err %v", err)
		fmt.Fprintln(w, err)
		return
	}

	http.Redirect(w, r, "/search?ip="+a, http.StatusTemporaryRedirect)
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
