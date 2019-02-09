package server

import (
	"github.com/wardle/go-terminology/dmd"
	"github.com/wardle/go-terminology/medicine"
	"golang.org/x/net/context"
)

// dmdSrv implements the medicine.dmdServer gRPC interface
type dmdSrv struct {
	dmd.DmdServer
	svc *medicine.Svc
}

func (ds *dmdSrv) ParseMedication(ctx context.Context, medicationString *dmd.MedicationString) (*dmd.ParsedMedication, error) {
	return ds.svc.ParseMedicationString(medicationString.S)
}

func (ds *dmdSrv) GetVTM(ctx context.Context, id *dmd.ItemID) (*dmd.VirtualTherapeuticMoiety, error) {
	return ds.svc.GetVTM(id.Id)
}

func (ds *dmdSrv) GetVMP(ctx context.Context, id *dmd.ItemID) (*dmd.VirtualMedicinalProduct, error) {
	return ds.svc.GetVMP(id.Id)
}

func (ds *dmdSrv) GetAMP(ctx context.Context, id *dmd.ItemID) (*dmd.ActualMedicinalProduct, error) {
	return ds.svc.GetAMP(id.Id)
}

func (ds *dmdSrv) GetTypeAttributes(ctx context.Context, id *dmd.MedicationType) (*dmd.MedicationTypeAttributes, error) {
	return ds.svc.GetTypeAttributes(id.Id)
}

func (ds *dmdSrv) GetMedicationType(ctx context.Context, pair *dmd.MedicationTypeMapping_FormRoutePair) (*dmd.MedicationTypeMapping, error) {
	return ds.svc.GetMedicationType(pair.Form, pair.Route)
}
