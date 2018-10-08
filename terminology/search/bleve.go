package bleve

import (
	"fmt"
	"path/filepath"
	"strconv"

	blevesearch "github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/index/store/moss"
	"github.com/blevesearch/bleve/index/upsidedown"

	//dbq "github.com/blevesearch/bleve/search/query"
	"github.com/wardle/go-terminology/snomed"
	"github.com/wardle/go-terminology/terminology/interfaces"
	"github.com/wardle/go-terminology/terminology/medicine"
)

type bleveIndexedDocument struct {
	ConceptId                 string
	ConceptIsActive           bool
	PreferredTerm             string
	DirectParentConceptIds    []string
	ConceptRefsetIds          []string
	RecursiveParentConceptIds []string
	Descriptions              []bleveIndexedDescriptions
}

type bleveIndexedDescriptions struct {
	SortWeight           string
	DescriptionId        string
	DescriptionIsActive  bool
	Term                 string
	Language             string
	DescriptionType      string
	ModuleId             string
	DescriptionRefsetIds []string
}

type bleveService struct {
	index blevesearch.Index
	_     interfaces.Search
}

func New(path string, readOnly bool) (interfaces.Search, error) {
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

		SubDocumentMapping := blevesearch.NewDocumentMapping()
		SubDocumentMapping.AddFieldMappingsAt("SortWeight", idMapping)
		SubDocumentMapping.AddFieldMappingsAt("DescriptionId", storedIDMapping)
		SubDocumentMapping.AddFieldMappingsAt("DescriptionIsActive", boolMapping)
		SubDocumentMapping.AddFieldMappingsAt("Term", textMapping)
		SubDocumentMapping.AddFieldMappingsAt("Language", idMapping)
		SubDocumentMapping.AddFieldMappingsAt("DescriptionType", idMapping)
		SubDocumentMapping.AddFieldMappingsAt("ModuleId", idMapping)
		SubDocumentMapping.AddFieldMappingsAt("DescriptionRefsetIds", idMapping)

		documentMapping := blevesearch.NewDocumentMapping()
		documentMapping.AddFieldMappingsAt("ConceptIsActive", boolMapping)
		documentMapping.AddFieldMappingsAt("PreferredTerm", textMapping)
		documentMapping.AddFieldMappingsAt("DirectParentConceptIds", idMapping)
		documentMapping.AddFieldMappingsAt("ConceptRefsetIds", idMapping)
		documentMapping.AddFieldMappingsAt("RecursiveParentConceptIds", idMapping)
		documentMapping.AddSubDocumentMapping("Descriptions", SubDocumentMapping)

		mapping := blevesearch.NewIndexMapping()
		mapping.StoreDynamic = false
		mapping.DefaultType = "bleveIndexedDocument"
		mapping.AddDocumentMapping("bleveIndexedDocument", documentMapping)
		kvconfig := map[string]interface{}{
			"mossLowerLevelStoreName": "mossStore",
		}

		index, err = blevesearch.NewUsing(filepath.Join(path, "bleve_index"), mapping, upsidedown.Name, moss.Name, kvconfig)
	} else {
		index, err = blevesearch.OpenUsing(filepath.Join(path, "bleve_index"), map[string]interface{}{
			"read_only": readOnly,
		})
	}
	return &bleveService{index: index}, err
}

func (bs *bleveService) Index(ics []*snomed.IndexedConcept) error {
	batch := bs.index.NewBatch()
	mapping := bs.index.Mapping()
	analyzer := mapping.AnalyzerNamed(mapping.AnalyzerNameForPath("Descriptions.Term"))

	for _, ic := range ics {
		var doc bleveIndexedDocument
		//Convert int64 to string as better efficiency in Bleve index as we aren't going to be doing range queries
		doc.ConceptId = strconv.FormatInt(ic.Concept.Id, 10)
		doc.ConceptIsActive = ic.Concept.Active
		doc.PreferredTerm = ic.PreferredDescription.Term

		for _, v := range ic.DirectParentIds {
			doc.DirectParentConceptIds = append(doc.DirectParentConceptIds, strconv.FormatInt(v, 10))
		}
		for _, v := range ic.ConceptRefsets {
			doc.ConceptRefsetIds = append(doc.ConceptRefsetIds, strconv.FormatInt(v, 10))
		}
		for _, v := range ic.RecursiveParentIds {
			doc.RecursiveParentConceptIds = append(doc.RecursiveParentConceptIds, strconv.FormatInt(v, 10))
		}

		for _, v := range ic.Descriptions {
			var desc bleveIndexedDescriptions

			desc.SortWeight = fmt.Sprintf("%02d", len(analyzer.Analyze([]byte(v.Description.Term))))
			desc.DescriptionId = strconv.FormatInt(v.Description.Id, 10)
			desc.DescriptionIsActive = v.Description.Active
			desc.Term = v.Description.Term
			desc.Language = v.Description.LanguageCode
			desc.DescriptionType = strconv.FormatInt(v.Description.TypeId, 10)
			desc.ModuleId = strconv.FormatInt(v.Description.ModuleId, 10)

			for _, v := range v.DescriptionRefsets {
				desc.DescriptionRefsetIds = append(desc.DescriptionRefsetIds, strconv.FormatInt(v, 10))
			}

			doc.Descriptions = append(doc.Descriptions, desc)
		}

		err := batch.Index(doc.ConceptId, doc)
		//fmt.Printf("%+v\n", doc)
		if err != nil {
			return err
		}
	}

	err := bs.index.Batch(batch)
	return err
}

