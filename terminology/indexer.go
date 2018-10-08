package terminology

import (
	"fmt"
	"os"
	"time"

	"github.com/wardle/go-terminology/snomed"
	"golang.org/x/text/language"
)

// Index build full descriptions and passes to the search service. Esentially complete clone of Export.
// TODO: Refactor Export to make portable across both Index and Export
func (svc *Svc) Index() error {
	var ics []*snomed.IndexedConcept

	count := 0
	start := time.Now()
	err := svc.Iterate(func(concept *snomed.Concept) error {
		var ic snomed.IndexedConcept
		var err error
		ic, err = createIndexedConcept(svc, concept)
		if err != nil {
			panic(err)
		}
		descs, err := svc.GetDescriptions(concept)
		if err != nil {
			panic(err)
		}
		for _, d := range descs {
			DescriptionRefsets, err := svc.GetReferenceSets(d.Id)
			if err != nil {
				return err
			}
			ic.Descriptions = append(ic.Descriptions, &snomed.IndexedConcept_DescriptionWithRefsets{
				Description:        d,
				DescriptionRefsets: DescriptionRefsets,
			})
		}
		ics = append(ics, &ic)
		count++
		if count%1000 == 0 {
			err = svc.Search.Index(ics)
			if err != nil {
				panic(err)
			}
			ics = []*snomed.IndexedConcept{}
			elapsed := time.Since(start)
			fmt.Fprintf(os.Stderr, "\rProcessed %d concepts in %s. Mean time per concepts: %s...", count, elapsed, elapsed/time.Duration(count))
		}
		return nil
	})
	fmt.Fprintf(os.Stderr, "\nProcessed total: %d concepts in %s.\n", count, time.Since(start))
	return err
}

// TODO: pass language as a parameter rather than hard-coding British English
func createIndexedConcept(svc *Svc, concept *snomed.Concept) (snomed.IndexedConcept, error) {
	var ic snomed.IndexedConcept
	var err error
	tags, _, _ := language.ParseAcceptLanguage("en-GB")

	copyConcept := *concept
	ic.Concept = &copyConcept
	ic.PreferredDescription = svc.MustGetPreferredSynonym(concept, tags)
	ic.RecursiveParentIds, err = svc.GetAllParentIDs(concept)
	if err != nil {
		return ic, err
	}
	ic.DirectParentIds, err = svc.GetParentIDsOfKind(concept, snomed.IsA)
	if err != nil {
		return ic, err
	}
	ic.ConceptRefsets, err = svc.GetReferenceSets(concept.Id) // get reference sets for concept
	if err != nil {
		return ic, err
	}

	return ic, nil
}
