package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/ivanglie/iploc/pkg/log"
	"github.com/rs/zerolog"

	"github.com/ivanglie/iploc/internal/database"
	"github.com/ivanglie/iploc/internal/httputils"
	"github.com/jessevdk/go-flags"
)

var (
	opts struct {
		Token string `long:"token" env:"TOKEN" description:"IP2Location token"`
		Ssl   bool   `long:"ssl" env:"SSL" description:"Use ssl"`
		Host  string `long:"host" env:"HOST" description:"Hostname"`
		Dbg   bool   `long:"dbg" env:"DEBUG" description:"Use debug"`
	}

	db      *database.DB
	version = "unknown"
)

func main() {
	fmt.Printf("iploc %s\n", version)

	p := flags.NewParser(&opts, flags.PrintErrors|flags.PassDoubleDash|flags.HelpFlag)
	if _, err := p.Parse(); err != nil {
		if err.(*flags.Error).Type != flags.ErrHelp {
			fmt.Printf("[ERROR] iploc error: %v", err)
		}
		os.Exit(2)
	}

	if opts.Dbg {
		log.SetLogConfig(zerolog.DebugLevel, os.Stdout)
	}

	db = database.NewDB()
	go func() {
		if err := db.Init(opts.Token, "."); err != nil {
			log.Error(err.Error())
		}
	}()

	h := http.NewServeMux()
	h.HandleFunc("/", index)
	h.HandleFunc("/search", search)

	s := httputils.NewServer(h, opts.Ssl, opts.Host, opts.Dbg)
	log.Info(s.String())

	log.Info("Listening...")
	if err := s.ListenAndServe(); err != nil {
		log.Error(err.Error())
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	log.Info("Index")

	a, _, err := httputils.UserIP(r)
	log.Info(fmt.Sprintf("user ip: %s", a))

	if err != nil {
		log.Error(err.Error())
		fmt.Fprintln(w, err)
		return
	}

	t, err := template.ParseFiles("web/template/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{"IP": a}
	if err = t.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func search(w http.ResponseWriter, r *http.Request) {
	log.Info("Search...")

	a := r.URL.Query().Get("ip")
	log.Info(fmt.Sprintf("user ip: %s", a))

	loc, err := db.Search(a)
	if err != nil {
		log.Error(err.Error())
		fmt.Fprintln(w, err)
		return
	}

	log.Debug(fmt.Sprintf("loc: %v", loc))
	log.Info("Search completed")

	if acceptHeader := r.Header.Get("Accept"); strings.Contains(acceptHeader, "application/json") {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(loc)

		return
	}

	formattedLoc, err := json.MarshalIndent(loc, "", "    ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t, err := template.New("result").Parse(`
            <pre>{{.}}</pre>
        `)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, string(formattedLoc))
}
