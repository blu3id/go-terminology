package bolt

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	boltdb "github.com/boltdb/bolt"
	"github.com/gogo/protobuf/proto"
	"github.com/wardle/go-terminology/snomed"
	"github.com/wardle/go-terminology/terminology/interfaces"
)

// bolt2Service is file-based storage service for SNOMED-CT based on Bolt
// It implements the `terminology.Store` Interface
type bolt2Service struct {
	db *boltdb.DB
	interfaces.Store
}

// itob returns binary []byte representation of int64
func itob(v int64) []byte {
	bufer := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(bufer, v)
	return bufer[:n]
}

// btoi returns int64 of binary []byte
func btoi(b []byte) int64 {
	x, n := binary.Varint(b)
	if n != len(b) {
		//panic("Error decoding []byte to int64")
	}
	return x
}

var (
	// Bucket structure
	rootBucket           = []byte("Root")          // Root Bucket, containing terminology data (Descriptions, Relationships, ReferenceSets)
	conceptBucket        = []byte("Concepts")      // Bucket containing Concepts
	reverseDescriptions  = []byte("Descriptions")  // Bucket containing Description (ID) -> Concept (ID) mapping
	reverseReferenceSets = []byte("ReferenceSets") // Bucket containing ReferenceSet (ID) + ReferenceSetMember -> Component (ID) mapping

	// Bolt options
	defaultOptions = &boltdb.Options{
		Timeout:    0,
		NoGrowSync: false,
		ReadOnly:   false,
	}
	readOnlyOptions = &boltdb.Options{
		Timeout:    0,
		NoGrowSync: false,
		ReadOnly:   true,
	}
)

// New creates a new storage service at the specified location, defaults to writable but can be opened read only
// Returns `interfaces.Store` interface
func New(filename string, readOnly bool) (interfaces.Store, error) {
	options := defaultOptions
	if readOnly {
		options = readOnlyOptions
	}
	db, err := boltdb.Open(filename, 0644, options)
	if err != nil {
		return nil, err
	}
	return &bolt2Service{db: db}, nil
}

// Close releases all database resources.
func (bs *bolt2Service) Close() error {
	return bs.db.Close()
}

// Put a slice of SNOMED-CT components into persistent storage.
// This is polymorphic but expects a slice of a core SNOMED CT component (Concept, Description, Relationship, ReferenceSetItem)
func (bs *bolt2Service) Put(components interface{}) error {
	var err error
	switch components.(type) {
	case []*snomed.Concept:
		err = bs.putConcepts(components.([]*snomed.Concept))
	case []*snomed.Description:
		err = bs.putDescriptions(components.([]*snomed.Description))
	case []*snomed.Relationship:
		err = bs.putRelationships(components.([]*snomed.Relationship))
	case []*snomed.ReferenceSetItem:
		err = bs.putReferenceSets(components.([]*snomed.ReferenceSetItem))
	default:
		err = fmt.Errorf("unknown component type: %T", components)
	}
	return err
}

// putConcepts persists the specified concepts
// Stored under key: <conceptID> in rootBucket as serialised Protobuf
func (bs *bolt2Service) putConcepts(concepts []*snomed.Concept) error {
	return bs.db.Update(func(tx *boltdb.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(conceptBucket)
		if err != nil {
			return err
		}
		for _, concept := range concepts {
			data, err := proto.Marshal(concept)
			if err != nil {
				return err
			}
			key := itob(concept.Id)
			if err := bucket.Put(key, data); err != nil {
				return err
			}
		}
		return nil
	})
}

// putConcepts persists the specified descriptions
// Stored under key: <conceptID>/d/<descriptionID> in rootBucket as serialised Protobuf
// Reverse lookup <conceptID> stored under <descriptionID> in reverseDescriptions
func (bs *bolt2Service) putDescriptions(descriptions []*snomed.Description) error {
	return bs.db.Update(func(tx *boltdb.Tx) error {
		rootBucket, err := tx.CreateBucketIfNotExists(rootBucket)
		if err != nil {
			return err
		}
		reverseDescriptions, err := tx.CreateBucketIfNotExists(reverseDescriptions)
		if err != nil {
			return err
		}
		for _, description := range descriptions {
			data, err := proto.Marshal(description)
			if err != nil {
				return err
			}
			descriptionKey := append(append(itob(description.ConceptId), []byte("d")...), itob(description.Id)...)
			if err := rootBucket.Put(descriptionKey, data); err != nil {
				return err
			}
			reverseDescriptionKey := itob(description.Id)
			if err := reverseDescriptions.Put(reverseDescriptionKey, itob(description.ConceptId)); err != nil {
				return err
			}
		}
		return nil
	})
}

