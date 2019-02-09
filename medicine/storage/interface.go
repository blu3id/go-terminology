package storage

import (
	"github.com/wardle/go-terminology/dmd"
)

// Store is an interface to the pluggable backend abstracted dm+d persistence
// service. A storage service must implement this interface.
type Store interface {
	Put(dmdItem interface{}) error
	PutVersion(dmdVersion string) error
	GetVersion() (string, error)
	GetVTM(vtmID int64) (*dmd.VirtualTherapeuticMoiety, error)
	GetVMP(vmpID int64) (*dmd.VirtualMedicinalProduct, error)
	GetAMP(ampID int64) (*dmd.ActualMedicinalProduct, error)
	GetTypeAttributes(MedicationType string) (*dmd.MedicationTypeAttributes, error)
	GetMedicationType(form int64, route int64) (*dmd.MedicationTypeMapping, error)
	Compact() error
	Close() error
}
