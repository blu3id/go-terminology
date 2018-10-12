package medicine

import (
	"encoding/json"
	"strconv"
	"strings"
)

//go:generate protoc -I. --go_out=plugins=gprc:. medicine.proto
//go:generate protoc -I. -I../../vendor/terminology/vendor/googleapis --go_out=plugins=grpc:. dmdservice.proto
//go:generate protoc -I. -I../../vendor/terminology/vendor/googleapis --grpc-gateway_out=logtostderr=true:. dmdservice.proto

func (m *ParsedMedication) equivalentDose() float64 {
	return m.Units.Conversion() * m.Dose
}

func (m *ParsedMedication) dailyEquivalentDose() float64 {
	if m.Frequency.ConceptId != 0 {
		return m.Frequency.DailyEquivalentDose(m.equivalentDose())
	}
	return 0
}

func (m *ParsedMedication) BuildString() string {
	var output string
	if m.MappedDrugName == "" {
		output = output + strings.Title(m.DrugName) + " "
	} else {
		output = output + m.MappedDrugName + " "
	}
	if m.Dose != 0 && m.Units != nil {
		output = output + strconv.FormatFloat(m.Dose, 'f', -1, 64) + m.Units.Abbreviation() + " "
	}
	if m.Frequency != nil {
		output = output + m.Frequency.Title() + " "
	}
	if m.Route != nil {
		output = output + m.Route.Abbreviation + " "
	}
	if m.AsRequired {
		output = output + "PRN" + " "
	}
	if m.Notes != "" {
		output = output + m.Notes
	}
	return strings.Trim(output, " ")
}

func (m *ParsedMedication) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		DrugName            string
		MappedDrugName      string
		Dose                float64
		EquivalentDose      float64
		DailyEquivalentDose float64
		Units               string
		Frequency           string
		Route               string
		AsRequired          bool
		ConceptId           int64
		Notes               string
		String              string
	}{
		DrugName:            m.DrugName,
		MappedDrugName:      m.MappedDrugName,
		Dose:                m.Dose,
		EquivalentDose:      m.equivalentDose(),
		DailyEquivalentDose: m.dailyEquivalentDose(),
		Units:               m.Units.Abbreviation(),
		Frequency:           m.Frequency.Title(),
		Route:               m.Route.Abbreviation,
		AsRequired:          m.AsRequired,
		ConceptId:           m.ConceptId,
		Notes:               m.Notes,
		String:              m.BuildString(),
	})
}

func (u *Units) Abbreviation() string {
	if len(u.Abbreviations) > 0 {
		return u.Abbreviations[0]
	}
	return ""
}

func (u *Units) Conversion() float64 {
	return UnitsConversion(u.ConceptId)
}

func (f *Frequency) Title() string {
	if len(f.Names) > 0 {
		return f.Names[0]
	}
	return ""
}

func (f *Frequency) DailyEquivalentDose(dose float64) float64 {
	return DailyEquivalentDose(f.ConceptId, dose)
}
