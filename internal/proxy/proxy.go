package proxy

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/blu3ph4ntom/port0/internal/cert"
	"github.com/blu3ph4ntom/port0/internal/state"
)

type Proxy struct {
	state      atomic.Pointer[state.State]
	httpServer *http.Server
	tlsServer  *http.Server
}

func New() *Proxy {
	p := &Proxy{}
	empty := &state.State{Projects: make(map[string]*state.Project)}
	p.state.Store(empty)
	return p
}

func (p *Proxy) UpdateState(st *state.State) {
	p.state.Store(st)
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name, domain := extractNameAndDomain(r.Host)
	if name == "" {
		writeError(w, http.StatusBadRequest, "missing host header")
		return
	}

	st := p.state.Load()

	// Try to find project by name + domain (subdomain routing)
	// e.g., host="api.myapp.localhost" -> name="api", domain="myapp"
	// looks for project "api" with Domain="myapp"
	var proj *state.Project
	for _, p := range st.Projects {
		if p.Name == name && p.Domain == domain {
			proj = p
			break
		}
	}

	// Fallback: look for project matching just the name (no domain/subdomain)
	if proj == nil {
		if p, ok := st.Projects[name]; ok && p.Domain == "" {
			proj = p
		}
	}

	if proj == nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("no project named %q", name))
		return
	}

	if isWebSocket(r) {
		p.tunnelWebSocket(w, r, proj.Port)
		return
	}

	target, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", proj.Port))
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("invalid target URL: %v", err))
		return
	}
	rp := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.Host = r.Host
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			slog.Error("proxy error", "project", name, "err", err)
			writeError(w, http.StatusBadGateway, fmt.Sprintf("upstream error: %v", err))
		},
	}
	rp.ServeHTTP(w, r)
}

func (p *Proxy) StartHTTP(addr string) error {
	p.httpServer = &http.Server{
		Addr:         addr,
		Handler:      p,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("proxy: listen http %s: %w", addr, err)
	}
	go p.httpServer.Serve(ln)
	return nil
}

func (p *Proxy) StartTLS(addr string) error {
	p.tlsServer = &http.Server{
		Addr:    addr,
		Handler: p,
		TLSConfig: &tls.Config{
			GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
				name, _ := extractNameAndDomain(hello.ServerName)
				c, err := cert.Load(name)
				if err != nil {
					return nil, fmt.Errorf("proxy: load cert for %s: %w", name, err)
				}
				return &c, nil
			},
		},
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	ln, err := tls.Listen("tcp", addr, p.tlsServer.TLSConfig)
	if err != nil {
		return fmt.Errorf("proxy: listen tls %s: %w", addr, err)
	}
	go p.tlsServer.Serve(ln)
	return nil
}

func (p *Proxy) Stop() error {
	var firstErr error
	if p.httpServer != nil {
		if err := p.httpServer.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if p.tlsServer != nil {
		if err := p.tlsServer.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (p *Proxy) tunnelWebSocket(w http.ResponseWriter, r *http.Request, port int) {
	upstream, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 5*time.Second)
	if err != nil {
		writeError(w, http.StatusBadGateway, fmt.Sprintf("cannot reach upstream: %v", err))
		return
	}
	defer upstream.Close()

	hj, ok := w.(http.Hijacker)
	if !ok {
		writeError(w, http.StatusInternalServerError, "hijack not supported")
		return
	}
	client, _, err := hj.Hijack()
	if err != nil {
		slog.Error("proxy: hijack failed", "err", err)
		return
	}
	defer client.Close()

	if err := r.Write(upstream); err != nil {
		slog.Error("proxy: write upgrade request", "err", err)
		return
	}

	done := make(chan struct{}, 2)
	copyFunc := func(dst, src net.Conn) {
		io.Copy(dst, src)
		done <- struct{}{}
	}
	go copyFunc(upstream, client)
	go copyFunc(client, upstream)
	<-done
}

func extractNameAndDomain(host string) (name, domain string) {
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}
	host = strings.ToLower(host)

	// Strip known TLD suffixes to get the routing domain
	// Supports: *.localhost, *.web, *.local, and naked names
	for _, suffix := range []string{".localhost", ".local", ".web"} {
		if strings.HasSuffix(host, suffix) {
			host = host[:len(host)-len(suffix)]
			break
		}
	}

	// Parse subdomain.domain format
	// e.g., "api.myapp" -> name="api", domain="myapp"
	// e.g., "myapp" -> name="myapp", domain=""
	if idx := strings.Index(host, "."); idx != -1 {
		return host[:idx], host[idx+1:]
	}
	return host, ""
}

func isWebSocket(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket")
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": msg}); err != nil {
		slog.Error("failed to encode error response", "err", err)
	}
}
