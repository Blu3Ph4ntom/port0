package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blu3ph4ntom/port0/internal/state"
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

func TestProxySubdomainRouting(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello from api"))
	}))
	defer backend.Close()

	_, portStr, _ := net.SplitHostPort(backend.Listener.Addr().String())
	port := 0
	fmt.Sscanf(portStr, "%d", &port)

	p := New()
	p.UpdateState(&state.State{
		Projects: map[string]*state.Project{
			"api": {Name: "api", Port: port, Domain: "myapp"},
			"web": {Name: "web", Port: 9999, Domain: "myapp"},
		},
	})

	// Request to api.myapp.localhost should route to "api" project
	req := httptest.NewRequest("GET", "http://api.myapp.localhost/", nil)
	req.Host = "api.myapp.localhost"
	w := httptest.NewRecorder()

	p.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if string(body) != "hello from api" {
		t.Errorf("body = %q, want %q", string(body), "hello from api")
	}
}

func TestProxySubdomainFallbackToNoDomain(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello from myapp"))
	}))
	defer backend.Close()

	_, portStr, _ := net.SplitHostPort(backend.Listener.Addr().String())
	port := 0
	fmt.Sscanf(portStr, "%d", &port)

	p := New()
	p.UpdateState(&state.State{
		Projects: map[string]*state.Project{
			"myapp": {Name: "myapp", Port: port, Domain: ""},
			"api":   {Name: "api", Port: 9999, Domain: "other"},
		},
	})

	// Request to myapp.localhost should route to "myapp" project (no domain)
	req := httptest.NewRequest("GET", "http://myapp.localhost/", nil)
	req.Host = "myapp.localhost"
	w := httptest.NewRecorder()

	p.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if string(body) != "hello from myapp" {
		t.Errorf("body = %q, want %q", string(body), "hello from myapp")
	}
}

func TestExtractNameAndDomain(t *testing.T) {
	tests := []struct {
		host       string
		wantName   string
		wantDomain string
	}{
		// Basic domains (no subdomain)
		{"myapp.localhost", "myapp", ""},
		{"myapp.localhost:80", "myapp", ""},
		{"myapp.web", "myapp", ""},
		{"myapp.web:8080", "myapp", ""},
		{"myapp.local", "myapp", ""},
		{"plain", "plain", ""},
		// Subdomains with domain - extracts name and domain separately
		{"api.myapp.localhost", "api", "myapp"},
		{"admin.myapp.localhost:80", "admin", "myapp"},
		{"api.myapp.web", "api", "myapp"},
		{"dashboard.myapp.web:8080", "dashboard", "myapp"},
		{"api.myapp.local", "api", "myapp"},
		// Multiple subdomains - first is name, rest is domain
		{"v1.api.myapp.localhost", "v1", "api.myapp"},
		{"app.web.myapp.web", "app", "web.myapp"},
		// Case insensitive
		{"API.MyApp.Localhost", "api", "myapp"},
		{"MYAPP.LOCALHOST", "myapp", ""},
	}
	for _, tt := range tests {
		gotName, gotDomain := extractNameAndDomain(tt.host)
		if gotName != tt.wantName || gotDomain != tt.wantDomain {
			t.Errorf("extractNameAndDomain(%q) = (%q, %q), want (%q, %q)", tt.host, gotName, gotDomain, tt.wantName, tt.wantDomain)
		}
	}
}
