package store

import (
	"fmt"
	"go/token"
	"sync"

	"golang.org/x/tools/go/analysis"
)

type Store interface {
	Add(value *Issue)
	Issues() []*Issue
	Len() int
	Report(msg string)
	Reportf(format string, args ...interface{})
}

type Issue struct {
	Pos  token.Pos
	Pass *analysis.Pass
	done bool
}

func (i *Issue) Report(msg string) {
	if i.done {
		return
	}

	i.Pass.Report(analysis.Diagnostic{
		Pos:     i.Pos,
		Message: msg,
	})

	i.done = true
}

func (i *Issue) Reportf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	i.Report(msg)

}

type store struct {
	mu     sync.Mutex
	issues []*Issue
}

func (s *store) Add(value *Issue) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.issues = append(s.issues, value)
}

func (s *store) Reportf(format string, args ...interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, issue := range s.issues {
		msg := fmt.Sprintf(format, args...)
		issue.Report(msg)
	}
}

func (s *store) Report(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, issue := range s.issues {
		issue.Report(msg)
	}
}

func (s *store) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.issues)
}

func (i *store) Issues() []*Issue {
	i.mu.Lock()
	defer i.mu.Unlock()

	return i.issues
}

func New() Store {
	return &store{
		mu:     sync.Mutex{},
		issues: make([]*Issue, 0),
	}
}
