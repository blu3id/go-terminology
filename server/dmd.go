package server

import (
	"github.com/wardle/go-terminology/terminology"
	"github.com/wardle/go-terminology/terminology/medicine"
	"golang.org/x/net/context"
)

// dmdSrv implements the medicine.dmdServer gRPC interface
type dmdSrv struct {
	medicine.DmdServer
	svc *terminology.Svc
}

func (ds *dmdSrv) ParseMedication(ctx context.Context, medicationString *medicine.MedicationString) (*medicine.ParsedMedication, error) {
	return ds.svc.ParseMedicationString(medicationString.S)
}
