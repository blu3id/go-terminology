package interfaces

import (
	"fmt"
	"strings"

	"github.com/wardle/go-terminology/snomed"
	"github.com/wardle/go-terminology/terminology/medicine"
)

// Statistics on the persistence store
type Statistics struct {
	Concepts      int
	Descriptions  int
	Relationships int
	RefsetItems   int
	Refsets       []string
}

func (st Statistics) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Number of concepts: %d\n", st.Concepts))
	b.WriteString(fmt.Sprintf("Number of descriptions: %d\n", st.Descriptions))
	b.WriteString(fmt.Sprintf("Number of relationships: %d\n", st.Relationships))
	b.WriteString(fmt.Sprintf("Number of reference set items: %d\n", st.RefsetItems))
	b.WriteString(fmt.Sprintf("Number of installed refsets: %d:\n", len(st.Refsets)))
	for _, s := range st.Refsets {
		b.WriteString(fmt.Sprintf("  Installed refset: %s\n", s))
	}
	return b.String()
}

// Store represents the backend opaque abstract SNOMED-CT persistence service.
type Store interface {
	GetConcept(conceptID int64) (*snomed.Concept, error)
	GetConcepts(conceptIsvc ...int64) ([]*snomed.Concept, error)
	GetDescription(descriptionID int64) (*snomed.Description, error)
	GetDescriptions(concept *snomed.Concept) ([]*snomed.Description, error)
	GetParentRelationships(concept *snomed.Concept) ([]*snomed.Relationship, error)
	GetChildRelationships(concept *snomed.Concept) ([]*snomed.Relationship, error)
	GetReferenceSets(componentID int64) ([]int64, error)
	GetReferenceSetItems(refset int64) (map[int64]bool, error)
	GetFromReferenceSet(refset int64, component int64) (*snomed.ReferenceSetItem, error)
	GetAllReferenceSets() ([]int64, error) // list of installed reference sets
	Put(components interface{}) error
	Iterate(fn func(*snomed.Concept) error) error
	GetStatistics() (Statistics, error)
	Close() error

	GetAllChildrenIDs(concept *snomed.Concept) ([]int64, error)
}

// SearchRequest is used to set the parameters on which to search
type SearchRequest struct {
	Search                string  `schema:"s"`                     // search term
	RecursiveParents      []int64 `schema:"root"`                  // one or more root concept identifiers (default 138875005)
	DirectParents         []int64 `schema:"is"`                    // zero or more direct parent concept identifiers
	Refsets               []int64 `schema:"refset"`                // filter to concepts within zero of more refsets
	Limit                 int     `schema:"maxHits"`               // number of hits (default 200)
	IncludeInactive       bool    `schema:"inactive"`              // whether to include inactive terms in search results (defaults to False)
	Fuzzy                 bool    `schema:"fuzzy"`                 // whether to use a fuzzy search for search (default to False)
	SuppressFallbackFuzzy bool    `schema:"suppressFuzzyFallback"` // whether to suppress automatic fallback to fuzzy search if no results found for non-fuzzy search (defaults to False)
}

// Search represents an opaque abstract SNOMED-CT search service.
type Search interface {
	// Search executes a search request and returns description identifiers
	Search(search *SearchRequest) ([][]int64, error)
	Index(indexedConcepts []*snomed.IndexedConcept) error
	ParseMedicationString(medication string) (*medicine.ParsedMedication, error)
	Close() error
}
