package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"github.com/ivanglie/iploc/internal/service"
	"github.com/ivanglie/iploc/pkg/log"
	"github.com/ivanglie/iploc/pkg/netutil"
)

type httpServer interface {
	ListenAndServe() error
}

type Server struct {
	service  *service.Service
	listener httpServer
}

// New creates a new Server.
func New(addr string) *Server {
	s := &Server{
		service: service.New(),
		listener: &http.Server{
			Addr: addr,
		},
	}

	m := http.NewServeMux()
	m.HandleFunc("/", s.index)
	m.HandleFunc("/search", s.search)

	s.listener.(*http.Server).Handler = m

	return s
}

// Start starts the server.
func (s *Server) Start(local bool, token, path string) error {
	if err := s.service.Prepare(local, token, path); err != nil {
		return err
	}

	return s.listener.ListenAndServe()
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	log.Info("Index")

	a, _, err := netutil.UserIP(r)
	log.Info(fmt.Sprintf("user ip: %s", a))

	if err != nil {
		log.Error(err.Error())
		fmt.Fprintln(w, err)
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

func (s *Server) search(w http.ResponseWriter, r *http.Request) {
	log.Info("Search...")

	a := r.URL.Query().Get("ip")
	log.Info(fmt.Sprintf("user ip: %s", a))

	loc, err := s.service.Search(a)
	if err != nil {
		log.Error(err.Error())
		fmt.Fprintln(w, err)
	}

	log.Debug(fmt.Sprintf("loc: %v", loc))
	log.Info("Search completed")

	if acceptHeader := r.Header.Get("Accept"); strings.Contains(acceptHeader, "application/json") {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, loc)

		return
	}

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(loc.String()), "", "\t"); err != nil {
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
	t.Execute(w, prettyJSON.String())
}
