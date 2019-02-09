package goleveldb

import (
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/dgraph-io/badger"
	"github.com/gogo/protobuf/proto"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/wardle/go-terminology/dmd"
	"github.com/wardle/go-terminology/medicine/storage"
)

// goleveldbService is file-based storage service for dm+d based on golevedb It
// implements the `storage.Store` Interface
type goleveldbService struct {
	db *leveldb.DB
	storage.Store
}

// Current version of schema
const currentVersion = 0.2

/*
	# Schema v0.2

	goLevelDB is a key value store. The following is the key -> value
	construction used to store dm+d items and associated index/reverse look-up
	information.

	Key                                 Value
	m/<vtm_id>                          dmd.VirtualTherapeuticMoiety   protobuf
	v/<vmp_id>                          dmd.VirtualMedicinalProduct    protobuf
	a/<amp_id>                          dmd.ActualMedicinalProduct     protobuf
	d/<medication_type>                 dmd.MedicationTypeAttributes   protobuf
	t/<route>/<form>/<medication_type>  <medication_type>              int32

	s/<key>                             <metadata value>
*/

var (
	schemaKey        = []byte("s/schemaversion") // Metadata key for schema version
	versionKey       = []byte("s/version")       // Metadata key for dm+d version
	vtmPrefix        = []byte("m/")
	vmpPrefix        = []byte("v/")
	ampPrefix        = []byte("a/")
	attributesPrefix = []byte("d/")
	typePrefix       = []byte("t/")
)

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

// getSchema fetches schema version of opened datastore.
func (bs *goleveldbService) getSchema() (float64, error) {
	var schema float64
	data, err := bs.db.Get(schemaKey, nil)
	if err != nil {
		return schema, err
	}
	return strconv.ParseFloat(string(data), 64)
}

// putSchema sets the schema version of opened datastore.
func (bs *goleveldbService) putSchema(schema float64) error {
	value := []byte(strconv.FormatFloat(schema, 'f', -1, 64))
	return bs.db.Put(schemaKey, value, nil)
}

// New creates a new storage service at the specified location, defaults to
// read-only but can be opened writable returns `storage.Store` interface
func New(path string, readOnly bool) (storage.Store, error) {
	var service goleveldbService

	opts := opt.Options{
		CompactionTableSizeMultiplier: 2,
		ReadOnly:                      readOnly,
	}

	db, err := leveldb.OpenFile(path, &opts)
	if err != nil {
		return &service, err
	}
	service.db = db

	schema, err := service.getSchema()
	if err != nil && err != errors.ErrNotFound {
		db.Close()
		return &service, err
	}

	if schema != currentVersion && err == nil {
		db.Close()
		return &service, fmt.Errorf("Incompatible dm+d datastore format v%v, needed v%v", schema, currentVersion)
	}

	if err == badger.ErrKeyNotFound {
		return &service, service.putSchema(currentVersion)
	}

	return &service, nil
}

// Close releases all datastore resources.
func (bs *goleveldbService) Close() error {
	return bs.db.Close()
}

// Compacts underlying datastore.
func (bs *goleveldbService) Compact() error {
	return bs.db.CompactRange(util.Range{})
}

// PutVersion persists puts dm+d version into datastore.
func (bs *goleveldbService) PutVersion(version string) error {
	value := []byte(version)
	return bs.db.Put(versionKey, value, nil)
}

// GetVersion fetches dm+d version of opened datastore.
func (bs *goleveldbService) GetVersion() (string, error) {
	var version string
	data, err := bs.db.Get(versionKey, nil)
	if err != nil {
		return version, err
	}
	return string(data), nil
}

// Put a slice of dm+d items into persistent storage.
// This is polymorphic but expects a slice of a dm+d items (VTM, VMP or AMP)
func (bs *goleveldbService) Put(dmdItems interface{}) error {
	var err error
	switch dmdItems.(type) {
	case []*dmd.VirtualTherapeuticMoiety:
		err = bs.putVTM(dmdItems.([]*dmd.VirtualTherapeuticMoiety))
	case []*dmd.VirtualMedicinalProduct:
		err = bs.putVMP(dmdItems.([]*dmd.VirtualMedicinalProduct))
	case []*dmd.ActualMedicinalProduct:
		err = bs.putAMP(dmdItems.([]*dmd.ActualMedicinalProduct))
	case []*dmd.MedicationTypeAttributes:
		err = bs.putTypeAttributes(dmdItems.([]*dmd.MedicationTypeAttributes))
	case []*dmd.MedicationTypeMapping:
		err = bs.putTypeMapping(dmdItems.([]*dmd.MedicationTypeMapping))
	default:
		err = fmt.Errorf("unknown dmdItems type: %T", dmdItems)
	}
	return err
}

// putVTM stores a slice of *dmd.VirtualTherapeuticMoiety in datastore
func (bs *goleveldbService) putVTM(vtms []*dmd.VirtualTherapeuticMoiety) error {
	batch := new(leveldb.Batch)
	for _, vtm := range vtms {
		key := append(vtmPrefix, itob(vtm.Id)...)
		data, err := proto.Marshal(vtm)
		if err != nil {
			return err
		}
		batch.Put(key, data)
	}
	return bs.db.Write(batch, nil)
}

