package server

import (
	"fmt"

	"github.com/wardle/go-terminology/snomed"
	"github.com/wardle/go-terminology/terminology"
	"golang.org/x/net/context"
	"golang.org/x/text/language"
)

// snomedCTSrv implements the snomed.SnomedCTServer gRPC interface
type snomedCTSrv struct {
	snomed.SnomedCTServer
	svc *terminology.Svc
}

func (ss *snomedCTSrv) GetConcept(ctx context.Context, conceptID *snomed.SctID) (*snomed.Concept, error) {
	return ss.svc.GetConcept(conceptID.Identifier)
}

func (ss *snomedCTSrv) GetExtendedConcept(ctx context.Context, conceptID *snomed.SctID) (*snomed.ExtendedConcept, error) {
	c, err := ss.svc.GetConcept(conceptID.Identifier)
	if err != nil {
		return nil, err
	}
	result := snomed.ExtendedConcept{}
	result.Concept = c
	refsets, err := ss.svc.GetReferenceSets(c.Id)
	if err != nil {
		return nil, err
	}
	result.ConceptRefsets = refsets
	relationships, err := ss.svc.GetParentRelationships(c)
	if err != nil {
		return nil, err
	}
	result.Relationships = relationships
	recursiveParentIDs, err := ss.svc.GetAllParentIDs(c)
	if err != nil {
		return nil, err
	}
	result.RecursiveParentIds = recursiveParentIDs
	directParents, err := ss.svc.GetParentIDsOfKind(c, snomed.IsA)
	if err != nil {
		return nil, err
	}
	result.DirectParentIds = directParents
	tags, _, _ := language.ParseAcceptLanguage("en-GB") // TODO(mw): better language support
	result.PreferredDescription = ss.svc.MustGetPreferredSynonym(c, tags)
	return &result, nil
}

func (ss *snomedCTSrv) GetDescriptions(conceptID *snomed.SctID, server snomed.SnomedCT_GetDescriptionsServer) error {
	c, err := ss.svc.GetConcept(conceptID.Identifier)
	if err != nil {
		return err
	}
	descs, err := ss.svc.GetDescriptions(c)
	if err != nil {
		return err
	}
	for _, d := range descs {
		server.Send(d)
	}
	return nil
}

func (ss *snomedCTSrv) GetDescription(ctx context.Context, id *snomed.SctID) (*snomed.Description, error) {
	return ss.svc.GetDescription(id.Identifier)
}

func (ss *snomedCTSrv) Translate(ctx context.Context, tr *snomed.TranslateRequest) (*snomed.TranslateResponse, error) {
	response := snomed.TranslateResponse{}
	target, err := ss.svc.GetFromReferenceSet(tr.TargetId, tr.ConceptId)
	if err != nil {
		return nil, err
	}
	if target != nil { // we have found our concept in the reference set, so return that entry
		simple := target.GetSimple() // a simple refset
		if simple != nil {           // found concept in a simple map, so just return it.
			rc := snomed.TranslateResponse_Concept{}
			rc.Concept, err = ss.svc.GetConcept(tr.ConceptId)
			response.Result = &rc
			return &response, err
		}
		refset := snomed.TranslateResponse_ReferenceSetItem{} // otherwise return the reference set item
		refset.ReferenceSetItem = target
		response.Result = &refset
		return &response, nil
	}
	c, err := ss.svc.GetConcept(tr.ConceptId)
	if err != nil {
		return nil, err
	}
	members, err := ss.svc.GetReferenceSetItems(tr.TargetId)
	if err != nil {
		return nil, err
	}
	generic, found := ss.svc.GenericiseTo(c, members)
	if found {
		result := snomed.TranslateResponse_Concept{}
		result.Concept = generic
		response.Result = &result
		return &response, nil
	}
	return nil, fmt.Errorf("Unable to translate %d to %d", tr.ConceptId, tr.TargetId)
}

// Subsumes determines whether code A subsumes code B, according to the definition
// in the HL7 FHIR terminology service specification.
// See https://www.hl7.org/fhir/terminology-service.html
func (ss *snomedCTSrv) Subsumes(ctx context.Context, r *snomed.SubsumptionRequest) (*snomed.SubsumptionResponse, error) {
	res := snomed.SubsumptionResponse{}
	if r.CodeA == r.CodeB {
		res.Result = snomed.SubsumptionResponse_EQUIVALENT
		return &res, nil
	}
	c, err := ss.svc.GetConcept(r.CodeB)
	if err != nil {
		return nil, err
	}
	if ss.svc.IsA(c, r.CodeA) {
		res.Result = snomed.SubsumptionResponse_SUBSUMES
		return &res, nil
	}
	c, err = ss.svc.GetConcept(r.CodeA)
	if err != nil {
		return nil, err
	}
	if ss.svc.IsA(c, r.CodeB) {
		res.Result = snomed.SubsumptionResponse_SUBSUMED_BY
		return &res, nil
	}
	res.Result = snomed.SubsumptionResponse_NOT_SUBSUMED
	return &res, nil
}
