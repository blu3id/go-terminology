package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"

	//"github.com/google/uuid"
	"github.com/schollz/progressbar/v3"
	"github.com/wardle/go-terminology/snomed"
	"github.com/wardle/go-terminology/terminology"
)

// fileType represents a type of SNOMED-CT distribution file
type fileType int

// Supported file types
// These are listed in order of importance for import
const (
	conceptsFileType fileType = iota
	descriptionsFileType
	relationshipsFileType
	refsetDescriptorRefsetFileType
	languageRefsetFileType
	simpleRefsetFileType
	simpleMapRefsetFileType
	extendedMapRefsetFileType
	complexMapRefsetFileType
	attributeValueRefsetFileType
	associationRefsetFileType
	lastFileType
)

var fileTypeNames = []string{
	"Concepts",
	"Descriptions",
	"Relationships",
	"Refset Descriptor refset",
	"Language refset",
	"Simple refset",
	"Simple map refset",
	"Extended map refset",
	"Complex map refset",
	"Attribute value refset",
	"Association refset",
}
var fileTypeColumnNames = [][]string{
	{"id", "effectiveTime", "active", "moduleId", "definitionStatusId"},
	{"id", "effectiveTime", "active", "moduleId", "conceptId", "languageCode", "typeId", "term", "caseSignificanceId"},
	{"id", "effectiveTime", "active", "moduleId", "sourceId", "destinationId", "relationshipGroup", "typeId", "characteristicTypeId", "modifierId"},
	{"id", "effectiveTime", "active", "moduleId", "refsetId", "referencedComponentId", "attributeDescription", "attributeType", "attributeOrder"},
	{"id", "effectiveTime", "active", "moduleId", "refsetId", "referencedComponentId", "acceptabilityId"},
	{"id", "effectiveTime", "active", "moduleId", "refsetId", "referencedComponentId"},
	{"id", "effectiveTime", "active", "moduleId", "refsetId", "referencedComponentId", "mapTarget"},
	{"id", "effectiveTime", "active", "moduleId", "refsetId", "referencedComponentId", "mapGroup", "mapPriority", "mapRule", "mapAdvice", "mapTarget", "correlationId", "mapCategoryId"},
	{"id", "effectiveTime", "active", "moduleId", "refsetId", "referencedComponentId", "mapGroup", "mapPriority", "mapRule", "mapAdvice", "mapTarget", "correlationId", "mapBlock"},
	{"id", "effectiveTime", "active", "moduleId", "refsetId", "referencedComponentId", "valueId"},
	{"id", "effectiveTime", "active", "moduleId", "refsetId", "referencedComponentId", "targetComponentId"},
}

// Filename patterns for the supported file types
var fileTypeFilenamePatterns = []*regexp.Regexp{
	regexp.MustCompile("sct2_Concept_Snapshot_\\S+_\\S+.txt"),
	regexp.MustCompile("sct2_Description_Snapshot-en\\S+_\\S+.txt"),
	regexp.MustCompile("sct2_(Stated)*Relationship_Snapshot_\\S+_\\S+.txt"),
	regexp.MustCompile("der2_cciRefset_RefsetDescriptorSnapshot_\\S+_\\S+.txt"),
	regexp.MustCompile("der2_cRefset_LanguageSnapshot-\\S+_\\S+.txt"),
	regexp.MustCompile("der2_Refset_SimpleSnapshot_\\S+_\\S+.txt"),
	regexp.MustCompile("der2_sRefset_SimpleMapSnapshot_\\S+_\\S+.txt"),
	regexp.MustCompile("der2_iisssccRefset_ExtendedMapSnapshot_\\S+_\\S+.txt"), // extended
	regexp.MustCompile("der2_iisssciRefset_ExtendedMapSnapshot_\\S+_\\S+.txt"), // complex
	regexp.MustCompile("der2_cRefset_AttributeValueSnapshot_\\S+_\\S+.txt"),
	regexp.MustCompile("der2_cRefset_AssociationSnapshot_\\S+_\\S+.txt"),
}

// Regexp returns the filename pattern for this file type
func (ft fileType) Regexp() *regexp.Regexp {
	return fileTypeFilenamePatterns[ft]
}

