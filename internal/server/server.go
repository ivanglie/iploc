package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ivanglie/iploc/internal/provider"
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
	m.HandleFunc("/", s.handleIndex)
	m.HandleFunc("/search", s.handleSearch)

	s.listener.(*http.Server).Handler = m

	return s
}

// Start starts the server.
func (s *Server) Start(local bool, token, path string) error {
	if local {
		s.service.SetBufferDivisor(2)
		log.Info("Copy...")
		s.service.SetZIP(filepath.Join(path, provider.DefaultCode))
		if err := service.Copy(filepath.Join(provider.ZipPath, provider.DefaultCode+".zip"), s.service.ZIP()); err != nil {
			return fmt.Errorf("copying: %v", err)
		}
		log.Info("Copying completed")
	} else {
		log.Info("Download...")
		if err := s.service.Download(token, path); err != nil {
			return fmt.Errorf("downloading: %v", err)
		}
		log.Info("Download completed")
	}

	log.Info("Unzip...")
	if err := s.service.Unzip(); err != nil {
		return err
	}
	log.Info("Unzip completed")

	log.Info("Split...")
	err := s.service.Split()
	if err != nil {
		return fmt.Errorf("splitting: %v", err)
	}

	log.Info("Split completed")

	return s.listener.ListenAndServe()
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
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
