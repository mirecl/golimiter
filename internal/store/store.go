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
	Pass *analysis.Pass
	Pos  token.Pos
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
	issues []*Issue
	mu     sync.Mutex
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

func (s *store) Issues() []*Issue {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.issues
}

func New() Store {
	return &store{
		mu:     sync.Mutex{},
		issues: make([]*Issue, 0),
	}
}