func (bs *bleveService) Search(search *interfaces.SearchRequest) ([][]int64, error) {

	/*type SearchRequest struct {
	Search                string  `schema:"s"`                     // search term
	RecursiveParents      []int64 `schema:"root"`                  // one or more root concept identifiers (default 138875005)
	DirectParents         []int64 `schema:"is"`                    // zero or more direct parent concept identifiers
	Refsets               []int64 `schema:"refset"`                // filter to concepts within zero of more refsets
	Limit                 int     `schema:"maxHits"`               // number of hits (default 200)
	IncludeInactive       bool    `schema:"inactive"`              // whether to include inactive terms in search results (defaults to False)
	Fuzzy                 bool    `schema:"fuzzy"`                 // whether to use a fuzzy search for search (default to False)
	SuppressFallbackFuzzy bool    `schema:"suppressFuzzyFallback"` // whether to suppress automatic fallback to fuzzy search if no results found for non-fuzzy search (defaults to False)
	}*/

	if len(search.RecursiveParents) == 0 {
		search.RecursiveParents = []int64{138875005}
	}

	if search.Limit == 0 {
		search.Limit = 200
	}

	mapping := bs.index.Mapping()
	analyzer := mapping.AnalyzerNamed(mapping.AnalyzerNameForPath("Descriptions.Term"))
	tokens := analyzer.Analyze([]byte(search.Search))
	booleanQuery := blevesearch.NewBooleanQuery()
	for _, token := range tokens {
		tokenString := string(token.Term)

		termQuery := blevesearch.NewTermQuery(tokenString)
		termQuery.SetField("Descriptions.Term")

		if len(tokenString) >= 3 {
			prefixQuery := blevesearch.NewPrefixQuery(tokenString)
			prefixQuery.SetField("Descriptions.Term")

			if search.Fuzzy {
				fuzzyQuery := blevesearch.NewFuzzyQuery(tokenString)
				fuzzyQuery.SetField("Descriptions.Term")
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

	//if !search.IncludeFSN {
	//excludeFSNQuery := blevesearch.NewTermQuery("900000000000003001")
	//excludeFSNQuery.SetField("Descriptions.DescriptionType")
	//booleanQuery.AddMustNot(excludeFSNQuery)
	//}

	query := blevesearch.NewConjunctionQuery(booleanQuery)

	for _, refset := range search.Refsets {
		refsetQuery := blevesearch.NewTermQuery(strconv.FormatInt(refset, 10))
		refsetQuery.SetField("ConceptRefsetIds")
		query.AddQuery(refsetQuery)
	}

	if !search.IncludeInactive {
		isActiveQuery := blevesearch.NewTermQuery("T")
		isActiveQuery.SetField("ConceptIsActive")
		query.AddQuery(isActiveQuery)
	}

	if len(search.RecursiveParents) > 0 {
		recursiveDisjunctionQuery := blevesearch.NewDisjunctionQuery()
		for _, recursiveParent := range search.RecursiveParents {
			recursiveParentQuery := blevesearch.NewTermQuery(strconv.FormatInt(recursiveParent, 10))
			recursiveParentQuery.SetField("RecursiveParentConceptIds")
			recursiveDisjunctionQuery.AddQuery(recursiveParentQuery)
		}
		query.AddQuery(recursiveDisjunctionQuery)
	}

	if len(search.DirectParents) > 0 {
		directDisjunctionQuery := blevesearch.NewDisjunctionQuery()
		for _, directParent := range search.DirectParents {
			directParentQuery := blevesearch.NewTermQuery(strconv.FormatInt(directParent, 10))
			directParentQuery.SetField("DirectParentConceptIds")
			directDisjunctionQuery.AddQuery(directParentQuery)
		}
		query.AddQuery(directDisjunctionQuery)
	}

	//dump, _ := dbq.DumpQuery(bs.index.Mapping(), query)
	//print(dump)

	searchRequest := blevesearch.NewSearchRequest(query)
	searchRequest.Size = search.Limit
	searchRequest.Fields = []string{"Descriptions.DescriptionId", "Descriptions.Term", "Term"}
	searchRequest.SortBy([]string{"Descriptions.SortWeight"})

	searchResults, err := bs.index.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	var results [][]int64
	for _, hit := range searchResults.Hits {
		conceptID, _ := strconv.ParseInt(hit.ID, 10, 64)
		descriptionID, _ := strconv.ParseInt(hit.Fields["Descriptions.DescriptionId"].([]interface{})[0].(string), 10, 64)
		fmt.Printf("%+v\n", hit.Locations)
		for k, v := range hit.Fields {
			fmt.Printf("Field %v = %v\n", k, v)
		}
		for fieldName, fieldMap := range hit.Locations {
			for termName, locations := range fieldMap {
				for _, location := range locations {
					fmt.Printf("    Field %s has term %s from %d to %d (Pos %d?)\n", fieldName, termName, location.Start, location.Size(), location.ArrayPositions)
				}
			}
		}

		results = append(results, []int64{conceptID, descriptionID})
	}

	if len(results) == 0 && !search.SuppressFallbackFuzzy && !search.Fuzzy {
		search.Fuzzy = true
		return bs.Search(search)
	}
	//fmt.Printf("%+v\n", results)
	return results, nil
}

func (bs *bleveService) ParseMedicationString(medication string) (*medicine.ParsedMedication, error) {
	return medicine.ParseMedicationString(medication), nil
}

func (bs *bleveService) Close() error {
	return nil
}
