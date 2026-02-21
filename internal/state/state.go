package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Project struct {
	Name      string    `json:"name"`
	Port      int       `json:"port"`
	Cmd       []string  `json:"cmd"`
	Cwd       string    `json:"cwd"`
	PID       int       `json:"pid"`
	StartedAt time.Time `json:"started_at"`
	LogFile   string    `json:"log_file"`
	Restart   string    `json:"restart"`
	Domain    string    `json:"domain,omitempty"` // parent domain for subdomain routing (e.g., "myapp" for api.myapp.localhost)
}

type State struct {
	Projects map[string]*Project `json:"projects"`
}

type Store struct {
	mu   sync.Mutex
	path string
}

func NewStore(dir string) *Store {
	return &Store{
		path: filepath.Join(dir, "state.json"),
	}
}

func DefaultStore() *Store {
	home, _ := os.UserHomeDir()
	return NewStore(filepath.Join(home, ".port0"))
}

func (s *Store) Path() string {
	return s.path
}

func (s *Store) Load() (*State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadLocked()
}

func (s *Store) loadLocked() (*State, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{Projects: make(map[string]*Project)}, nil
		}
		return nil, fmt.Errorf("state: read: %w", err)
	}

	var st State
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, fmt.Errorf("state: unmarshal: %w", err)
	}
	if st.Projects == nil {
		st.Projects = make(map[string]*Project)
	}
	return &st, nil
}

func (s *Store) Save(st *State) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked(st)
}

func (s *Store) saveLocked(st *State) error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("state: mkdir: %w", err)
	}

	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return fmt.Errorf("state: marshal: %w", err)
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("state: write tmp: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("state: rename: %w", err)
	}
	return nil
}

func (s *Store) Get(name string) (*Project, error) {
	st, err := s.Load()
	if err != nil {
		return nil, err
	}
	p, ok := st.Projects[name]
	if !ok {
		return nil, fmt.Errorf("state: project not found: %s", name)
	}
	return p, nil
}

func (s *Store) Set(p *Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	st, err := s.loadLocked()
	if err != nil {
		return err
	}
	st.Projects[p.Name] = p
	return s.saveLocked(st)
}

func (s *Store) Delete(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	st, err := s.loadLocked()
	if err != nil {
		return err
	}
	delete(st.Projects, name)
	return s.saveLocked(st)
}

func (s *Store) All() (map[string]*Project, error) {
	st, err := s.Load()
	if err != nil {
		return nil, err
	}
	return st.Projects, nil
}
