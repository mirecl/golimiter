package store

import (
	"fmt"
	"go/token"
	"sync"

	"golang.org/x/tools/go/analysis"
)

// Store is interface for global state for each analyzer.
type Store interface {
	// Add new issue.
	Add(value *Issue)
	// Get all issues.
	Issues() []*Issue
	// Get count all issues.
	Len() int
	// Report problem.
	Report(msg string)
	// Report problem.
	Reportf(format string, args ...interface{})
}

// Issue problem in analysis.
type Issue struct {
	Pass *analysis.Pass
	Pos  token.Pos
	done bool
}

// Report set problem.
func (i *Issue) Report(msg string) {
	// check status issue - report problem or not.
	if i.done {
		return
	}

	i.Pass.Report(analysis.Diagnostic{
		Pos:     i.Pos,
		Message: msg,
	})

	i.done = true
}

// Reportf set problem.
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

// New create global state.
func New() Store {
	return &store{
		mu:     sync.Mutex{},
		issues: make([]*Issue, 0),
	}
}
