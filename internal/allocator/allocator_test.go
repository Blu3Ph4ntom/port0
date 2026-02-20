package allocator

import (
	"net"
	"testing"
)

func TestPickReturnsPortInRange(t *testing.T) {
	a := New(14000, 14010)
	port, err := a.Pick(nil)
	if err != nil {
		t.Fatal(err)
	}
	if port < 14000 || port > 14010 {
		t.Errorf("port %d not in range 14000-14010", port)
	}
}

func TestPickAvoidsTaken(t *testing.T) {
	a := New(14020, 14022)
	taken := map[int]bool{14020: true, 14021: true}
	port, err := a.Pick(taken)
	if err != nil {
		t.Fatal(err)
	}
	if port != 14022 {
		t.Errorf("expected 14022, got %d", port)
	}
}

func TestPickExhausted(t *testing.T) {
	a := New(14030, 14031)

	ln1, err := net.Listen("tcp", "127.0.0.1:14030")
	if err != nil {
		t.Skip("cannot bind 14030")
	}
	defer ln1.Close()

	ln2, err := net.Listen("tcp", "127.0.0.1:14031")
	if err != nil {
		t.Skip("cannot bind 14031")
	}
	defer ln2.Close()

	_, err = a.Pick(nil)
	if err == nil {
		t.Fatal("expected error when all ports exhausted")
	}
}

func TestParseRange(t *testing.T) {
	a, err := ParseRange("5000-5999")
	if err != nil {
		t.Fatal(err)
	}
	if a.Min != 5000 || a.Max != 5999 {
		t.Errorf("got %d-%d, want 5000-5999", a.Min, a.Max)
	}

	_, err = ParseRange("invalid")
	if err == nil {
		t.Fatal("expected error for invalid range")
	}

	_, err = ParseRange("5000-4000")
	if err == nil {
		t.Fatal("expected error for min >= max")
	}
}
