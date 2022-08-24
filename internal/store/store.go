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

type issues struct {
	mu     sync.Mutex
	issues []*Issue
}

func (i *issues) Add(value *Issue) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.issues = append(i.issues, value)
}

func (i *issues) Reportf(format string, args ...interface{}) {
	i.mu.Lock()
	defer i.mu.Unlock()

	for _, issue := range i.issues {
		msg := fmt.Sprintf(format, args...)
		issue.Report(msg)
	}
}

func (i *issues) Report(msg string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	for _, issue := range i.issues {
		issue.Report(msg)
	}
}

func (i *issues) Len() int {
	i.mu.Lock()
	defer i.mu.Unlock()

	return len(i.issues)
}

func (i *issues) Issues() []*Issue {
	i.mu.Lock()
	defer i.mu.Unlock()

	return i.issues
}

func New() Store {
	return &issues{
		mu:     sync.Mutex{},
		issues: make([]*Issue, 0),
	}
}
