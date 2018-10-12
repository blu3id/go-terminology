package bleve

import (
	"strconv"

	blevesearch "github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"

	//dbq "github.com/blevesearch/bleve/search/query"
	"github.com/wardle/go-terminology/snomed"
	"github.com/wardle/go-terminology/terminology/search"
)

// bleveIndexedDocument is a struct defining the document indexed by Bleve
type bleveIndexedDocument struct {
	SortWeight                string
	Term                      string
	PreferredTerm             string
	ConceptId                 string
	RecursiveParentConceptIds []string
	DirectParentConceptIds    []string
	Language                  string
	DescriptionIsActive       bool
	ConceptIsActive           bool
	DescriptionId             string
	DescriptionType           string
	ModuleId                  string
	ConceptRefsetIds          []string
	DescriptionRefsetIds      []string
}

// bleveService is a search service for SNOMED-CT that implements the search.Search interface
type bleveService struct {
	index blevesearch.Index
	_     search.Search
}

func New(path string, readOnly bool) (search.Search, error) {
	var index blevesearch.Index
	var err error
	if !readOnly {

		textMapping := blevesearch.NewTextFieldMapping()
		textMapping.IncludeInAll = false
		textMapping.Store = false

		boolMapping := blevesearch.NewBooleanFieldMapping()
		boolMapping.IncludeInAll = false
		boolMapping.Store = false

		storedIDMapping := blevesearch.NewTextFieldMapping()
		storedIDMapping.IncludeInAll = false
		storedIDMapping.IncludeTermVectors = false
		storedIDMapping.Store = true
		storedIDMapping.Analyzer = keyword.Name

		idMapping := blevesearch.NewTextFieldMapping()
		idMapping.IncludeInAll = false
		idMapping.IncludeTermVectors = false
		idMapping.Store = false
		idMapping.Analyzer = keyword.Name

		documentMapping := blevesearch.NewDocumentMapping()
		documentMapping.AddFieldMappingsAt("Term", textMapping)
		documentMapping.AddFieldMappingsAt("PreferredTerm", textMapping)
		documentMapping.AddFieldMappingsAt("ConceptId", storedIDMapping)
		documentMapping.AddFieldMappingsAt("RecursiveParentConceptIds", idMapping)
		documentMapping.AddFieldMappingsAt("DirectParentConceptIds", idMapping)
		documentMapping.AddFieldMappingsAt("Language", idMapping)
		documentMapping.AddFieldMappingsAt("DescriptionIsActive", boolMapping)
		documentMapping.AddFieldMappingsAt("ConceptIsActive", boolMapping)
		documentMapping.AddFieldMappingsAt("DescriptionId", idMapping)
		documentMapping.AddFieldMappingsAt("DescriptionType", idMapping)
		documentMapping.AddFieldMappingsAt("ModuleId", idMapping)
		documentMapping.AddFieldMappingsAt("ConceptRefsetIds", idMapping)
		documentMapping.AddFieldMappingsAt("DescriptionRefsetIds", idMapping)

		mapping := blevesearch.NewIndexMapping()
		mapping.StoreDynamic = false
		mapping.DefaultType = "bleveIndexedDocument"
		mapping.AddDocumentMapping("bleveIndexedDocument", documentMapping)

		/*
			//moss index - fast indexing
			kvconfig := map[string]interface{}{
				"mossLowerLevelStoreName": "mossStore",
			}
			index, err = blevesearch.NewUsing(path, mapping, upsidedown.Name, moss.Name, kvconfig)
		*/

		/*
			//goleveldb index - space efficient as slow as bolt indexing TODO: Optimise compaction with options
			index, err = blevesearch.NewUsing(path, mapping, upsidedown.Name, goleveldb.Name, map[string]interface{}{})
		*/

		//bolt index (default) - space ineficient, slow indexing
		index, err = blevesearch.New(path, mapping)
	} else {
		index, err = blevesearch.OpenUsing(path, map[string]interface{}{
			"read_only": readOnly,
		})
	}
	return &bleveService{index: index}, err
}

func (bs *bleveService) Index(eds []*snomed.ExtendedDescription) error {
	batch := bs.index.NewBatch()

	for _, ed := range eds {
		var doc bleveIndexedDocument

		//Convert int64 to string as better efficiency in Bleve index as we aren't going to be doing range queries
		doc.Term = ed.Description.Term
		doc.PreferredTerm = ed.PreferredDescription.Term
		doc.ConceptId = strconv.FormatInt(ed.Concept.Id, 10)
		doc.Language = ed.Description.LanguageCode
		doc.DescriptionIsActive = ed.Description.Active
		doc.ConceptIsActive = ed.Concept.Active
		doc.DescriptionId = strconv.FormatInt(ed.Description.Id, 10)
		doc.DescriptionType = strconv.FormatInt(ed.Description.TypeId, 10)
		doc.ModuleId = strconv.FormatInt(ed.Description.ModuleId, 10)

		for _, v := range ed.RecursiveParentIds {
			doc.RecursiveParentConceptIds = append(doc.RecursiveParentConceptIds, strconv.FormatInt(v, 10))
		}
		for _, v := range ed.DirectParentIds {
			doc.DirectParentConceptIds = append(doc.DirectParentConceptIds, strconv.FormatInt(v, 10))
		}
		for _, v := range ed.ConceptRefsets {
			doc.ConceptRefsetIds = append(doc.ConceptRefsetIds, strconv.FormatInt(v, 10))
		}
		for _, v := range ed.DescriptionRefsets {
			doc.DescriptionRefsetIds = append(doc.DescriptionRefsetIds, strconv.FormatInt(v, 10))
		}

		err := batch.Index(doc.DescriptionId, doc)
		//fmt.Printf("%+v\n", doc)
		if err != nil {
			return err
		}
	}

	err := bs.index.Batch(batch)
	return err
}

