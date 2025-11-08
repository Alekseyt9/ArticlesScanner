package scanner

import (
	"context"
	"fmt"
	"time"

	"ArticlesScanner/internal/domain"
)

// Category describes a concrete section endpoint provided by config.
type Category struct {
	Name string
	URL  string
}

// Request carries all parameters required to execute a scan.
type Request struct {
	Day        time.Time
	SiteName   string
	Categories []Category
	Options    map[string]string
}

// Scanner captures a single strategy implementation (Arxiv, IEEE, etc.).
type Scanner interface {
	Name() string
	Scan(ctx context.Context, req Request) ([]domain.Article, error)
}

// Registry keeps a mapping from scanner names to their implementations.
type Registry struct {
	scanners map[string]Scanner
}

// NewRegistry builds an empty registry.
func NewRegistry() *Registry {
	return &Registry{scanners: map[string]Scanner{}}
}

// Register adds or replaces a scanner implementation.
func (r *Registry) Register(scanner Scanner) {
	if r.scanners == nil {
		r.scanners = map[string]Scanner{}
	}
	r.scanners[scanner.Name()] = scanner
}

// Resolve returns a scanner by name or an error if it is absent.
func (r *Registry) Resolve(name string) (Scanner, error) {
	if scanner, ok := r.scanners[name]; ok {
		return scanner, nil
	}
	return nil, fmt.Errorf("scanner %s is not registered", name)
}
