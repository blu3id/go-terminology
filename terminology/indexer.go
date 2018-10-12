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
	var eds []*snomed.ExtendedDescription

	count := 0
	start := time.Now()
	err := svc.Iterate(func(concept *snomed.Concept) error {
		var ed snomed.ExtendedDescription
		var err error
		ed, err = createExtendedDescriptionFromConcept(svc, concept)
		if err != nil {
			panic(err)
		}
		descs, err := svc.GetDescriptions(concept)
		if err != nil {
			panic(err)
		}
		for _, d := range descs {
			edCopy := ed
			err = updateExtendedDescriptionFromDescription(svc, &edCopy, d)
			if err != nil {
				panic(err)
			}
			eds = append(eds, &edCopy)
			count++
			if count%10000 == 0 {
				err = svc.Search.Index(eds)
				if err != nil {
					panic(err)
				}
				eds = []*snomed.ExtendedDescription{}
				elapsed := time.Since(start)
				fmt.Fprintf(os.Stderr, "\rProcessed %d descriptions in %s. Mean time per description: %s...", count, elapsed, elapsed/time.Duration(count))
			}
		}
		err = svc.Search.Index(eds)
		if err != nil {
			panic(err)
		}
		return nil
	})
	fmt.Fprintf(os.Stderr, "\nProcessed total: %d descriptions in %s.\n", count, time.Since(start))
	return err
}

// TODO: pass language as a parameter rather than hard-coding British English
func createExtendedDescriptionFromConcept(svc *Svc, concept *snomed.Concept) (snomed.ExtendedDescription, error) {
	var ed snomed.ExtendedDescription
	var err error
	tags, _, _ := language.ParseAcceptLanguage("en-GB")

	copyConcept := *concept
	ed.Concept = &copyConcept
	ed.PreferredDescription = svc.MustGetPreferredSynonym(concept, tags)
	ed.RecursiveParentIds, err = svc.GetAllParentIDs(concept)
	if err != nil {
		return ed, err
	}
	ed.DirectParentIds, err = svc.GetParentIDsOfKind(concept, snomed.IsA)
	if err != nil {
		return ed, err
	}
	ed.ConceptRefsets, err = svc.GetReferenceSets(concept.Id)
	if err != nil {
		return ed, err
	}

	return ed, nil
}

// TODO: pass language as a parameter rather than hard-coding British English
func updateExtendedDescriptionFromDescription(svc *Svc, ed *snomed.ExtendedDescription, description *snomed.Description) error {
	var err error
	ed.Description = description
	ed.DescriptionRefsets, err = svc.GetReferenceSets(description.Id)
	if err != nil {
		return err
	}
	return nil
}
