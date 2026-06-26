package httpapi

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Adapter        string `yaml:"adapter"`
	Listen         string `yaml:"listen"`
	DispatchURL    string `yaml:"dispatch_url"`
	SlotManagerURL string `yaml:"slot_manager_url"`
	UpstreamType   string `yaml:"upstream_type"`
}

func Load(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var c Config
	return c, yaml.Unmarshal(b, &c)
}

// publicInfo is the curated view served from /v1/info. It deliberately omits internal
// topology (dispatch_url, slot_manager_url) so the unauthenticated endpoint does not
// disclose where the adapter forwards traffic.
type publicInfo struct {
	Adapter      string `json:"adapter"`
	Listen       string `json:"listen"`
	UpstreamType string `json:"upstream_type"`
}

// PublicInfo returns the subset of the config safe to expose publicly.
func (c Config) PublicInfo() publicInfo {
	return publicInfo{Adapter: c.Adapter, Listen: c.Listen, UpstreamType: c.UpstreamType}
}

type Server struct{ cfg Config }

func New(cfg Config) *Server { return &Server{cfg: cfg} }

func (s *Server) Addr() string { return s.cfg.Listen }

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("ok")); err != nil {
			log.Printf("httpapi: /healthz write: %v", err)
		}
	})
	mux.HandleFunc("/v1/info", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, "/v1/info", s.cfg.PublicInfo())
	})
	mux.HandleFunc("/v1/chat/completions", s.chatCompletions)
	return mux
}

func (s *Server) chatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("httpapi: /v1/chat/completions read body: %v", err)
		http.Error(w, "cannot read request body", http.StatusBadRequest)
		return
	}
	var req chatRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, "/v1/chat/completions", map[string]any{
		"adapter": s.cfg.Adapter, "stub": true,
		"dispatch_url": s.cfg.DispatchURL,
		"note":         "forward to cofiswarm-dispatch in production",
		"model":        req.Model,
		"messages":     len(req.Messages),
		"bytes":        len(body),
	})
}

// writeJSON encodes v to w, logging any encode error. The response status is already
// committed by the time Encode runs, so logging is the only recourse — but we never
// swallow the failure silently.
func writeJSON(w http.ResponseWriter, route string, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("httpapi: %s encode response: %v", route, err)
	}
}
