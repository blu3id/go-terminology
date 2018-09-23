package terminology

import (
	"fmt"
	"os"
	"time"

	"github.com/wardle/go-terminology/snomed"
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
			if count%1000 == 0 {
				err = svc.search.Index(eds)
				if err != nil {
					panic(err)
				}
				eds = []*snomed.ExtendedDescription{}
				elapsed := time.Since(start)
				fmt.Fprintf(os.Stderr, "\rProcessed %d descriptions in %s. Mean time per description: %s...", count, elapsed, elapsed/time.Duration(count))
			}
		}
		return nil
	})
	fmt.Fprintf(os.Stderr, "\nProcessed total: %d descriptions in %s.\n", count, time.Since(start))
	return err
}

func (svc *Svc) IndexConcept(concept *snomed.Concept) error {
	var eds []*snomed.ExtendedDescription
	var ed snomed.ExtendedDescription
	ed, err := createExtendedDescriptionFromConcept(svc, concept)
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
	}
	err = svc.search.Index(eds)
	return err
}