// putRelationships persists the specified relationships
// Relationship stored under Parent (source) at key: <conceptID>/p/<relationshipID> in rootBucket as serialised Protobuf
// Reverse lookup from Child (destination) stored at key: <conceptID>/c/<relationshipID> in rootBucket as <conceptID>
func (bs *bolt2Service) putRelationships(relationships []*snomed.Relationship) error {
	return bs.db.Update(func(tx *boltdb.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(rootBucket)
		if err != nil {
			return err
		}
		for _, relationship := range relationships {
			data, err := proto.Marshal(relationship)
			if err != nil {
				return err
			}
			parentRelationshipKey := append(append(itob(relationship.SourceId), []byte("p")...), itob(relationship.Id)...)
			if err := bucket.Put(parentRelationshipKey, data); err != nil {
				return err
			}
			childRelationshipKey := append(append(itob(relationship.DestinationId), []byte("c")...), itob(relationship.Id)...)
			if err := bucket.Put(childRelationshipKey, itob(relationship.SourceId)); err != nil {
				return err
			}
		}
		return nil
	})
}

// putReferenceSets persists the specified reference set items
// Stored under key: <conceptID>/r/<refsetID> in rootBucket as serialised Protobuf
// Reverse lookup <conceptID> stored under <refsetID>/m/<refsetMemberId:UUID> in reverseReferenceSets
func (bs *bolt2Service) putReferenceSets(refsetItems []*snomed.ReferenceSetItem) error {
	return bs.db.Update(func(tx *boltdb.Tx) error {
		rootBucket, err := tx.CreateBucketIfNotExists(rootBucket)
		if err != nil {
			return err
		}
		reverseReferenceSets, err := tx.CreateBucketIfNotExists(reverseReferenceSets)
		if err != nil {
			return err
		}
		for _, refsetItem := range refsetItems {
			data, err := proto.Marshal(refsetItem)
			if err != nil {
				return err
			}
			refsetKey := append(append(itob(refsetItem.ReferencedComponentId), []byte("r")...), itob(refsetItem.RefsetId)...)
			if err := rootBucket.Put(refsetKey, data); err != nil {
				return err
			}
			reverseRefsetKey := append(append(itob(refsetItem.RefsetId), []byte("m")...), []byte(refsetItem.Id)...)
			if err := reverseReferenceSets.Put(reverseRefsetKey, itob(refsetItem.ReferencedComponentId)); err != nil {
				return err
			}
		}
		return nil
	})
}

// GetConcept fetches a concept with the given identifier
func (bs *bolt2Service) GetConcept(conceptID int64) (*snomed.Concept, error) {
	var concept snomed.Concept
	err := bs.db.View(func(tx *boltdb.Tx) error {
		bucket := tx.Bucket(conceptBucket)
		if bucket == nil {
			return errors.New("No Root Bucket found. Has data been loaded yet?")
		}
		data := bucket.Get(itob(conceptID))
		if data == nil {
			return fmt.Errorf("No object found with id: %d", conceptID)
		}
		if err := proto.Unmarshal(data, &concept); err != nil {
			return err
		}
		return nil
	})
	return &concept, err
}

// GetConcepts returns a slice of concepts with the given identifiers
func (bs *bolt2Service) GetConcepts(conceptIDs ...int64) ([]*snomed.Concept, error) {
	result := make([]*snomed.Concept, len(conceptIDs))
	err := bs.db.View(func(tx *boltdb.Tx) error {
		bucket := tx.Bucket(conceptBucket)
		if bucket == nil {
			return errors.New("No Root Bucket found. Has data been loaded yet?")
		}
		for i, conceptID := range conceptIDs {
			var concept snomed.Concept
			data := bucket.Get(itob(conceptID))
			if data == nil {
				return fmt.Errorf("No object found with id: %d", conceptID)
			}
			if err := proto.Unmarshal(data, &concept); err != nil {
				return err
			}
			result[i] = &concept
		}
		return nil
	})
	return result, err
}