// putVMP stores a slice of *dmd.VirtualMedicinalProduct in datastore
func (bs *goleveldbService) putVMP(vmps []*dmd.VirtualMedicinalProduct) error {
	batch := new(leveldb.Batch)
	for _, vmp := range vmps {
		key := append(vmpPrefix, itob(vmp.Id)...)
		data, err := proto.Marshal(vmp)
		if err != nil {
			return err
		}
		batch.Put(key, data)
	}
	return bs.db.Write(batch, nil)
}

// putAMP stores a slice of *dmd.ActualMedicinalProduct in datastore
func (bs *goleveldbService) putAMP(amps []*dmd.ActualMedicinalProduct) error {
	batch := new(leveldb.Batch)
	for _, amp := range amps {
		key := append(ampPrefix, itob(amp.Id)...)
		data, err := proto.Marshal(amp)
		if err != nil {
			return err
		}
		batch.Put(key, data)
	}
	return bs.db.Write(batch, nil)
}

// putTypeAttributes stores a slice of *dmd.MedicationTypeAttributes in datastore
func (bs *goleveldbService) putTypeAttributes(typeAttributes []*dmd.MedicationTypeAttributes) error {
	batch := new(leveldb.Batch)
	for _, group := range typeAttributes {
		key := append(attributesPrefix, []byte(group.Id)...)
		data, err := proto.Marshal(group)
		if err != nil {
			return err
		}
		batch.Put(key, data)
	}
	return bs.db.Write(batch, nil)
}

// putTypeMapping stores a slice of *dmd.MedicationTypeMapping in datastore
func (bs *goleveldbService) putTypeMapping(typeMapping []*dmd.MedicationTypeMapping) error {
	for _, mapping := range typeMapping {
		batch := new(leveldb.Batch)

		medicationType := dmd.MedicationTypeMapping{
			Id:   mapping.Id,
			Name: mapping.Name,
		}
		data, err := proto.Marshal(&medicationType)
		if err != nil {
			return err
		}

		for _, formRoutePair := range mapping.FormRoutePairs {
			var key []byte
			route := itob(formRoutePair.Route)
			form := itob(formRoutePair.Form)
			medTypeId := itob(int64(mapping.Id))
			sperator := []byte("/")
			for _, r := range [][]byte{typePrefix, route, sperator, form, sperator, medTypeId} {
				key = append(key, r...)
			}
			batch.Put(key, data)
		}

		err = bs.db.Write(batch, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetVTM get the *dmd.VirtualTherapeuticMoiety for associated vtmID int64
func (bs *goleveldbService) GetVTM(vtmID int64) (*dmd.VirtualTherapeuticMoiety, error) {
	var vtm dmd.VirtualTherapeuticMoiety
	data, err := bs.db.Get(append(vtmPrefix, itob(vtmID)...), nil)
	if err != nil {
		return &vtm, err
	}
	err = proto.Unmarshal(data, &vtm)

	return &vtm, err
}

// GetVMP get the *dmd.VirtualMedicinalProduct for associated vmpID int64
func (bs *goleveldbService) GetVMP(vmpID int64) (*dmd.VirtualMedicinalProduct, error) {
	var vmp dmd.VirtualMedicinalProduct
	data, err := bs.db.Get(append(vmpPrefix, itob(vmpID)...), nil)
	if err != nil {
		return &vmp, err
	}
	err = proto.Unmarshal(data, &vmp)

	return &vmp, err
}

// GetAMP get the *dmd.ActualMedicinalProduct for associated ampID int64
func (bs *goleveldbService) GetAMP(ampID int64) (*dmd.ActualMedicinalProduct, error) {
	var amp dmd.ActualMedicinalProduct
	data, err := bs.db.Get(append(ampPrefix, itob(ampID)...), nil)
	if err != nil {
		return &amp, err
	}
	err = proto.Unmarshal(data, &amp)

	return &amp, err
}

// GetTypeAttributes get the *dmd.MedicationTypeAttributes for associated Medication Type string
func (bs *goleveldbService) GetTypeAttributes(MedicationType string) (*dmd.MedicationTypeAttributes, error) {
	var attributes dmd.MedicationTypeAttributes
	data, err := bs.db.Get(append(attributesPrefix, []byte(MedicationType)...), nil)
	if err != nil {
		return &attributes, err
	}
	err = proto.Unmarshal(data, &attributes)

	return &attributes, err
}

// GetMedicationType get the *dmd.MedicationTypeMapping for associated Form int64, Route int64 pair
func (bs *goleveldbService) GetMedicationType(form int64, route int64) (*dmd.MedicationTypeMapping, error) {
	var medicationType dmd.MedicationTypeMapping
	var key []byte

	formb := itob(form)
	routeb := itob(route)
	sperator := []byte("/")
	for _, r := range [][]byte{typePrefix, routeb, sperator, formb} {
		key = append(key, r...)
	}

	iter := bs.db.NewIterator(util.BytesPrefix(key), nil)
	for iter.Next() {
		err := proto.Unmarshal(iter.Value(), &medicationType)
		if err != nil {
			return &medicationType, err
		}
	}
	iter.Release()
	err := iter.Error()

	return &medicationType, err
}