func (bs *bleveService) Search(search *snomed.SearchRequest) ([]int64, error) {
	/*
		// SearchRequest permits an arbitrary free-text search of the hierarchy.
		message SearchRequest {
			string search  = 1; // the search string
			repeated int64 recursive_parent_ids = 2; 	// limit search to descendents of these parents
			repeated int64 direct_parent_ids = 3; 		// limit search to direct descendents of these parents
			repeated int64 reference_set_ids = 4; 		// limit search to members of the specified reference sets
			int32 maximum_hits = 5;						// limit for maximum hits
			bool include_inactive = 6;					// search inactive terms, default false
			Fuzzy fuzzy = 7;  							// fuzziness preference
			string accepted_languages = 8;				// accepted languages, formatted as per https://tools.ietf.org/html/rfc7231#section-5.3.5

			enum Fuzzy {
				FALLBACK_FUZZY = 0;			// try a fuzzy match only if there are no results without using fuzzy
				ALWAYS_FUZZY = 1; 			// use fuzzy for the search
				NO_FUZZY = 2;				// do not use fuzzy matching at all
			}
		}
	*/

	if len(search.RecursiveParentIds) == 0 {
		search.RecursiveParentIds = []int64{138875005}
	}

	if search.MaximumHits == 0 {
		search.MaximumHits = 200
	}

	mapping := bs.index.Mapping()
	analyzer := mapping.AnalyzerNamed(mapping.AnalyzerNameForPath("Term"))
	tokens := analyzer.Analyze([]byte(search.Search))
	booleanQuery := blevesearch.NewBooleanQuery()
	for _, token := range tokens {
		tokenString := string(token.Term)

		termQuery := blevesearch.NewTermQuery(tokenString)
		termQuery.SetField("Term")

		if len(tokenString) >= 3 {
			prefixQuery := blevesearch.NewPrefixQuery(tokenString)
			prefixQuery.SetField("Term")

			if search.Fuzzy == snomed.SearchRequest_ALWAYS_FUZZY {
				fuzzyQuery := blevesearch.NewFuzzyQuery(tokenString)
				fuzzyQuery.SetField("Term")
				fuzzyQuery.SetFuzziness(2)
				prefixBooleanQuery := blevesearch.NewBooleanQuery()
				prefixBooleanQuery.AddShould(prefixQuery)
				prefixBooleanQuery.AddShould(fuzzyQuery)
				booleanQuery.AddMust(prefixBooleanQuery)
			} else {
				booleanQuery.AddMust(prefixQuery)
			}
		} else {
			booleanQuery.AddMust(termQuery)
		}
	}

	//Exclude FSN (no current option to disable in snomed.SearchRequest)
	excludeFSNQuery := blevesearch.NewTermQuery("900000000000003001")
	excludeFSNQuery.SetField("DescriptionType")
	booleanQuery.AddMustNot(excludeFSNQuery)

	query := blevesearch.NewConjunctionQuery(booleanQuery)

	for _, refset := range search.ReferenceSetIds {
		refsetQuery := blevesearch.NewTermQuery(strconv.FormatInt(refset, 10))
		refsetQuery.SetField("ConceptRefsetIds")
		query.AddQuery(refsetQuery)
	}

	if !search.IncludeInactive {
		isActiveQuery := blevesearch.NewTermQuery("T")
		isActiveQuery.SetField("ConceptIsActive")
		query.AddQuery(isActiveQuery)
	}

	if len(search.RecursiveParentIds) > 0 {
		recursiveDisjunctionQuery := blevesearch.NewDisjunctionQuery()
		for _, recursiveParent := range search.RecursiveParentIds {
			recursiveParentQuery := blevesearch.NewTermQuery(strconv.FormatInt(recursiveParent, 10))
			recursiveParentQuery.SetField("RecursiveParentConceptIds")
			recursiveDisjunctionQuery.AddQuery(recursiveParentQuery)
		}
		query.AddQuery(recursiveDisjunctionQuery)
	}

	if len(search.DirectParentIds) > 0 {
		directDisjunctionQuery := blevesearch.NewDisjunctionQuery()
		for _, directParent := range search.DirectParentIds {
			directParentQuery := blevesearch.NewTermQuery(strconv.FormatInt(directParent, 10))
			directParentQuery.SetField("DirectParentConceptIds")
			directDisjunctionQuery.AddQuery(directParentQuery)
		}
		query.AddQuery(directDisjunctionQuery)
	}

	//dump, _ := dbq.DumpQuery(bs.index.Mapping(), query)
	//print(dump)

	searchRequest := blevesearch.NewSearchRequest(query)
	searchRequest.Size = int(search.MaximumHits)
	searchRequest.Fields = []string{"ConceptId"}

	searchResults, err := bs.index.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	var results []int64
	for _, hit := range searchResults.Hits {
		//conceptID, _ := strconv.ParseInt(hit.Fields["ConceptId"].(string), 10, 64)
		descriptionID, _ := strconv.ParseInt(hit.ID, 10, 64)
		results = append(results, descriptionID)
	}

	if (len(results) == 0) && (search.Fuzzy != snomed.SearchRequest_ALWAYS_FUZZY) && (search.Fuzzy != snomed.SearchRequest_NO_FUZZY) {
		search.Fuzzy = snomed.SearchRequest_ALWAYS_FUZZY
		return bs.Search(search)
	}
	//fmt.Printf("%+v\n", results)
	return results, nil
}

func (bs *bleveService) Close() error {
	return bs.index.Close()
}
