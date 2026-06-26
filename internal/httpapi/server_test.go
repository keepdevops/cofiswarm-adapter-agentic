package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testConfig() Config {
	return Config{
		Adapter:        "adapter-agentic",
		Listen:         ":8032",
		DispatchURL:    "http://127.0.0.1:8010",
		SlotManagerURL: "http://127.0.0.1:8013",
		UpstreamType:   "agentic",
	}
}

func TestLoadValid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "adapter.yaml")
	yaml := "adapter: adapter-agentic\nlisten: \":8032\"\n" +
		"dispatch_url: http://127.0.0.1:8010\nslot_manager_url: http://127.0.0.1:8013\n" +
		"upstream_type: agentic\n"
	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg != testConfig() {
		t.Fatalf("got %+v, want %+v", cfg, testConfig())
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "nope.yaml")); err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadMalformedYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.yaml")
	if err := os.WriteFile(path, []byte("adapter: [unterminated\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected error for malformed yaml, got nil")
	}
}

func TestAddr(t *testing.T) {
	if got := New(testConfig()).Addr(); got != ":8032" {
		t.Fatalf("Addr() = %q, want %q", got, ":8032")
	}
}

func TestHealthz(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	New(testConfig()).Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if body := rr.Body.String(); body != "ok" {
		t.Fatalf("body = %q, want %q", body, "ok")
	}
}

func TestInfoReturnsPublicView(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/info", nil)
	New(testConfig()).Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	var got map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got["adapter"] != "adapter-agentic" || got["upstream_type"] != "agentic" || got["listen"] != ":8032" {
		t.Fatalf("unexpected public view: %+v", got)
	}
	// Internal topology must not leak from an unauthenticated endpoint.
	for _, leaked := range []string{"dispatch_url", "slot_manager_url"} {
		if _, ok := got[leaked]; ok {
			t.Fatalf("%s leaked in /v1/info: %+v", leaked, got)
		}
	}
}

func TestChatCompletionsPost(t *testing.T) {
	payload := `{"model":"gpt-x","messages":[{"role":"user","content":"hi"}]}`
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(payload))
	New(testConfig()).Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	var got map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got["adapter"] != "adapter-agentic" || got["stub"] != true {
		t.Fatalf("unexpected body: %+v", got)
	}
	if got["model"] != "gpt-x" || got["messages"].(float64) != 1 {
		t.Fatalf("echoed request mismatch: %+v", got)
	}
	if got["bytes"].(float64) != float64(len(payload)) {
		t.Fatalf("bytes = %v, want %d", got["bytes"], len(payload))
	}
}

func TestChatCompletionsValidation(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"invalid json", `{not json`},
		{"missing model", `{"messages":[{"role":"user","content":"hi"}]}`},
		{"no messages", `{"model":"gpt-x","messages":[]}`},
		{"bad role", `{"model":"gpt-x","messages":[{"role":"wizard","content":"hi"}]}`},
		{"empty content", `{"model":"gpt-x","messages":[{"role":"user","content":""}]}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(tc.body))
			New(testConfig()).Handler().ServeHTTP(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d (body: %s)", rr.Code, http.StatusBadRequest, rr.Body.String())
			}
		})
	}
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("boom") }

func TestChatCompletionsBodyReadError(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", errReader{})
	New(testConfig()).Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestChatCompletionsRejectsNonPost(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/chat/completions", nil)
	New(testConfig()).Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}
