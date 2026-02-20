package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bluephantom/port0/internal/state"
)

func TestProxyForwardsRequest(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello from backend"))
	}))
	defer backend.Close()

	_, portStr, _ := net.SplitHostPort(backend.Listener.Addr().String())
	port := 0
	fmt.Sscanf(portStr, "%d", &port)

	p := New()
	p.UpdateState(&state.State{
		Projects: map[string]*state.Project{
			"testproject": {Name: "testproject", Port: port},
		},
	})

	req := httptest.NewRequest("GET", "http://testproject.localhost/", nil)
	req.Host = "testproject.localhost"
	w := httptest.NewRecorder()

	p.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if string(body) != "hello from backend" {
		t.Errorf("body = %q, want %q", string(body), "hello from backend")
	}
}

func TestProxyNotFound(t *testing.T) {
	p := New()

	req := httptest.NewRequest("GET", "http://nonexistent.localhost/", nil)
	req.Host = "nonexistent.localhost"
	w := httptest.NewRecorder()

	p.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Errorf("status = %d, want 404", w.Code)
	}

	var result map[string]string
	json.NewDecoder(w.Body).Decode(&result)
	if result["error"] == "" {
		t.Error("expected error message in response")
	}
}

func TestExtractName(t *testing.T) {
	tests := []struct {
		host string
		want string
	}{
		{"myapp.localhost", "myapp"},
		{"myapp.localhost:80", "myapp"},
		{"myapp.web", "myapp"},
		{"myapp.web:8080", "myapp"},
		{"myapp.local", "myapp"},
		{"plain", "plain"},
	}
	for _, tt := range tests {
		got := extractName(tt.host)
		if got != tt.want {
			t.Errorf("extractName(%q) = %q, want %q", tt.host, got, tt.want)
		}
	}
}
