package terminology

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/wardle/go-terminology/snomed"
	"golang.org/x/text/language"
)

// Index orchestrates gouroutine build process of snomed.ExtendedDescription for
// each concept in datastore and passes to the search service for indexing.
func (svc *Svc) Index() error {
	var (
		wg          sync.WaitGroup
		conceptsChn = make(chan snomed.Concept)
		indexChn    = make(chan snomed.ExtendedDescription, 1000) // Create buffered channel
	)
	go func() {
		var count = 0
		svc.Iterate(func(concept *snomed.Concept) error {
			conceptsChn <- *concept
			count++
			if count == 10000 {
				return fmt.Errorf("Break")
			}
			return nil
		})
		close(conceptsChn) // Close conceptsChn to signal to buildExtendedDescriptionsFromConcept goroutines to return
		wg.Wait()          // Wait for all buildExtendedDescriptionsFromConcept goroutines to finish
		close(indexChn)    // Close indexChan to signal to batchIndex to return
	}()

	cpu := runtime.NumCPU()
	for i := 0; i < cpu; i++ {
		wg.Add(1) // Add to WaitGroup for each goroutines launched
		go buildExtendedDescriptionsFromConcept(svc, &wg, conceptsChn, indexChn)
	}
	batchIndex(svc, indexChn) // Start batch indexer
	return nil
}

// batchIndex receives snomed.ExtendedDescription on a channel and batches them
// before writing to the search service using Search.Index()
func batchIndex(svc *Svc, in <-chan snomed.ExtendedDescription) {
	var (
		count = 0
		eds   []*snomed.ExtendedDescription
	)
	start := time.Now()
	for ed := range in {
		copyEd := ed
		eds = append(eds, &copyEd)
		count++
		if count%10000 == 0 {
			if err := svc.Search.Index(eds); err != nil {
				panic(err)
			}
			eds = []*snomed.ExtendedDescription{}
			elapsed := time.Since(start)
			fmt.Fprintf(os.Stderr, "\rProcessed %d descriptions in %s. Mean time per description: %s...", count, elapsed, elapsed/time.Duration(count))
		}
	}
	if err := svc.Search.Index(eds); err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stderr, "\nProcessed total: %d descriptions in %s.\n", count, time.Since(start))
}

// buildExtendedDescriptionsFromConcept reads from conceptsChn and builds a
// snomed.ExtendedDescription for each description in the specified concept and
// passes to indexChn channel
func buildExtendedDescriptionsFromConcept(svc *Svc, wg *sync.WaitGroup, conceptsChn chan snomed.Concept, indexChn chan snomed.ExtendedDescription) {
	defer wg.Done()
	for concept := range conceptsChn {
		var (
			ed         snomed.ExtendedDescription
			err        error
			tags, _, _ = language.ParseAcceptLanguage("en-GB")
		)

		ed.Concept = &concept
		ed.PreferredDescription = svc.MustGetPreferredSynonym(&concept, tags)
		ed.RecursiveParentIds, err = svc.GetAllParentIDs(&concept)
		if err != nil {
			panic(err)
		}
		ed.DirectParentIds, err = svc.GetParentIDsOfKind(&concept, snomed.IsA)
		if err != nil {
			panic(err)
		}
		ed.ConceptRefsets, err = svc.GetReferenceSets(concept.Id)
		if err != nil {
			panic(err)
		}

		descriptions, err := svc.GetDescriptions(&concept)
		if err != nil {
			panic(err)
		}
		for _, description := range descriptions {
			edCopy := ed
			edCopy.Description = description
			edCopy.DescriptionRefsets, err = svc.GetReferenceSets(description.Id)
			if err != nil {
				panic(err)
			}
			indexChn <- edCopy
		}
	}
}

/*
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
		return nil
	})
	err = svc.Search.Index(eds)
	if err != nil {
		panic(err)
	}

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

func updateExtendedDescriptionFromDescription(svc *Svc, ed *snomed.ExtendedDescription, description *snomed.Description) error {
	var err error
	ed.Description = description
	ed.DescriptionRefsets, err = svc.GetReferenceSets(description.Id)
	if err != nil {
		return err
	}
	return nil
}
*/
