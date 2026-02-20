package dns

import (
	"fmt"
	"net"
	"testing"
	"time"

	mdns "github.com/miekg/dns"
)

func findFreeUDPPort() int {
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return 15353
	}
	port := conn.LocalAddr().(*net.UDPAddr).Port
	conn.Close()
	return port
}

func TestDNSServerARecord(t *testing.T) {
	port := findFreeUDPPort()
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	srv := New()
	if err := srv.Start(addr); err != nil {
		t.Fatal(err)
	}
	defer srv.Stop()

	time.Sleep(100 * time.Millisecond)

	c := new(mdns.Client)
	m := new(mdns.Msg)
	m.SetQuestion("myapp.web.", mdns.TypeA)

	r, _, err := c.Exchange(m, addr)
	if err != nil {
		t.Fatal(err)
	}

	if len(r.Answer) == 0 {
		t.Fatal("expected at least one answer")
	}

	a, ok := r.Answer[0].(*mdns.A)
	if !ok {
		t.Fatalf("expected A record, got %T", r.Answer[0])
	}
	if a.A.String() != "127.0.0.1" {
		t.Errorf("expected 127.0.0.1, got %s", a.A.String())
	}
}

func TestDNSServerAAAARecord(t *testing.T) {
	port := findFreeUDPPort()
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	srv := New()
	if err := srv.Start(addr); err != nil {
		t.Fatal(err)
	}
	defer srv.Stop()

	time.Sleep(100 * time.Millisecond)

	c := new(mdns.Client)
	m := new(mdns.Msg)
	m.SetQuestion("myapp.web.", mdns.TypeAAAA)

	r, _, err := c.Exchange(m, addr)
	if err != nil {
		t.Fatal(err)
	}

	if len(r.Answer) == 0 {
		t.Fatal("expected at least one answer")
	}

	aaaa, ok := r.Answer[0].(*mdns.AAAA)
	if !ok {
		t.Fatalf("expected AAAA record, got %T", r.Answer[0])
	}
	if aaaa.AAAA.String() != "::1" {
		t.Errorf("expected ::1, got %s", aaaa.AAAA.String())
	}
}

func TestDNSServerNonWebDomain(t *testing.T) {
	port := findFreeUDPPort()
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	srv := New()
	if err := srv.Start(addr); err != nil {
		t.Fatal(err)
	}
	defer srv.Stop()

	time.Sleep(100 * time.Millisecond)

	c := new(mdns.Client)
	m := new(mdns.Msg)
	m.SetQuestion("google.com.", mdns.TypeA)

	r, _, err := c.Exchange(m, addr)
	if err != nil {
		t.Fatal(err)
	}

	if r.Rcode != mdns.RcodeServerFailure {
		t.Errorf("expected SERVFAIL for non-.web domain, got rcode %d", r.Rcode)
	}
}
