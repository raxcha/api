package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

type Server struct {
	cfg Config
	mu  sync.RWMutex
}

func NewServer(cfg Config) *Server {
	return &Server{cfg: cfg}
}

func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/shhh", s.withFrontAuth(s.handleIndex))
	mux.HandleFunc("/page", s.withAuth(s.handlePage))
	return http.ListenAndServe(s.cfg.Addr, mux)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/shhh" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	page, err := readPage(s.cfg.Root, ".", -1)
	s.mu.RUnlock()
	if err != nil {
		http.Error(w, "storage not found", http.StatusInternalServerError)
		return
	}

	body, err := json.MarshalIndent(page, "", "  ")
	if err != nil {
		http.Error(w, "encode error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(body)
}

func (s *Server) handlePage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleFetch(w, r)
	case http.MethodPut:
		s.handlePush(w, r)
	case http.MethodDelete:
		s.handleDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleFetch(w http.ResponseWriter, r *http.Request) {

	path, _ := url.QueryUnescape(r.URL.Query().Get("path"))
	depth, _ := strconv.Atoi(r.URL.Query().Get("depth"))
	if depth == 0 {
		depth = -1
	}

	s.mu.RLock()
	page, err := readPage(s.cfg.Root, path, depth)
	s.mu.RUnlock()

	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	body, err := encodePage(page)
	if err != nil {
		http.Error(w, "encode error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func (s *Server) handlePush(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}

	p, err := decodePage(body)
	if err != nil {
		http.Error(w, "decode error", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	err = writePage(s.cfg.Root, p)
	s.mu.Unlock()

	if err != nil {
		http.Error(w, "write error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {

	path, _ := url.QueryUnescape(r.URL.Query().Get("path"))

	s.mu.Lock()
	err := deletePage(s.cfg.Root, path)
	s.mu.Unlock()

	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.Token != "" && r.Header.Get("Authorization") != "Bearer "+s.cfg.Token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func (s *Server) withFrontAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.FrontPassword == "" {
			http.Error(w, "front page password is not configured", http.StatusServiceUnavailable)
			return
		}

		user, pass, ok := r.BasicAuth()
		if !ok || user != s.cfg.FrontUser || pass != s.cfg.FrontPassword {
			w.Header().Set("WWW-Authenticate", `Basic realm="prsnl.spc"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}
