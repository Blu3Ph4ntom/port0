package allocator

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
)

type Allocator struct {
	Min int
	Max int
}

func New(min, max int) *Allocator {
	return &Allocator{Min: min, Max: max}
}

func ParseRange(s string) (*Allocator, error) {
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("allocator: invalid range %q, expected min-max", s)
	}
	min, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("allocator: invalid min %q: %w", parts[0], err)
	}
	max, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("allocator: invalid max %q: %w", parts[1], err)
	}
	if min >= max {
		return nil, fmt.Errorf("allocator: min (%d) must be less than max (%d)", min, max)
	}
	return New(min, max), nil
}

func (a *Allocator) Pick(taken map[int]bool) (int, error) {
	size := a.Max - a.Min + 1
	perm := rand.Perm(size)
	for _, offset := range perm {
		port := a.Min + offset
		if taken[port] {
			continue
		}
		if !isPortFree(port) {
			continue
		}
		return port, nil
	}
	return 0, fmt.Errorf("allocator: all ports in range %d-%d are taken", a.Min, a.Max)
}

func isPortFree(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

func TakenPorts(projects map[int]bool) map[int]bool {
	return projects
}

func TakenPortsFromMap(projects map[string]int) map[int]bool {
	taken := make(map[int]bool, len(projects))
	for _, port := range projects {
		taken[port] = true
	}
	return taken
}
