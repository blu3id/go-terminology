package search

import (
	"github.com/wardle/go-terminology/snomed"
)

// Search is an interface to the pluggable backend abstracted SNOMED-CT
// search service. A search service must implement this interface.
type Search interface {
	// Search executes a search request and returns description identifiers
	Search(search *snomed.SearchRequest) ([]int64, error)
	Index(extendedDescriptions []*snomed.ExtendedDescription) error
	Close() error
}
