package server

import (
	"github.com/wardle/go-terminology/snomed"
	"github.com/wardle/go-terminology/terminology"
	"golang.org/x/net/context"
	"golang.org/x/text/language"
)

// searchSrv implements the snomed.SearchServer gRPC interface
type searchSrv struct {
	snomed.SearchServer
	svc *terminology.Svc
}

// Search implements the Search gRPC server method returning a *snomed.SearchResponse result
func (ss *searchSrv) Search(ctx context.Context, searchRequest *snomed.SearchRequest) (*snomed.SearchResponse, error) {
	var output snomed.SearchResponse
	results, err := ss.svc.Search.Search(searchRequest)
	if err != nil {
		return &output, err
	}

	for _, v := range results {
		description, err := ss.svc.GetDescription(v)
		if err != nil {
			return &output, err
		}
		concept, err := ss.svc.GetConcept(description.ConceptId)
		if err != nil {
			return &output, err
		}

		tags, _, _ := language.ParseAcceptLanguage(searchRequest.AcceptedLanguages)
		preferredDescription := ss.svc.MustGetPreferredSynonym(concept, tags)

		output.Items = append(output.Items, &snomed.SearchResponse_Item{
			Term:          description.Term,
			ConceptId:     description.ConceptId,
			PreferredTerm: preferredDescription.Term,
		})
	}

	return &output, nil
}

/*
// Search implementation for streaming results
func (ss *searchSrv) Search(searchRequest *snomed.SearchRequest, server snomed.Search_SearchServer) error {
	results, err := ss.svc.Search.Search(searchRequest)
	if err != nil {
		return err
	}

	for _, v := range results {
		description, err := ss.svc.GetDescription(v)
		if err != nil {
			return err
		}
		concept, err := ss.svc.GetConcept(description.ConceptId)
		if err != nil {
			return err
		}

		tags, _, _ := language.ParseAcceptLanguage(searchRequest.AcceptedLanguages)
		preferredDescription := ss.svc.MustGetPreferredSynonym(concept, tags)

		err = server.Send(&snomed.SearchResponse_Item{
			Term:          description.Term,
			ConceptId:     description.ConceptId,
			PreferredTerm: preferredDescription.Term,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
*/
