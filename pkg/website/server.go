// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package website

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

type ServerOpts struct {
	ListenAddr      string
	RedirectToHTTPS bool
	CheckCookie     bool
	TemplateFunc    func([]byte) ([]byte, error)
	ErrorFunc       func(error) ([]byte, error)
}

type Server struct {
	opts ServerOpts
}

func NewServer(opts ServerOpts) *Server {
	return &Server{opts}
}

func (s *Server) Mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.redirectToHTTPS(s.noCacheHandler(s.mainHandler)))
	mux.HandleFunc("/js/", s.redirectToHTTPS(s.noCacheHandler(s.assetHandler)))
	mux.HandleFunc("/examples", s.redirectToHTTPS(s.noCacheHandler(s.corsHandler(s.exampleSetsHandler))))
	mux.HandleFunc("/examples/", s.redirectToHTTPS(s.noCacheHandler(s.corsHandler(s.examplesHandler))))
	// no need for caching as it's a POST
	mux.HandleFunc("/template", s.redirectToHTTPS(s.corsHandler(s.templateHandler)))
	mux.HandleFunc("/alpha-test", s.redirectToHTTPS(s.noCacheHandler(s.alphaTestHandler)))
	mux.HandleFunc("/health", s.healthHandler)
	return mux
}

func (s *Server) Run() error {
	server := &http.Server{
		Addr:    s.opts.ListenAddr,
		Handler: s.Mux(),
	}
	fmt.Printf("Listening on http://%s\n", server.Addr)
	return server.ListenAndServe()
}

func (s *Server) mainHandler(w http.ResponseWriter, r *http.Request) {
	s.write(w, []byte(Files["templates/index.html"].Content))
}

func (s *Server) assetHandler(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, ".css") {
		w.Header().Set("Content-Type", "text/css")
	}
	if strings.HasSuffix(r.URL.Path, ".js") {
		w.Header().Set("Content-Type", "application/javascript")
	}
	s.write(w, []byte(Files[strings.TrimPrefix(r.URL.Path, "/")].Content))
}

func (s *Server) exampleSetsHandler(w http.ResponseWriter, r *http.Request) {
	slimexampleSets := []exampleSet{}

	for _, eg := range exampleSets {
		examples := []Example{}
		for _, example := range eg.Examples {
			examples = append(examples, Example{
				ID:          example.ID,
				DisplayName: example.DisplayName,
			})
		}

		slimexampleSets = append(slimexampleSets, exampleSet{
			ID:          eg.ID,
			DisplayName: eg.DisplayName,
			Description: eg.Description,
			Examples:    examples,
		})
	}

	listBytes, err := json.Marshal(slimexampleSets)
	if err != nil {
		s.logError(w, err)
		return
	}

	s.write(w, listBytes)
}

func (s *Server) examplesHandler(w http.ResponseWriter, r *http.Request) {
	for _, eg := range exampleSets {
		for _, example := range eg.Examples {
			if example.ID == strings.TrimPrefix(r.URL.Path, "/examples/") {
				exampleBytes, err := json.Marshal(example)
				if err != nil {
					s.logError(w, err)
					return
				}

				s.write(w, exampleBytes)
				return
			}
		}
	}

	s.logError(w, fmt.Errorf("Did not find example: %v", strings.TrimPrefix(r.URL.Path, "/examples/")))
}

func (s *Server) templateHandler(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		s.logError(w, err)
		return
	}

	resp, err := s.opts.TemplateFunc(data)
	if err != nil {
		s.logError(w, err)
		return
	}

	s.write(w, resp)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	s.write(w, []byte("ok"))
}

func (s *Server) logError(w http.ResponseWriter, err error) {
	log.Print(err.Error())

	resp, err := s.opts.ErrorFunc(err)
	if err != nil {
		fmt.Fprintf(w, "generation error: %s", err.Error())
		return
	}

	s.write(w, resp)
}

func (s *Server) write(w http.ResponseWriter, data []byte) {
	w.Write(data) // not fmt.Fprintf!
}

func (s *Server) redirectToHTTPS(wrappedFunc func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	if !s.opts.RedirectToHTTPS {
		return wrappedFunc
	}
	return func(w http.ResponseWriter, r *http.Request) {
		checkHTTPS := true
		clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			if clientIP == "127.0.0.1" {
				checkHTTPS = false
			}
		}

		if checkHTTPS && r.Header.Get(http.CanonicalHeaderKey("x-forwarded-proto")) != "https" {
			if r.Method == http.MethodGet || r.Method == http.MethodHead {
				host := r.Header.Get("host")
				if len(host) == 0 {
					s.logError(w, fmt.Errorf("expected non-empty Host header"))
					return
				}

				http.Redirect(w, r, "https://"+host, http.StatusMovedPermanently)
				return
			}

			// Fail if it's not a GET or HEAD since req may have carried body insecurely
			s.logError(w, fmt.Errorf("expected HTTPs connection"))
			return
		}

		wrappedFunc(w, r)
	}
}

var (
	noCacheHeaders = map[string]string{
		"Expires":         time.Unix(0, 0).Format(time.RFC1123),
		"Cache-Control":   "no-cache, private, max-age=0",
		"Pragma":          "no-cache",
		"X-Accel-Expires": "0",
	}
)

func (s *Server) noCacheHandler(wrappedFunc func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		for k, v := range noCacheHeaders {
			w.Header().Set(k, v)
		}

		wrappedFunc(w, r)
	}
}

func (s *Server) corsHandler(wrappedFunc func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		wrappedFunc(w, r)
	}
}

func (s *Server) alphaTestHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