// Column returns the expected column names for this file type
func (ft fileType) Columns() []string {
	return fileTypeColumnNames[ft]
}

// String returns the human readable name for the file type
func (ft fileType) String() string {
	return fileTypeNames[ft]
}

// findImportableFiles is a helper function that walks the specified `rootPath`
// and attempts to match the files against the `Regexp` for all defined
// `fileType`
func findImportableFiles(rootPath string) (importable map[fileType][]string, totalSize int64, err error) {
	importable = make(map[fileType][]string)

	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		file := filepath.Base(path)
		for ft := conceptsFileType; ft < lastFileType; ft++ {
			if ft.Regexp().MatchString(file) {
				importable[ft] = append(importable[ft], path)
				totalSize += info.Size()
				break
			}
		}

		return nil
	})

	return
}

func runImport2(svc *terminology.Svc, batchSize int, verbose bool) {
	start := time.Now()
	importRoot := "_REF2_data_files"
	imports, totalSize, err := findImportableFiles(importRoot)
	if err != nil {
		log.Printf("Error finding importable files at root %q: %v\n", importRoot, err)
		return
	}

	var writerWg sync.WaitGroup
	parsedComponent := make(chan interface{}, 2)
	writerWg.Add(1)
	go batchWriter(parsedComponent, svc, batchSize, &writerWg)

	progressWriter := progressbar.NewOptions64(
		totalSize,
		progressbar.OptionSetDescription("Importing"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(50),
		progressbar.OptionThrottle(500*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionSpinnerType(14),
	)
	progressWriter.RenderBlank()

	// Inverse order of files is faster to import due to lower frequency of
	// sequential records in initial filetypes which is lest costly to insert
	// into an empty LeveDB.
	//
	// There is no need to use Goroutines to parallelise the reading of import
	// files as this is io bound rather than CPU bound. Performance testing the
	// import code without parsing and writing to the datastore shows that this
	// is not a bottleneck.
	//
	// for ft := conceptsFileType; ft < lastFileType; ft++ {
	for ft := lastFileType; ft > -1; ft-- {
		if paths, ok := imports[ft]; ok {
			progressWriter.Describe(fmt.Sprintf("Importing: %s [0/%d]", ft, len(paths)))
			i := 0
			for _, path := range paths {
				i++
				if verbose {
					progressWriter.Clear()
					fmt.Printf("%s [%d/%d]: %s\n", ft, i, len(paths), path)
				}
				progressWriter.Describe(fmt.Sprintf("Importing: %s [%d/%d]", ft, i, len(paths)))
				readFile(ft, path, parsedComponent, progressWriter)
			}
		}
	}

	close(parsedComponent)
	writerWg.Wait()
	log.Printf("done in %s", time.Since(start))
}

// readFile reads the file specified and parses it using `parseByFileType’ using
// the specified `filetype` and writes the snomed structures to the
// `parsedComponent` channel (which is consumed by the `batchWriter` Goroutine).
func readFile(ft fileType, path string, parsedComponent chan interface{}, progressWriter io.Writer) {
	f, err := os.Open(path)
	if err != nil {
		log.Panicf("unable to process file %q: %s", path, err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(io.TeeReader(f, progressWriter))

	// read the first line and check that we have the right column names
	if scanner.Scan() == false {
		log.Panicf("empty file %s", path)
	}
	fileHeadings := strings.Split(scanner.Text(), "\t")
	if !reflect.DeepEqual(fileHeadings, ft.Columns()) {
		log.Panicf("expecting column names: %v, got: %v", ft.Columns(), fileHeadings)
	}

	// The following is an experiment to load complete Refset with
	// non-sequential entries into memory and sort before importing into LeveDB.
	// Significant performance improvement but also significant memory cost that
	// may not be suitable for all users and isn't configurable by tweaking
	// batch size.
	/*
		// Largest most non-sequential files (could include more Refsets)
		if ft == complexMapRefsetFileType || ft == simpleRefsetFileType {
			type RowMap struct {
				Key []byte
				Row []byte
			}
			var rowMap []RowMap

			for scanner.Scan() {
				line := scanner.Bytes()
				lineData := append([]byte{}, line...)
				column := bytes.Split(line, []byte{'\t'})
				key, _ := uuid.MustParse(string(column[0])).MarshalBinary()
				rowMap = append(rowMap, RowMap{key, lineData})
			}

			sort.Slice(rowMap, func(i, j int) bool {
				return bytes.Compare(rowMap[i].Key, rowMap[j].Key) < 0
			})

			for _, mappedline := range rowMap {
				sni, _ := parseByFileType(ft, mappedline.Row)
				parsedComponent <- sni
			}

			return
		}
	*/

	for scanner.Scan() {
		line := scanner.Bytes()
		sni, _ := parseByFileType(ft, line)
		parsedComponent <- sni
		//switch sni.(type) {
		//case *snomed.Concept:
		//case *snomed.Description:
		//case *snomed.Relationship:
		//case *snomed.ReferenceSetItem:
		//default:
		//	log.Panic("unknow snomed type")
		//}
	}
}

// batchWriter is a Goroutine that listens to the parsedComponent channel. It
// batches the snomed structures written to the channel by the filetype parsers
// into `batchSize` before writing them to the datastore (through `pooledPut`).
// This enables records to remain in sequence which is important for LevelDB
// writing performance.
func batchWriter(parsedComponent <-chan interface{}, svc *terminology.Svc, batchSize int, writerWg *sync.WaitGroup) {
	var (
		concepts          []*snomed.Concept
		descriptions      []*snomed.Description
		relationships     []*snomed.Relationship
		referencesetitems []*snomed.ReferenceSetItem
		wg                sync.WaitGroup
	)

	defer writerWg.Done()

	// Set maximum number in pool thorugh buffer size of control channel and
	// initial messages on channel
	control := make(chan struct{}, runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		control <- struct{}{}
	}

	for {
		m, ok := <-parsedComponent

		if !ok {
			pooledPut(concepts, svc, control, &wg)
			pooledPut(descriptions, svc, control, &wg)
			pooledPut(relationships, svc, control, &wg)
			pooledPut(referencesetitems, svc, control, &wg)
			break
		}

		switch m.(type) {
		case *snomed.Concept:
			concepts = append(concepts, m.(*snomed.Concept))
			if len(concepts) > batchSize {
				pooledPut(concepts, svc, control, &wg)
				concepts = []*snomed.Concept{}
			}
		case *snomed.Description:
			descriptions = append(descriptions, m.(*snomed.Description))
			if len(descriptions) > batchSize {
				pooledPut(descriptions, svc, control, &wg)
				descriptions = []*snomed.Description{}
			}
		case *snomed.Relationship:
			relationships = append(relationships, m.(*snomed.Relationship))
			if len(relationships) > batchSize {
				pooledPut(relationships, svc, control, &wg)
				relationships = []*snomed.Relationship{}
			}
		case *snomed.ReferenceSetItem:
			referencesetitems = append(referencesetitems, m.(*snomed.ReferenceSetItem))
			if len(referencesetitems) > batchSize {
				pooledPut(referencesetitems, svc, control, &wg)
				referencesetitems = []*snomed.ReferenceSetItem{}
			}
		default:
			log.Panic("unknow snomed type")
		}
	}

	wg.Wait()
}

// pooledPut is a wrapper of the *terminology.Svc Put command. This enables
// spawning of a pool of Goroutine workers for the computationaly expensive
// sorting of records in each batch and the LevelDB writes (compressing and
// (re-)sorting of blocks)
func pooledPut(items interface{}, svc *terminology.Svc, control chan struct{}, wg *sync.WaitGroup) {
	// Block waiting for recive from control channel before spawning Goroutine
	<-control

	wg.Add(1)
	go func(items interface{}, svc *terminology.Svc, control chan struct{}, wg *sync.WaitGroup) {
		defer wg.Done()
		sort.Slice(items, func(i, j int) bool {
			switch items.(type) {
			case []*snomed.Concept:
				return items.([]*snomed.Concept)[i].Id < items.([]*snomed.Concept)[j].Id
			case []*snomed.Description:
				return items.([]*snomed.Description)[i].Id < items.([]*snomed.Description)[j].Id
			case []*snomed.Relationship:
				return items.([]*snomed.Relationship)[i].Id < items.([]*snomed.Relationship)[j].Id
			case []*snomed.ReferenceSetItem:
				return items.([]*snomed.ReferenceSetItem)[i].Id < items.([]*snomed.ReferenceSetItem)[j].Id
			default:
				log.Panic("unknow snomed type")
				return false
			}
		})
		ctx := context.Background()
		svc.Put(ctx, items)

		// Send to control channel to signal Goroutine is ending to enable
		// spawning of another Gouroutine
		control <- struct{}{}

	}(items, svc, control, wg)
}

// parseByFileType selects the appropriate parser for the each row of `[]byte`
// based on the specified `fileType` and returns a snomed component
func parseByFileType(ft fileType, row []byte) (interface{}, error) {
	switch ft {
	case conceptsFileType:
		return parseConcept(row)
	case descriptionsFileType:
		return parseDescription(row)
	case relationshipsFileType:
		return parseRelationship(row)
	case refsetDescriptorRefsetFileType:
		return parseRefsetDescriptorRefset(row)
	case languageRefsetFileType:
		return parseLanguageRefset(row)
	case simpleRefsetFileType:
		return parseSimpleRefset(row)
	case simpleMapRefsetFileType:
		return parseSimpleMapRefset(row)
	case extendedMapRefsetFileType:
		return parseExtendedMapRefset(row)
	case complexMapRefsetFileType:
		return parseComplexMapRefset(row)
	case attributeValueRefsetFileType:
		return parseAttributeValueRefset(row)
	case associationRefsetFileType:
		return parseAssociationRefset(row)
	default:
		return nil, fmt.Errorf("unable to process filetype %s", ft)
	}
}

func parseIdentifier(b []byte) int64 {
	return parseInt(b)
}

func parseInt(b []byte) int64 {
	i, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		log.Panicln("Error parsing file,", err)
	}
	return i
}

func parseBoolean(b []byte) bool {
	r, err := strconv.ParseBool(string(b))
	if err != nil {
		log.Panicln("Error parsing file,", err)
	}
	return r
}
func parseDate(b []byte) *timestamp.Timestamp {
	t, err := time.Parse("20060102", string(b))
	if err != nil {
		log.Panicln("Error parsing file,", err)
	}
	tp, err := ptypes.TimestampProto(t)
	if err != nil {
		log.Panicln("Error parsing file,", err)
	}
	return tp
}

func parseConcept(row []byte) (concept *snomed.Concept, err error) {
	column := bytes.Split(row, []byte{'\t'})
	return &snomed.Concept{
		Id:                 parseIdentifier(column[0]),
		EffectiveTime:      parseDate(column[1]),
		Active:             parseBoolean(column[2]),
		ModuleId:           parseIdentifier(column[3]),
		DefinitionStatusId: parseIdentifier(column[4]),
	}, nil
}

func parseDescription(row []byte) (*snomed.Description, error) {
	column := bytes.Split(row, []byte{'\t'})
	return &snomed.Description{
		Id:               parseIdentifier(column[0]),
		EffectiveTime:    parseDate(column[1]),
		Active:           parseBoolean(column[2]),
		ModuleId:         parseIdentifier(column[3]),
		ConceptId:        parseIdentifier(column[4]),
		LanguageCode:     string(column[5]),
		TypeId:           parseIdentifier(column[6]),
		Term:             string(column[7]),
		CaseSignificance: parseIdentifier(column[8]),
	}, nil
}

func parseRelationship(row []byte) (*snomed.Relationship, error) {
	column := bytes.Split(row, []byte{'\t'})
	return &snomed.Relationship{
		Id:                   parseIdentifier(column[0]),
		EffectiveTime:        parseDate(column[1]),
		Active:               parseBoolean(column[2]),
		ModuleId:             parseIdentifier(column[3]),
		SourceId:             parseIdentifier(column[4]),
		DestinationId:        parseIdentifier(column[5]),
		RelationshipGroup:    parseInt(column[6]),
		TypeId:               parseIdentifier(column[7]),
		CharacteristicTypeId: parseIdentifier(column[8]),
		ModifierId:           parseIdentifier(column[9]),
	}, nil

}

func parseReferenceSetHeader(column [][]byte) (*snomed.ReferenceSetItem, error) {
	return &snomed.ReferenceSetItem{
		Id:                    string(column[0]), // identifier is a long unique uuid string,
		EffectiveTime:         parseDate(column[1]),
		Active:                parseBoolean(column[2]),
		ModuleId:              parseIdentifier(column[3]),
		RefsetId:              parseIdentifier(column[4]),
		ReferencedComponentId: parseIdentifier(column[5]),
	}, nil
}

func parseRefsetDescriptorRefset(row []byte) (*snomed.ReferenceSetItem, error) {
	column := bytes.Split(row, []byte{'\t'})
	item, err := parseReferenceSetHeader(column)
	item.Body = &snomed.ReferenceSetItem_RefsetDescriptor{
		RefsetDescriptor: &snomed.RefSetDescriptorReferenceSet{
			AttributeDescriptionId: parseInt(column[6]),
			AttributeTypeId:        parseInt(column[7]),
			AttributeOrder:         uint32(parseInt(column[8])),
		},
	}
	return item, err
}

func parseLanguageRefset(row []byte) (*snomed.ReferenceSetItem, error) {
	column := bytes.Split(row, []byte{'\t'})
	item, err := parseReferenceSetHeader(column)
	item.Body = &snomed.ReferenceSetItem_Language{
		Language: &snomed.LanguageReferenceSet{
			AcceptabilityId: parseInt(column[6]),
		},
	}
	return item, err
}

func parseSimpleRefset(row []byte) (*snomed.ReferenceSetItem, error) {
	column := bytes.Split(row, []byte{'\t'})
	item, err := parseReferenceSetHeader(column)
	item.Body = &snomed.ReferenceSetItem_Simple{
		Simple: &snomed.SimpleReferenceSet{},
	}
	return item, err
}

func parseSimpleMapRefset(row []byte) (*snomed.ReferenceSetItem, error) {
	column := bytes.Split(row, []byte{'\t'})
	item, err := parseReferenceSetHeader(column)
	item.Body = &snomed.ReferenceSetItem_SimpleMap{
		SimpleMap: &snomed.SimpleMapReferenceSet{
			MapTarget: string(column[6]),
		},
	}
	return item, err
}

func parseExtendedMapRefset(row []byte) (*snomed.ReferenceSetItem, error) {
	column := bytes.Split(row, []byte{'\t'})
	item, err := parseReferenceSetHeader(column)
	item.Body = &snomed.ReferenceSetItem_ComplexMap{
		ComplexMap: &snomed.ComplexMapReferenceSet{
			MapGroup:    parseInt(column[6]),
			MapPriority: parseInt(column[7]),
			MapRule:     string(column[8]),
			MapAdvice:   string(column[9]),
			MapTarget:   strings.TrimSpace(string(column[10])),
			Correlation: parseInt(column[11]),
			MapCategory: parseInt(column[12]),
		},
	}
	return item, err
}
func parseComplexMapRefset(row []byte) (*snomed.ReferenceSetItem, error) {
	column := bytes.Split(row, []byte{'\t'})
	item, err := parseReferenceSetHeader(column)
	item.Body = &snomed.ReferenceSetItem_ComplexMap{
		ComplexMap: &snomed.ComplexMapReferenceSet{
			MapGroup:    parseInt(column[6]),
			MapPriority: parseInt(column[7]),
			MapRule:     string(column[8]),
			MapAdvice:   string(column[9]),
			MapTarget:   strings.TrimSpace(string(column[10])),
			Correlation: parseInt(column[11]),
			MapBlock:    parseInt(column[12]),
		},
	}
	return item, err
}
func parseAttributeValueRefset(row []byte) (*snomed.ReferenceSetItem, error) {
	column := bytes.Split(row, []byte{'\t'})
	item, err := parseReferenceSetHeader(column)
	item.Body = &snomed.ReferenceSetItem_AttributeValue{
		AttributeValue: &snomed.AttributeValueReferenceSet{
			ValueId: parseInt(column[6]),
		},
	}
	return item, err
}
func parseAssociationRefset(row []byte) (*snomed.ReferenceSetItem, error) {
	column := bytes.Split(row, []byte{'\t'})
	item, err := parseReferenceSetHeader(column)
	item.Body = &snomed.ReferenceSetItem_Association{
		Association: &snomed.AssociationReferenceSet{
			TargetComponentId: parseInt(column[6]),
		},
	}
	return item, err
}
