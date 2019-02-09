package medicine

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/wardle/go-terminology/dmd"
	"github.com/wardle/go-terminology/medicine/storage"
	"github.com/wardle/go-terminology/medicine/storage/goleveldb"
	"github.com/wardle/go-terminology/snomed"
	"github.com/wardle/go-terminology/terminology"
	"golang.org/x/text/language"
)

// Svc encapsulates the terminology, persistent and search services and extends
// it by providing useful helpers for the dm+d
type Svc struct {
	Terminology *terminology.Svc
	storage.Store
}

// Options is a struct used as an argument to medicine.New() for setting an
// alternate path and readOnly state for the search service instead of using
// those specified for the persistence service
type Options struct {
	Index         string
	IndexReadOnly bool
}

// New opens or creates a dm+d service. Requires an instance of terminology.Svs
// and a path to pass to the persistence service
func New(terminologySvc *terminology.Svc, path string, readOnly bool, options ...Options) (*Svc, error) {
	// Creates a new instance of the "goLevelDB" persistence service
	leveldb, err := goleveldb.New(filepath.Join(path, "dmd_store.leveldb"), readOnly)
	if err != nil {
		return nil, err
	}
	/*
		// Creates a new instance of the "BadgerDB" persistence service
		badger, err := badgerdb.New(filepath.Join(path, "dmd_store.badger"), readOnly)
		if err != nil {
			return nil, err
		}

		// Set default options for index and load values from options argument
		var (
			indexPath     = filepath.Join(path, "bleve_index")
			indexReadOnly = readOnly
		)
		if len(options) > 0 {
			indexPath = options[0].Index
			indexReadOnly = options[0].IndexReadOnly
			// Fix path if using default path
			if path == options[0].Index {
				indexPath = filepath.Join(path, "bleve_index")
			}
		}

		// Creat a new instance of the "bleve" search service
		bleve, err := bleve.New(indexPath, indexReadOnly)
		if err != nil {
			return nil, err
		}
	*/

	return &Svc{Terminology: terminologySvc, Store: leveldb}, nil
}

// ParseMedicationString takes a medication string e.g Furosemide 20mg bd po and
// returns a *dmd.ParsedMedication with mapped SNOMED-CT concepts
func (svc *Svc) ParseMedicationString(medicationString string) (*dmd.ParsedMedication, error) {
	parsedMedication := dmd.ParseMedicationString(medicationString)

	var request snomed.SearchRequest
	request.Search = parsedMedication.DrugName
	// request.RecursiveParentIds = []int64{373873005}      // 373873005 | Pharmaceutical / biologic product (product)

	// 10363701000001104 | Virtual therapeutic moiety (product)
	// 10363901000001102 | Actual medicinal product (product)
	// 9191801000001103 | Trade family (product)
	// ^- UK Specific
	request.DirectParentIds = []int64{10363701000001104, 9191801000001103, 10363901000001102}
	request.MaximumHits = 1
	result, err := svc.Terminology.Search.Search(&request)
	if err != nil {
		return &dmd.ParsedMedication{}, err
	}

	if len(result) == 1 {
		description, err := svc.Terminology.GetDescription(result[0])
		if err != nil {
			return &dmd.ParsedMedication{}, err
		}
		concept, err := svc.Terminology.GetConcept(description.ConceptId)
		if err != nil {
			return &dmd.ParsedMedication{}, err
		}
		tags, _, _ := language.ParseAcceptLanguage("en-GB") //using hardcoded language TODO:Fix
		preferredDescription := svc.Terminology.MustGetPreferredSynonym(concept, tags)
		parsedMedication.MappedDrugName = preferredDescription.Term
		parsedMedication.ConceptId = description.ConceptId
		parsedMedication.String_ = parsedMedication.BuildString()
	}
	return parsedMedication, nil
}

func (svc *Svc) PerformImport(importPath string) error {
	version := strings.TrimPrefix(filepath.Base(importPath), "nhsbsa_dmd_")
	if err := svc.PutVersion(version); err != nil {
		return err
	}
	if err := dmd.PerformImport(importPath, svc.Put); err != nil {
		return err
	}

	return svc.Compact()
}

func (svc *Svc) Version() error {
	version, err := svc.GetVersion()
	if err != nil {
		return err
	}
	log.Printf("dm+d version: %s", version)
	return nil
}

func (svc *Svc) GetRoutesForVTM(vtmID int64) ([]int64, error) {
	var output []int64
	outputMap := make(map[int64]bool)

	vtm, err := svc.GetVTM(vtmID)
	if err != nil {
		return output, err
	}

	for _, vmpID := range vtm.VirtualMedicinalProducts {
		vmp, err := svc.GetVMP(vmpID)
		if err != nil {
			return output, err
		}
		if vmp.NonAvailability == dmd.VirtualMedicinalProduct_PRODUCTS_NOT_AVAILABLE {
			continue
		}
		for _, route := range vmp.DrugRoute {
			outputMap[route] = true
		}
	}

	for k := range outputMap {
		output = append(output, k)
	}

	return output, nil
}

func (svc *Svc) GetFormForVTMRoute(vtmID int64, route int64) ([]int64, error) {
	var output []int64
	outputMap := make(map[int64]bool)

	vtm, err := svc.GetVTM(vtmID)
	if err != nil {
		return output, err
	}

	for _, vmpID := range vtm.VirtualMedicinalProducts {
		vmp, err := svc.GetVMP(vmpID)
		if err != nil {
			return output, err
		}
		if vmp.NonAvailability == dmd.VirtualMedicinalProduct_PRODUCTS_NOT_AVAILABLE {
			continue
		}
		for _, vmpRoute := range vmp.DrugRoute {
			if vmpRoute == route {
				for _, form := range vmp.DrugForm {
					outputMap[form] = true
				}
				break
			}
		}
	}

	for k := range outputMap {
		output = append(output, k)
	}

	return output, nil
}

func (svc *Svc) GetTypeForVTMRoute(vtmID int64, route int64) ([]*dmd.MedicationTypeMapping, error) {
	var output []*dmd.MedicationTypeMapping
	outputMap := make(map[int32]*dmd.MedicationTypeMapping)

	forms, err := svc.GetFormForVTMRoute(vtmID, route)
	if err != nil {
		return output, err
	}

	for _, form := range forms {
		medicationType, err := svc.GetMedicationType(form, route)
		if err != nil {
			return output, err
		}
		outputMap[medicationType.Id] = medicationType
	}

	for _, v := range outputMap {
		output = append(output, v)
	}

	return output, nil
}
