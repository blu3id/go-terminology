package medicine

import (
	"encoding/json"
	"strconv"
	"strings"
)

type ParsedMedication struct {
	drugName       string
	mappedDrugName string
	dose           float64
	units          Units
	frequency      Frequency
	route          Route
	asRequired     bool
	conceptID      int64
	notes          string
}

func (m *ParsedMedication) DrugName() string {
	return m.drugName
}

func (m *ParsedMedication) MappedDrugName() string {
	return m.mappedDrugName
}

func (m *ParsedMedication) SetMappedDrugName(mappedDrugName string) {
	m.mappedDrugName = mappedDrugName
}

func (m *ParsedMedication) Dose() float64 {
	return m.dose
}

func (m *ParsedMedication) EquivalentDose() float64 {
	return m.units.conversion * m.dose
}

func (m *ParsedMedication) DailyEquivalentDose() float64 {
	if m.frequency.conceptID != 0 {
		return m.frequency.DailyEquivalentDose(m.EquivalentDose())
	}
	return 0
}

func (m *ParsedMedication) Units() Units {
	return m.units
}

func (m *ParsedMedication) Frequency() Frequency {
	return m.frequency
}

func (m *ParsedMedication) Route() Route {
	return m.route
}

func (m *ParsedMedication) AsRequired() bool {
	return m.asRequired
}

func (m *ParsedMedication) ConceptID() int64 {
	return m.conceptID
}

func (m *ParsedMedication) SetConceptID(conceptID int64) {
	m.conceptID = conceptID
}

func (m *ParsedMedication) Notes() string {
	return m.notes
}

func (m *ParsedMedication) String() string {
	var output string
	if m.mappedDrugName == "" {
		output = output + strings.Title(m.drugName) + " "
	} else {
		output = output + m.mappedDrugName + " "
	}
	if m.dose != 0 && m.units.conceptID != 0 {
		output = output + strconv.FormatFloat(m.dose, 'f', -1, 64) + m.units.Abbreviation() + " "
	}
	if m.frequency.conceptID != 0 {
		output = output + m.frequency.Title() + " "
	}
	if m.route.conceptID != 0 {
		output = output + m.route.Abbreviation() + " "
	}
	if m.asRequired {
		output = output + "PRN" + " "
	}
	if m.notes != "" {
		output = output + m.notes
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
		ConceptID           int64
		Notes               string
		String              string
	}{
		DrugName:            m.drugName,
		MappedDrugName:      m.mappedDrugName,
		Dose:                m.dose,
		EquivalentDose:      m.EquivalentDose(),
		DailyEquivalentDose: m.DailyEquivalentDose(),
		Units:               m.units.Abbreviation(),
		Frequency:           m.frequency.Title(),
		Route:               m.route.Abbreviation(),
		AsRequired:          m.asRequired,
		ConceptID:           m.conceptID,
		Notes:               m.notes,
		String:              m.String(),
	})
}

type PrescribingType int8

const (
	DoseBased    PrescribingType = 1
	ProductBased PrescribingType = 2
)

type Units struct {
	prescribingType PrescribingType
	conceptID       int64
	abbreviations   []string
	conversion      float64
}

func (u *Units) PrescribingType() PrescribingType {
	return u.prescribingType
}

func (u *Units) ConceptID() int64 {
	return u.conceptID
}

func (u *Units) Abbreviation() string {
	if len(u.abbreviations) > 0 {
		return u.abbreviations[0]
	}
	return ""
}

func (u *Units) Abbreviations() []string {
	return u.abbreviations
}

type Frequency struct {
	conceptID      int64
	names          []string
	equivalentDose func(float64) float64
}

func (f *Frequency) Title() string {
	if len(f.names) > 0 {
		return f.names[0]
	}
	return ""
}

func (f *Frequency) ConceptID() int64 {
	return f.conceptID
}

func (f *Frequency) Names() []string {
	return f.names
}

func (f *Frequency) DailyEquivalentDose(dose float64) float64 {
	return f.equivalentDose(dose)
}

type Route struct {
	conceptID    int64
	abbreviation string
}

func (r *Route) ConceptID() int64 {
	return r.conceptID
}

func (r *Route) Abbreviation() string {
	return r.abbreviation
}