// GetDescription returns the description with the given identifier
func (bs *bolt2Service) GetDescription(descriptionID int64) (*snomed.Description, error) {
	var description snomed.Description
	err := bs.db.View(func(tx *boltdb.Tx) error {
		rootBucket := tx.Bucket(rootBucket)
		if rootBucket == nil {
			return errors.New("No Root Bucket found. Has data been loaded yet?")
		}
		reverseDescriptions := tx.Bucket(reverseDescriptions)
		if reverseDescriptions == nil {
			return errors.New("No reverseDescriptions Bucket found. Has data been loaded yet?")
		}
		conceptID := reverseDescriptions.Get(itob(descriptionID))
		if conceptID == nil {
			return fmt.Errorf("No object found with id: %d", conceptID)
		}
		descriptionKey := append(append(conceptID, []byte("d")...), itob(descriptionID)...)
		data := reverseDescriptions.Get(descriptionKey)
		if data == nil {
			return fmt.Errorf("No object found with id: %d", descriptionKey)
		}
		if err := proto.Unmarshal(data, &description); err != nil {
			return err
		}
		return nil
	})
	return &description, err
}

// GetDescriptions returns the descriptions for this concept.
func (bs *bolt2Service) GetDescriptions(concept *snomed.Concept) ([]*snomed.Description, error) {
	result := make([]*snomed.Description, 0)
	err := bs.db.View(func(tx *boltdb.Tx) error {
		bucket := tx.Bucket(rootBucket)
		if bucket == nil {
			return errors.New("No Root Bucket found. Has data been loaded yet?")
		}
		c := bucket.Cursor()
		prefix := append(itob(concept.Id), []byte("d")...)
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			var description snomed.Description
			if err := proto.Unmarshal(v, &description); err != nil {
				return err
			}
			result = append(result, &description)
		}
		return nil
	})
	return result, err
}

// GetParentRelationships returns the parent relationships for this concept.
// Parent relationships are relationships in which this concept is the source.
func (bs *bolt2Service) GetParentRelationships(concept *snomed.Concept) ([]*snomed.Relationship, error) {
	result := make([]*snomed.Relationship, 0)
	err := bs.db.View(func(tx *boltdb.Tx) error {
		bucket := tx.Bucket(rootBucket)
		if bucket == nil {
			return errors.New("No Root Bucket found. Has data been loaded yet?")
		}
		c := bucket.Cursor()
		prefix := append(itob(concept.Id), []byte("p")...)
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			var relationship snomed.Relationship
			if err := proto.Unmarshal(v, &relationship); err != nil {
				return err
			}
			result = append(result, &relationship)
		}
		return nil
	})
	return result, err
}

// GetChildRelationships returns the child relationships for this concept.
// Child relationships are relationships in which this concept is the destination.
func (bs *bolt2Service) GetChildRelationships(concept *snomed.Concept) ([]*snomed.Relationship, error) {
	result := make([]*snomed.Relationship, 0)
	err := bs.db.View(func(tx *boltdb.Tx) error {
		bucket := tx.Bucket(rootBucket)
		if bucket == nil {
			return errors.New("No Root Bucket found. Has data been loaded yet?")
		}
		c := bucket.Cursor()
		prefix := append(itob(concept.Id), []byte("c")...)
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			var relationship snomed.Relationship
			relationshipID := bytes.TrimPrefix(k, prefix)
			data := bucket.Get(append(append(v, []byte("p")...), relationshipID...))
			if err := proto.Unmarshal(data, &relationship); err != nil {
				return err
			}
			result = append(result, &relationship)
		}
		return nil
	})
	return result, err
}

// GetReferenceSets returns the refset identifiers to which this component is a member
func (bs *bolt2Service) GetReferenceSets(componentID int64) ([]int64, error) {
	result := make([]int64, 0)
	err := bs.db.View(func(tx *boltdb.Tx) error {
		bucket := tx.Bucket(rootBucket)
		if bucket == nil {
			return errors.New("No Root Bucket found. Has data been loaded yet?")
		}
		c := bucket.Cursor()
		prefix := append(itob(componentID), []byte("r")...)
		for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			refsetID := bytes.TrimPrefix(k, prefix)
			result = append(result, btoi(refsetID))
			//var referenceSetItem snomed.ReferenceSetItem
			//if err := proto.Unmarshal(v, &referenceSetItem); err != nil {
			//	return err
			//}
			//result = append(result, &referenceSetItem)
		}
		return nil
	})
	return result, err
}

