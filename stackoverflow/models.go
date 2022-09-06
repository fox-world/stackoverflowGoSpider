package stackoverflow

import (
	"sync"
	"time"
)

type Post struct {
	Title        string
	Link         string
	Tags         []string
	PostUser     string
	PostUserLink string
	PostTime     time.Time
	Vote         int
	Viewed       int
}

type Status struct {
	mu  sync.Mutex
	run bool
}

func (s *Status) UpdateStatus(status bool) {
	s.mu.Lock()
	s.run = status
	s.mu.Unlock()
}

func (s *Status) IsRun() (stop bool) {
	s.mu.Lock()
	stop = s.run
	s.mu.Unlock()
	return
}
