package storage

import (
	"fmt"
	"strings"

	"github.com/wardle/go-terminology/snomed"
)

// Store is an interface to the pluggable backend abstracted SNOMED-CT
// persistence service. A storage service must implement this interface.
type Store interface {
	GetConcept(conceptID int64) (*snomed.Concept, error)
	GetConcepts(conceptIsvc ...int64) ([]*snomed.Concept, error)
	GetDescription(descriptionID int64) (*snomed.Description, error)
	GetDescriptions(concept *snomed.Concept) ([]*snomed.Description, error)
	GetParentRelationships(concept *snomed.Concept) ([]*snomed.Relationship, error)
	GetChildRelationships(concept *snomed.Concept) ([]*snomed.Relationship, error)
	GetAllChildrenIDs(concept *snomed.Concept) ([]int64, error)
	GetReferenceSets(componentID int64) ([]int64, error)
	GetReferenceSetItems(refset int64) (map[int64]bool, error)
	GetFromReferenceSet(refset int64, component int64) (*snomed.ReferenceSetItem, error)
	GetAllReferenceSets() ([]int64, error) // list of installed reference sets
	Put(components interface{}) error
	Iterate(fn func(*snomed.Concept) error) error
	GetStatistics() (Statistics, error)
	Close() error
}

// Statistics on the persistence store
type Statistics struct {
	Concepts      int
	Descriptions  int
	Relationships int
	RefsetItems   int
	Refsets       []string
}

// String produces formated output of persistence store statistics
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
