// Copyright 2018 Mark Wardle / Eldrix Ltd
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.
//

package terminology

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/gogo/protobuf/io"
	"golang.org/x/text/language"

	"github.com/wardle/go-terminology/snomed"
)

// Export exports all descriptions in delimited protobuf format to the command line.
func (svc *Svc) Export() error {
	w := io.NewDelimitedWriter(os.Stdout)
	defer w.Close()

	count := 0
	start := time.Now()
	svc.Iterate(func(concept *snomed.Concept) error {
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
			w.WriteMsg(&edCopy)
			count++
			if count%10000 == 0 {
				elapsed := time.Since(start)
				fmt.Fprintf(os.Stderr, "\rProcessed %d descriptions in %s. Mean time per description: %s...", count, elapsed, elapsed/time.Duration(count))
				return errors.New("End of run")
			}
		}
		return nil
	})
	fmt.Fprintf(os.Stderr, "\nProcessed total: %d descriptions in %s.\n", count, time.Since(start))
	return nil
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
	ed.ConceptRefsets, err = svc.GetReferenceSets(concept.Id) // get reference sets for concept
	if err != nil {
		return ed, err
	}

	return ed, nil
}

// TODO: pass language as a parameter rather than hard-coding British English
func updateExtendedDescriptionFromDescription(svc *Svc, ed *snomed.ExtendedDescription, description *snomed.Description) error {
	var err error
	ed.Description = description
	ed.DescriptionRefsets, err = svc.GetReferenceSets(description.Id) // reference sets for description
	if err != nil {
		return err
	}
	return nil
}
