package badgerdb

import (
	"encoding/binary"
	"fmt"
	"runtime"
	"strconv"

	"github.com/dgraph-io/badger"
	"github.com/gogo/protobuf/proto"
	"github.com/wardle/go-terminology/dmd"
	"github.com/wardle/go-terminology/medicine/storage"
)

// badgerdbService is file-based storage service for dm+d based on BadgerDb It
// implements the `storage.Store` Interface
type badgerdbService struct {
	db *badger.DB
	storage.Store
}

// Current version of schema
const currentVersion = 0.1

/*
	# Schema v0.1

	BadgerDB is a key value store. The following is the key -> value
	construction used to store dm+d items and associated index/reverse look-up
	information.

	Key         Value
	m/<vtm_id>  dmd.VirtualTherapeuticMoiety   protobuf
	v/<vmp_id>  dmd.VirtualMedicinalProduct    protobuf
	a/<amp_id>  dmd.ActualMedicinalProduct     protobuf
	s/<key>     <metadata value>
*/

var (
	schemaKey  = []byte("s/schemaversion") // Metadata key for schema version
	versionKey = []byte("s/version")       // Metadata key for dm+d version
	vtmPrefix  = []byte("m/")
	vmpPrefix  = []byte("v/")
	ampPrefix  = []byte("a/")
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
func (bs *badgerdbService) getSchema() (float64, error) {
	var schema float64
	err := bs.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(schemaKey)
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			schema, err = strconv.ParseFloat(string(val), 64)
			return err
		})
		return err
	})

	return schema, err
}

// putSchema sets the schema version of opened datastore.
func (bs *badgerdbService) putSchema(schema float64) error {
	err := bs.db.Update(func(txn *badger.Txn) error {
		value := []byte(strconv.FormatFloat(schema, 'f', -1, 64))
		if err := txn.Set(schemaKey, value); err != nil {
			return err
		}
		return nil
	})
	return err
}

// New creates a new storage service at the specified location, defaults to
// read-only but can be opened writable returns `storage.Store` interface
func New(path string, readOnly bool) (storage.Store, error) {
	var service badgerdbService

	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path
	opts.ValueLogFileSize = 1024 * 1024 * 20

	if runtime.GOOS != "windows" {
		opts.ReadOnly = readOnly
	}

	db, err := badger.Open(opts)
	if err != nil {
		return &service, err
	}
	service.db = db

	schema, err := service.getSchema()
	if err != nil && err != badger.ErrKeyNotFound {
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
func (bs *badgerdbService) Close() error {
	return bs.db.Close()
}

// Compacts underlying datastore. Not implemented.
func (bs *badgerdbService) Compact() error {
	return nil
}

// PutVersion persists puts dm+d version into datastore.
func (bs *badgerdbService) PutVersion(version string) error {
	err := bs.db.Update(func(txn *badger.Txn) error {
		value := []byte(version)
		if err := txn.Set(versionKey, value); err != nil {
			return err
		}
		return nil
	})
	return err
}

// GetVersion fetches dm+d version of opened datastore.
func (bs *badgerdbService) GetVersion() (string, error) {
	var version string
	err := bs.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(versionKey)
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			version = string(val)
			return err
		})
		return err
	})

	return version, err
}

// Put a slice of dm+d items into persistent storage.
// This is polymorphic but expects a slice of a dm+d items (VTM, VMP or AMP)
func (bs *badgerdbService) Put(dmdItems interface{}) error {
	var err error
	switch dmdItems.(type) {
	case []*dmd.VirtualTherapeuticMoiety:
		err = bs.putVTM(dmdItems.([]*dmd.VirtualTherapeuticMoiety))
	case []*dmd.VirtualMedicinalProduct:
		err = bs.putVMP(dmdItems.([]*dmd.VirtualMedicinalProduct))
	case []*dmd.ActualMedicinalProduct:
		err = bs.putAMP(dmdItems.([]*dmd.ActualMedicinalProduct))
	default:
		err = fmt.Errorf("unknown dmdItems type: %T", dmdItems)
	}
	return err
}

// putVTM stores a slice of *dmd.VirtualTherapeuticMoiety in datastore
func (bs *badgerdbService) putVTM(vtms []*dmd.VirtualTherapeuticMoiety) error {
	err := bs.db.Update(func(txn *badger.Txn) error {
		for _, vtm := range vtms {
			key := append(vtmPrefix, itob(vtm.Id)...)
			data, err := proto.Marshal(vtm)
			if err != nil {
				return err
			}
			if err := txn.Set(key, data); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// putVMP stores a slice of *dmd.VirtualMedicinalProduct in datastore
func (bs *badgerdbService) putVMP(vmps []*dmd.VirtualMedicinalProduct) error {
	err := bs.db.Update(func(txn *badger.Txn) error {
		for _, vmp := range vmps {
			key := append(vmpPrefix, itob(vmp.Id)...)
			data, err := proto.Marshal(vmp)
			if err != nil {
				return err
			}
			if err := txn.Set(key, data); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// putAMP stores a slice of *dmd.ActualMedicinalProduct in datastore
func (bs *badgerdbService) putAMP(amps []*dmd.ActualMedicinalProduct) error {
	err := bs.db.Update(func(txn *badger.Txn) error {
		for _, amp := range amps {
			key := append(ampPrefix, itob(amp.Id)...)
			data, err := proto.Marshal(amp)
			if err != nil {
				return err
			}
			if err := txn.Set(key, data); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// GetVTM get the *dmd.VirtualTherapeuticMoiety for associated vtmID int64
func (bs *badgerdbService) GetVTM(vtmID int64) (*dmd.VirtualTherapeuticMoiety, error) {
	var vtm dmd.VirtualTherapeuticMoiety
	err := bs.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(append(vtmPrefix, itob(vtmID)...))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			if err := proto.Unmarshal(val, &vtm); err != nil {
				return err
			}
			return nil
		})
		return err
	})
	return &vtm, err
}

// GetVMP get the *dmd.VirtualMedicinalProduct for associated vmpID int64
func (bs *badgerdbService) GetVMP(vmpID int64) (*dmd.VirtualMedicinalProduct, error) {
	var vmp dmd.VirtualMedicinalProduct
	err := bs.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(append(vmpPrefix, itob(vmpID)...))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			if err := proto.Unmarshal(val, &vmp); err != nil {
				return err
			}
			return nil
		})
		return err
	})
	return &vmp, err
}

// GetAMP get the *dmd.ActualMedicinalProduct for associated ampID int64
func (bs *badgerdbService) GetAMP(ampID int64) (*dmd.ActualMedicinalProduct, error) {
	var amp dmd.ActualMedicinalProduct
	err := bs.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(append(ampPrefix, itob(ampID)...))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			if err := proto.Unmarshal(val, &amp); err != nil {
				return err
			}
			return nil
		})
		return err
	})
	return &amp, err
}
