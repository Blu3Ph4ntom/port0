package state

import (
	"sync"
	"testing"
	"time"
)

func tempStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	return NewStore(dir)
}

func TestLoadEmpty(t *testing.T) {
	s := tempStore(t)
	st, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Projects) != 0 {
		t.Errorf("expected empty projects, got %d", len(st.Projects))
	}
}

func TestSaveAndLoad(t *testing.T) {
	s := tempStore(t)
	now := time.Now().Truncate(time.Second)
	p := &Project{
		Name:      "myapp",
		Port:      4200,
		Cmd:       []string{"npm", "run", "dev"},
		Cwd:       "/home/user/myapp",
		PID:       12345,
		StartedAt: now,
		LogFile:   "/home/user/.port0/logs/myapp.log",
		Restart:   "no",
	}

	if err := s.Set(p); err != nil {
		t.Fatal(err)
	}

	got, err := s.Get("myapp")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "myapp" {
		t.Errorf("name = %q, want myapp", got.Name)
	}
	if got.Port != 4200 {
		t.Errorf("port = %d, want 4200", got.Port)
	}
	if got.PID != 12345 {
		t.Errorf("pid = %d, want 12345", got.PID)
	}
}

func TestGetNotFound(t *testing.T) {
	s := tempStore(t)
	_, err := s.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent project")
	}
}

func TestDelete(t *testing.T) {
	s := tempStore(t)
	p := &Project{Name: "todelete", Port: 4001, Restart: "no"}
	if err := s.Set(p); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete("todelete"); err != nil {
		t.Fatal(err)
	}
	_, err := s.Get("todelete")
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestConcurrentWrite(t *testing.T) {
	s := tempStore(t)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			p := &Project{
				Name:    "myapp",
				Port:    4000 + i,
				Restart: "no",
			}
			s.Set(p)
		}(i)
	}
	wg.Wait()

	got, err := s.Get("myapp")
	if err != nil {
		t.Fatal(err)
	}
	if got.Port < 4000 || got.Port > 4019 {
		t.Errorf("unexpected port %d", got.Port)
	}
}

func TestAll(t *testing.T) {
	s := tempStore(t)
	s.Set(&Project{Name: "a", Port: 4001, Restart: "no"})
	s.Set(&Project{Name: "b", Port: 4002, Restart: "no"})

	all, err := s.All()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 projects, got %d", len(all))
	}
}