// GetReferenceSetItems returns the component (concept) identifiers in this refset
// TODO Returns Map Should be Slice
func (bs *bolt2Service) GetReferenceSetItems(refsetID int64) (map[int64]bool, error) {
	result := make(map[int64]bool)
	err := bs.db.View(func(tx *boltdb.Tx) error {
		bucket := tx.Bucket(reverseReferenceSets)
		if bucket == nil {
			return errors.New("No reverseReferenceSets Bucket found. Has data been loaded yet?")
		}
		c := bucket.Cursor()
		prefix := append(itob(refsetID), []byte("m")...)
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			result[btoi(v)] = true
			//var referenceSetItem snomed.ReferenceSetItem
			//refsetID := bytes.TrimPrefix(k, prefix)
			//data := bucket.Get(append(append(v, []byte("r")...), refsetID...))
			//if err := proto.Unmarshal(v, &referenceSetItem); err != nil {
			//	return err
			//}
			//result = append(result, &referenceSetItem)
		}
		return nil
	})
	return result, err
}

// GetFromReferenceSet gets the specified components from the specified refset, or error
func (bs *bolt2Service) GetFromReferenceSet(refsetID int64, componentID int64) (*snomed.ReferenceSetItem, error) {
	var referenceSetItem snomed.ReferenceSetItem
	err := bs.db.View(func(tx *boltdb.Tx) error {
		bucket := tx.Bucket(rootBucket)
		if bucket == nil {
			return errors.New("No Root Bucket found. Has data been loaded yet?")
		}
		data := bucket.Get(append(append(itob(componentID), []byte("r")...), itob(refsetID)...))
		if data == nil {
			return nil
			//return fmt.Errorf("No object found with id: %d %d", componentID, refsetID)
		}
		if err := proto.Unmarshal(data, &referenceSetItem); err != nil {
			return err
		}
		return nil
	})
	return &referenceSetItem, err
}

// GetAllReferenceSets returns a list of installed reference sets
func (bs *bolt2Service) GetAllReferenceSets() ([]int64, error) {
	result := make([]int64, 0)
	err := bs.db.View(func(tx *boltdb.Tx) error {
		bucket := tx.Bucket(reverseReferenceSets)
		if bucket == nil {
			return errors.New("No reverseReferenceSets Bucket found. Has data been loaded yet?")
		}
		var previous []byte
		bucket.ForEach(func(k, v []byte) error {
			current := k[:binary.MaxVarintLen64]
			if !bytes.Equal(current, previous) {
				if previous != nil {
					result = append(result, btoi(previous))
				}
				previous = k[:binary.MaxVarintLen64]
			}
			return nil
		})
		return nil
	})
	return result, err
}

// Iterate is a crude iterator for all concepts, useful for pre-processing and pre-computations
func (bs *bolt2Service) Iterate(fn func(*snomed.Concept) error) error {
	return bs.db.View(func(tx *boltdb.Tx) error {
		bucket := tx.Bucket([]byte(conceptBucket))
		var concept snomed.Concept
		return bucket.ForEach(func(k, v []byte) error {
			if err := proto.Unmarshal(v, &concept); err != nil {
				return err
			}
			return fn(&concept)
		})
	})
}

// GetAllChildrenIDs Not Implemented
// TODO Move to service not as primitive of store (until ?more efficient transative closure or materialized path implemented)
func GetAllChildrenIDs(concept *snomed.Concept) ([]int64, error) {
	return []int64{}, errors.New("GetAllChildrenIDs Not Implemented")
}

// GetStatistics returns statistics for the backend store
// This is crude and inefficient at the moment
// TODO(wardle): improve efficiency and speed
func (bs *bolt2Service) GetStatistics() (interfaces.Statistics, error) {
	stats := interfaces.Statistics{}
	refsetNames := make([]string, 0)
	err := bs.db.View(func(tx *boltdb.Tx) error {
		// concepts
		cBucket := tx.Bucket(conceptBucket)
		stats.Concepts = cBucket.Stats().KeyN
		// descriptions
		dBucket := tx.Bucket(reverseDescriptions)
		stats.Descriptions = dBucket.Stats().KeyN

		// reference sets
		rs, err := bs.GetAllReferenceSets()
		if err != nil {
			return err
		}
		stats.RefsetItems = len(rs)
		for _, refset := range rs {
			concept, err := bs.GetConcept(refset)
			if err != nil {
				return err
			}
			descs, err := bs.GetDescriptions(concept)
			if err != nil {
				return err
			}
			if len(descs) > 0 {
				refsetName := fmt.Sprintf("%s (%d)", descs[0].Term, concept.Id)
				refsetNames = append(refsetNames, refsetName)
			}
		}
		stats.Refsets = refsetNames
		return nil
	})
	return stats, err
}
