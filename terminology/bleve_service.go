package terminology

import (
	"path/filepath"
	"strconv"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"

	// dbq "github.com/blevesearch/bleve/search/query"
	"github.com/wardle/go-terminology/snomed"
	"github.com/wardle/go-terminology/terminology/medicine"
)

type bleveIndexedDocument struct {
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

type bleveService struct {
	path  string
	index bleve.Index
}

func newBleveService(path string, readOnly bool) (*bleveService, error) {
	var index bleve.Index
	var err error
	if !readOnly {

		textMapping := bleve.NewTextFieldMapping()
		textMapping.IncludeInAll = false
		textMapping.Store = false

		boolMapping := bleve.NewBooleanFieldMapping()
		boolMapping.IncludeInAll = false
		boolMapping.Store = false

		storedIDMapping := bleve.NewTextFieldMapping()
		storedIDMapping.IncludeInAll = false
		storedIDMapping.IncludeTermVectors = false
		storedIDMapping.Store = true
		storedIDMapping.Analyzer = keyword.Name

		idMapping := bleve.NewTextFieldMapping()
		idMapping.IncludeInAll = false
		idMapping.IncludeTermVectors = false
		idMapping.Store = false
		idMapping.Analyzer = keyword.Name

		documentMapping := bleve.NewDocumentMapping()
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

		mapping := bleve.NewIndexMapping()
		mapping.StoreDynamic = false
		mapping.DefaultType = "bleveIndexedDocument"
		mapping.AddDocumentMapping("bleveIndexedDocument", documentMapping)

		index, err = bleve.New(filepath.Join(path, "bleve_index"), mapping)
	} else {
		index, err = bleve.OpenUsing(filepath.Join(path, "bleve_index"), map[string]interface{}{
			"read_only": readOnly,
		})
	}
	return &bleveService{path: path, index: index}, err
}

func (bs bleveService) Index(eds []*snomed.ExtendedDescription) error {
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

func (bs bleveService) Search(search *SearchRequest) ([][]int64, error) {

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
	analyzer := mapping.AnalyzerNamed(mapping.AnalyzerNameForPath("Term"))
	tokens := analyzer.Analyze([]byte(search.Search))
	booleanQuery := bleve.NewBooleanQuery()
	for _, token := range tokens {
		tokenString := string(token.Term)

		termQuery := bleve.NewTermQuery(tokenString)
		termQuery.SetField("Term")

		if len(tokenString) >= 3 {
			prefixQuery := bleve.NewPrefixQuery(tokenString)
			prefixQuery.SetField("Term")

			if search.Fuzzy {
				fuzzyQuery := bleve.NewFuzzyQuery(tokenString)
				fuzzyQuery.SetField("Term")
				fuzzyQuery.SetFuzziness(2)
				prefixBooleanQuery := bleve.NewBooleanQuery()
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

	query := bleve.NewConjunctionQuery(booleanQuery)

	for _, refset := range search.Refsets {
		refsetQuery := bleve.NewTermQuery(strconv.FormatInt(refset, 10))
		refsetQuery.SetField("ConceptRefsetIds")
		query.AddQuery(refsetQuery)
	}

	if !search.IncludeInactive {
		isActiveQuery := bleve.NewTermQuery("T")
		isActiveQuery.SetField("ConceptIsActive")
		query.AddQuery(isActiveQuery)
	}

	if len(search.RecursiveParents) > 0 {
		recursiveDisjunctionQuery := bleve.NewDisjunctionQuery()
		for _, recursiveParent := range search.RecursiveParents {
			recursiveParentQuery := bleve.NewTermQuery(strconv.FormatInt(recursiveParent, 10))
			recursiveParentQuery.SetField("RecursiveParentConceptIds")
			recursiveDisjunctionQuery.AddQuery(recursiveParentQuery)
		}
		query.AddQuery(recursiveDisjunctionQuery)
	}

	if len(search.DirectParents) > 0 {
		directDisjunctionQuery := bleve.NewDisjunctionQuery()
		for _, directParent := range search.DirectParents {
			directParentQuery := bleve.NewTermQuery(strconv.FormatInt(directParent, 10))
			directParentQuery.SetField("DirectParentConceptIds")
			directDisjunctionQuery.AddQuery(directParentQuery)
		}
		query.AddQuery(directDisjunctionQuery)
	}

	//dump, _ := dbq.DumpQuery(bs.index.Mapping(), query)
	//print(dump)

	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Size = search.Limit
	searchRequest.Fields = []string{"ConceptId"}

	searchResults, err := bs.index.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	var results [][]int64
	for _, hit := range searchResults.Hits {
		conceptID, _ := strconv.ParseInt(hit.Fields["ConceptId"].(string), 10, 64)
		descriptionID, _ := strconv.ParseInt(hit.ID, 10, 64)
		results = append(results, []int64{conceptID, descriptionID})
	}

	if len(results) == 0 && !search.SuppressFallbackFuzzy && !search.Fuzzy {
		search.Fuzzy = true
		return bs.Search(search)
	}
	//fmt.Printf("%+v\n", results)
	return results, nil
}

func (bs bleveService) ParseMedicationString(medication string) (*medicine.ParsedMedication, error) {
	return medicine.ParseMedicationString(medication), nil
}

func (bs bleveService) Close() error {
	return nil
}
