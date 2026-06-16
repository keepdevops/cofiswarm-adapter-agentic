package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Adapter         string `yaml:"adapter"`
	Listen          string `yaml:"listen"`
	DispatchURL     string `yaml:"dispatch_url"`
	SlotManagerURL  string `yaml:"slot_manager_url"`
	UpstreamType    string `yaml:"upstream_type"`
}

func Load(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var c Config
	return c, yaml.Unmarshal(b, &c)
}

type Server struct{ cfg Config }

func New(cfg Config) *Server { return &Server{cfg: cfg} }

func (s *Server) Addr() string { return s.cfg.Listen }

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/v1/info", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(s.cfg)
	})
	mux.HandleFunc("/v1/chat/completions", s.chatCompletions)
	return mux
}

func (s *Server) chatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	body, _ := io.ReadAll(r.Body)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"adapter": s.cfg.Adapter, "stub": true,
		"dispatch_url": s.cfg.DispatchURL,
		"note":         "forward to cofiswarm-dispatch in production",
		"bytes":        len(body),
	})
}
