package dns

import (
	"fmt"
	"log/slog"
	"net"
	"strings"
	"sync"

	mdns "github.com/miekg/dns"
)

type Server struct {
	udp  *mdns.Server
	tcp  *mdns.Server
	addr string
	mu   sync.Mutex
}

func New() *Server {
	return &Server{}
}

func (s *Server) Start(addr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	handler := mdns.HandlerFunc(s.handleDNS)

	s.addr = addr
	s.udp = &mdns.Server{
		Addr:    addr,
		Net:     "udp",
		Handler: handler,
	}
	s.tcp = &mdns.Server{
		Addr:    addr,
		Net:     "tcp",
		Handler: handler,
	}

	errCh := make(chan error, 2)
	go func() { errCh <- s.udp.ListenAndServe() }()
	go func() { errCh <- s.tcp.ListenAndServe() }()

	return nil
}

func (s *Server) StartWithFallback(preferred, fallback string) (string, error) {
	ln, err := net.Listen("tcp", preferred)
	if err != nil {
		slog.Warn("dns: port 53 unavailable, falling back", "fallback", fallback)
		if err := s.Start(fallback); err != nil {
			return "", fmt.Errorf("dns: start fallback: %w", err)
		}
		return fallback, nil
	}
	ln.Close()

	if err := s.Start(preferred); err != nil {
		slog.Warn("dns: preferred addr failed, trying fallback", "err", err)
		if err := s.Start(fallback); err != nil {
			return "", fmt.Errorf("dns: start fallback: %w", err)
		}
		return fallback, nil
	}
	return preferred, nil
}

func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var firstErr error
	if s.udp != nil {
		if err := s.udp.Shutdown(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if s.tcp != nil {
		if err := s.tcp.Shutdown(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (s *Server) handleDNS(w mdns.ResponseWriter, r *mdns.Msg) {
	msg := new(mdns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true

	for _, q := range r.Question {
		name := strings.ToLower(q.Name)
		if !strings.HasSuffix(name, ".web.") && !strings.HasSuffix(name, ".local.") {
			msg.Rcode = mdns.RcodeServerFailure
			w.WriteMsg(msg)
			return
		}

		switch q.Qtype {
		case mdns.TypeA:
			msg.Answer = append(msg.Answer, &mdns.A{
				Hdr: mdns.RR_Header{
					Name:   q.Name,
					Rrtype: mdns.TypeA,
					Class:  mdns.ClassINET,
					Ttl:    60,
				},
				A: net.ParseIP("127.0.0.1"),
			})
		case mdns.TypeAAAA:
			msg.Answer = append(msg.Answer, &mdns.AAAA{
				Hdr: mdns.RR_Header{
					Name:   q.Name,
					Rrtype: mdns.TypeAAAA,
					Class:  mdns.ClassINET,
					Ttl:    60,
				},
				AAAA: net.ParseIP("::1"),
			})
		}
	}

	w.WriteMsg(msg)
}

func (s *Server) Addr() string {
	return s.addr
}
