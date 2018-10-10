package server

import (
	"github.com/wardle/go-terminology/terminology/medicine"
	"golang.org/x/net/context"
)

// dmdSrv implements the medicine.dmdServer gRPC interface
type dmdSrv struct {
	medicine.DmdServer
}

func (ds *dmdSrv) ParseMedication(ctx context.Context, medicationString *medicine.MedicationString) (*medicine.ParsedMedication, error) {
	pm := medicine.ParseMedicationString(medicationString.S)
	return pm, nil
}
