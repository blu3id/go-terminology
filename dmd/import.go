package dmd

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type xmlVTM struct {
	XMLName xml.Name `xml:"VTM"`
	Id      int64    `xml:"VTMID"`
	Invalid bool     `xml:"INVALID"`
	Name    string   `xml:"NM"`
}

type xmlVMP struct {
	XMLName                    xml.Name `xml:"VMP"`
	Id                         int64    `xml:"VPID"`
	VirtualTherapeuticMoietyId int64    `xml:"VTMID"`
	Invalid                    bool     `xml:"INVALID"`
	Name                       string   `xml:"NM"`
	CombinationProduct         int      `xml:"COMBPRODCD"`
	PrescribingStatus          int      `xml:"PRES_STATCD"`
	NonAvailability            int      `xml:"NON_AVAILCD"`
	DoseForm                   int      `xml:"DF_INDCD"`
	UnitDoseFormSize           float64  `xml:"UDFS"`
	UnitDoseFormUnits          int64    `xml:"UDFS_UOMCD"`
	UnitDoseUnits              int64    `xml:"UNIT_DOSE_UOMCD"`
}

type xmlVPI struct {
	XMLName                   xml.Name `xml:"VPI"`
	VirtualMedicinalProductId int64    `xml:"VPID"`
	Ingredient                int64    `xml:"ISID"`
	BasisOfStrength           int      `xml:"BASIS_STRNTCD"`
	BasisOfStrengthSubstance  int64    `xml:"BS_SUBID"`
	StrengthNumerator         float64  `xml:"STRNT_NMRTR_VAL"`
	StrengthNumeratorUnit     int64    `xml:"STRNT_NMRTR_UOMCD"`
	StrengthDenominator       float64  `xml:"STRNT_DNMTR_VAL"`
	StrengthDenominatorUnit   int64    `xml:"STRNT_DNMTR_UOMCD"`
}

type xmlDForm struct {
	XMLName                   xml.Name `xml:"DFORM"`
	VirtualMedicinalProductId int64    `xml:"VPID"`
	DrugForm                  int64    `xml:"FORMCD"`
}

type xmlDRoute struct {
	XMLName                   xml.Name `xml:"DROUTE"`
	VirtualMedicinalProductId int64    `xml:"VPID"`
	DrugRoute                 int64    `xml:"ROUTECD"`
}

type xmlCDInfo struct {
	XMLName                   xml.Name `xml:"CONTROL_INFO"`
	VirtualMedicinalProductId int64    `xml:"VPID"`
	ControledDrugCategory     int      `xml:"CATCD"`
}

type xmlAMP struct {
	XMLName                   xml.Name `xml:"AMP"`
	Id                        int64    `xml:"APID"`
	Invalid                   bool     `xml:"INVALID"`
	VirtualMedicinalProductId int64    `xml:"VPID"`
	Name                      string   `xml:"NM"`
	Description               string   `xml:"DESC"`
	Supplier                  int64    `xml:"SUPPCD"`
	LicensingAuthority        int      `xml:"LIC_AUTHCD"`
	AvailabilityRestriction   int      `xml:"AVAIL_RESTRICTCD"`
}

type xmlLRoute struct {
	XMLName                  xml.Name `xml:"LIC_ROUTE"`
	ActualMedicinalProductId int64    `xml:"APID"`
	LicensedRoute            int64    `xml:"ROUTECD"`
}

type xmlAttribute struct {
	Name        string `xml:"name,attr"`
	Cardinality string `xml:"cardinality,attr"`
}

type xmlAttributeGroup struct {
	XMLName    xml.Name       `xml:"GROUP"`
	Id         string         `xml:"type,attr"`
	Name       string         `xml:"name,attr"`
	Usage      string         `xml:"Usage"`
	Attributes []xmlAttribute `xml:"attribute"`
}

type xmlFormRoutePair struct {
	Form struct {
		Form int64 `xml:"code,attr"`
	} `xml:"form"`
	Route struct {
		Route int64 `xml:"code,attr"`
	} `xml:"route"`
}

type xmlMedicationType struct {
	XMLName        xml.Name           `xml:"medication_type"`
	Id             int32              `xml:"type,attr"`
	Name           string             `xml:"name,attr"`
	RouteFormPairs []xmlFormRoutePair `xml:"form_route_pair"`
}

var (
	VTMFileRegEx             = regexp.MustCompile(`f_vtm2_\d+.xml`)
	VMPFileRegEx             = regexp.MustCompile(`f_vmp2_\d+.xml`)
	AMPFileRegEx             = regexp.MustCompile(`f_amp2_\d+.xml`)
	AttributesFileRegEx      = regexp.MustCompile(`UK_SNOMED_CT_Medication_Types_Spec_Data_Specifications_\d+.xml`)
	MedicationTypesFileRegEx = regexp.MustCompile(`UK_SNOMED_CT_Medication_Types_Form_Route_Pairs_\d+.xml`)
)

// PerformImport takes a path to dm+d .xml relase files and a handler function
// that accepts slices of dm+d items to persist.
//
// TODO: Make more memory efficient as currently hold entire dm+d import in
// memory to aid precomputation and denormalisation from import files as memory
// is cheap + ubiquitious and dataset isn't to large.
func PerformImport(path string, handler func(interface{}) error) error {
	var (
		VTMFile, VMPFile, AMPFile           string
		AttributesFile, MedicationTypesFile string
		vtms                                []*VirtualTherapeuticMoiety
		vmps                                []*VirtualMedicinalProduct
		amps                                []*ActualMedicinalProduct
		count                               int
	)

	f, err := os.Open(path)
	if err != nil {
		fmt.Errorf("Error opening dm+d import path: %v", err)
	}
	files, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		fmt.Errorf("Error opening dm+d import path: %v", err)
	}

	for _, file := range files {
		if VTMFileRegEx.Match([]byte(file.Name())) {
			VTMFile = filepath.Join(path, file.Name())
		}
		if VMPFileRegEx.Match([]byte(file.Name())) {
			VMPFile = filepath.Join(path, file.Name())
		}
		if AMPFileRegEx.Match([]byte(file.Name())) {
			AMPFile = filepath.Join(path, file.Name())
		}
		if AttributesFileRegEx.Match([]byte(file.Name())) {
			AttributesFile = filepath.Join(path, file.Name())
		}
		if MedicationTypesFileRegEx.Match([]byte(file.Name())) {
			MedicationTypesFile = filepath.Join(path, file.Name())
		}
	}

	if VTMFile == "" {
		return fmt.Errorf("No VTM file found matching f_vtm2_<release>.xml")
	}
	if VMPFile == "" {
		return fmt.Errorf("No VMP file found matching f_vmp2_<release>.xml")
	}
	if AMPFile == "" {
		return fmt.Errorf("No AMP file found matching f_amp2_<release>.xml")
	}
	if AttributesFile == "" {
		return fmt.Errorf("No Type Attributes File file found matching UK_SNOMED_CT_Medication_Types_Spec_Data_Specifications_<release>.xml")
	}
	if MedicationTypesFile == "" {
		return fmt.Errorf("No Medication Types File file found matching UK_SNOMED_CT_Medication_Types_Form_Route_Pairs_<release>.xml")
	}

	// Parse Virtual Therapeutic Moieties
	vtmm, err := readVTM2(VTMFile)
	if err != nil {
		return fmt.Errorf("Error parsing VTM file: %v", err)
	}

	// Parse Virtual Medicinal Products
	vmpm, err := readVMP2(VMPFile)
	if err != nil {
		return fmt.Errorf("Error parsing VTM file: %v", err)
	}

	// Parse Virtual Medicinal Products
	ampm, err := readAMP2(AMPFile)
	if err != nil {
		return fmt.Errorf("Error parsing VTM file: %v", err)
	}

	// Parse Medication Type Attributes & Persist
	typeAttributes, err := readAttributesSpec(AttributesFile)
	if err != nil {
		return fmt.Errorf("Error parsing Type Attributes file: %v", err)
	}
	if err := handler(typeAttributes); err != nil {
		return fmt.Errorf("Error persisting Type Attributes: %v", err)
	}

	// Parse Medication Types & Persist
	medicationTypes, err := readMedicationTypes(MedicationTypesFile)
	if err != nil {
		return fmt.Errorf("Error parsing Medication Types file: %v", err)
	}
	if err := handler(medicationTypes); err != nil {
		return fmt.Errorf("Error persisting Medication Types: %v", err)
	}

	// Precompute children and build slices for handler
	count = 0
	for _, amp := range ampm {
		if amp.Id == 0 {
			continue
		}
		if amp.VirtualMedicinalProductId != 0 {
			vmpm[amp.VirtualMedicinalProductId].ActualMedicinalProducts = append(vmpm[amp.VirtualMedicinalProductId].ActualMedicinalProducts, amp.Id)
		}
		amps = append(amps, amp)
		count++
		if count%100 == 0 {
			if err := handler(amps); err != nil {
				return fmt.Errorf("Error persisting AMPs: %v", err)
			}
			amps = []*ActualMedicinalProduct{}
		}
	}
	if err := handler(amps); err != nil {
		return fmt.Errorf("Error persisting AMPs: %v", err)
	}

	count = 0
	for _, vmp := range vmpm {
		if vmp.Id == 0 {
			continue
		}
		if vmp.VirtualTherapeuticMoietyId != 0 {
			vtmm[vmp.VirtualTherapeuticMoietyId].VirtualMedicinalProducts = append(vtmm[vmp.VirtualTherapeuticMoietyId].VirtualMedicinalProducts, vmp.Id)
		}
		vmps = append(vmps, vmp)
		count++
		if count%100 == 0 {
			if err := handler(vmps); err != nil {
				return fmt.Errorf("Error persisting VMPs: %v", err)
			}
			vmps = []*VirtualMedicinalProduct{}
		}
	}
	if err := handler(vmps); err != nil {
		return fmt.Errorf("Error persisting VMPs: %v", err)
	}

	count = 0
	for _, vtm := range vtmm {
		if vtm.Id == 0 {
			continue
		}
		vtms = append(vtms, vtm)
		count++
		if count%100 == 0 {
			if err := handler(vtms); err != nil {
				return fmt.Errorf("Error persisting VTMs: %v", err)
			}
			vtms = []*VirtualTherapeuticMoiety{}
		}
	}
	if err := handler(vtms); err != nil {
		return fmt.Errorf("Error persisting VTMs: %v", err)
	}

	return nil
}

func readVTM2(path string) (map[int64]*VirtualTherapeuticMoiety, error) {
	inMemory := make(map[int64]*VirtualTherapeuticMoiety)
	f, err := os.Open(path)
	if err != nil {
		return inMemory, fmt.Errorf("Error opening file: %v", err)
	}
	defer f.Close()
	d := xml.NewDecoder(f)
	for {
		t, tokenErr := d.Token()
		if tokenErr != nil {
			if tokenErr == io.EOF {
				break
			}
			return inMemory, tokenErr
		}

		switch t := t.(type) {
		case xml.StartElement:
			if t.Name.Local == "VTM" {
				var vtm xmlVTM
				if err := d.DecodeElement(&vtm, &t); err != nil {
					return inMemory, err
				}
				if vtm.Invalid {
					continue
				}
				inMemory[vtm.Id] = &VirtualTherapeuticMoiety{
					Id:   vtm.Id,
					Name: vtm.Name,
				}
			}
		}
	}
	return inMemory, nil
}

func readVMP2(path string) (map[int64]*VirtualMedicinalProduct, error) {
	inMemory := make(map[int64]*VirtualMedicinalProduct)
	f, err := os.Open(path)
	if err != nil {
		return inMemory, fmt.Errorf("Error opening file: %v", err)
	}
	defer f.Close()
	d := xml.NewDecoder(f)
	for {
		t, tokenErr := d.Token()
		if tokenErr != nil {
			if tokenErr == io.EOF {
				break
			}
			return inMemory, tokenErr
		}

		switch t := t.(type) {
		case xml.StartElement:
			if t.Name.Local == "VMP" {
				var vmp xmlVMP
				if err := d.DecodeElement(&vmp, &t); err != nil {
					return inMemory, err
				}
				if vmp.Invalid {
					continue
				}
				inMemory[vmp.Id] = &VirtualMedicinalProduct{
					Id:                         vmp.Id,
					VirtualTherapeuticMoietyId: vmp.VirtualTherapeuticMoietyId,
					Name:                       vmp.Name,
					CombinationProduct:         VirtualMedicinalProduct_CombinationProduct(vmp.CombinationProduct),
					PrescribingStatus:          VirtualMedicinalProduct_PrescribingStatus(vmp.PrescribingStatus),
					NonAvailability:            VirtualMedicinalProduct_NonAvailability(vmp.NonAvailability),
					DoseForm:                   VirtualMedicinalProduct_DoseForm(vmp.DoseForm),
					UnitDoseFormSize:           vmp.UnitDoseFormSize,
					UnitDoseFormUnits:          vmp.UnitDoseFormUnits,
					UnitDoseUnits:              vmp.UnitDoseUnits,
				}
			}

			if t.Name.Local == "VPI" {
				var vpi xmlVPI
				if err := d.DecodeElement(&vpi, &t); err != nil {
					return inMemory, err
				}
				if _, ok := inMemory[vpi.VirtualMedicinalProductId]; ok {
					inMemory[vpi.VirtualMedicinalProductId].Ingredients = append(inMemory[vpi.VirtualMedicinalProductId].Ingredients,
						&VirtualMedicinalProduct_Ingredient{
							Ingredient:               vpi.Ingredient,
							BasisOfStrength:          VirtualMedicinalProduct_Ingredient_BasisOfStrength(vpi.BasisOfStrength),
							BasisOfStrengthSubstance: vpi.BasisOfStrengthSubstance,
							StrengthNumerator:        vpi.StrengthNumerator,
							StrengthNumeratorUnit:    vpi.StrengthNumeratorUnit,
							StrengthDenominator:      vpi.StrengthDenominator,
							StrengthDenominatorUnit:  vpi.StrengthDenominatorUnit,
						})
				}
			}

			if t.Name.Local == "DFORM" {
				var df xmlDForm
				if err := d.DecodeElement(&df, &t); err != nil {
					return inMemory, err
				}
				if _, ok := inMemory[df.VirtualMedicinalProductId]; ok {
					inMemory[df.VirtualMedicinalProductId].DrugForm = append(inMemory[df.VirtualMedicinalProductId].DrugForm, df.DrugForm)
				}
			}

			if t.Name.Local == "DROUTE" {
				var dr xmlDRoute
				if err := d.DecodeElement(&dr, &t); err != nil {
					return inMemory, err
				}
				if _, ok := inMemory[dr.VirtualMedicinalProductId]; ok {
					inMemory[dr.VirtualMedicinalProductId].DrugRoute = append(inMemory[dr.VirtualMedicinalProductId].DrugRoute, dr.DrugRoute)
				}
			}
			if t.Name.Local == "CONTROL_INFO" {
				var cd xmlCDInfo
				if err := d.DecodeElement(&cd, &t); err != nil {
					return inMemory, err
				}
				if _, ok := inMemory[cd.VirtualMedicinalProductId]; ok {
					inMemory[cd.VirtualMedicinalProductId].ControledDrugCategory = VirtualMedicinalProduct_ControledDrugCategory(cd.ControledDrugCategory)
				}
			}
		}
	}
	return inMemory, nil
}

func readAMP2(path string) (map[int64]*ActualMedicinalProduct, error) {
	inMemory := make(map[int64]*ActualMedicinalProduct)
	f, err := os.Open(path)
	if err != nil {
		return inMemory, fmt.Errorf("Error opening file: %v", err)
	}
	defer f.Close()
	d := xml.NewDecoder(f)
	for {
		t, tokenErr := d.Token()
		if tokenErr != nil {
			if tokenErr == io.EOF {
				break
			}
			return inMemory, tokenErr
		}

		switch t := t.(type) {
		case xml.StartElement:
			if t.Name.Local == "AMP" {
				var amp xmlAMP
				if err := d.DecodeElement(&amp, &t); err != nil {
					return inMemory, err
				}
				if amp.Invalid {
					continue
				}
				inMemory[amp.Id] = &ActualMedicinalProduct{
					Id:                        amp.Id,
					VirtualMedicinalProductId: amp.VirtualMedicinalProductId,
					Name:                      amp.Name,
					Description:               amp.Description,
					Supplier:                  amp.Supplier,
					LicensingAuthority:        ActualMedicinalProduct_LicensingAuthority(amp.LicensingAuthority),
					AvailabilityRestriction:   ActualMedicinalProduct_AvailabilityRestriction(amp.AvailabilityRestriction),
				}
			}

			if t.Name.Local == "LIC_ROUTE" {
				var lr xmlLRoute
				if err := d.DecodeElement(&lr, &t); err != nil {
					return inMemory, err
				}
				if _, ok := inMemory[lr.ActualMedicinalProductId]; ok {
					inMemory[lr.ActualMedicinalProductId].LicensedRoute = append(inMemory[lr.ActualMedicinalProductId].LicensedRoute, lr.LicensedRoute)
				}
			}
		}
	}
	return inMemory, nil
}

func readAttributesSpec(path string) ([]*MedicationTypeAttributes, error) {
	var output []*MedicationTypeAttributes
	f, err := os.Open(path)
	if err != nil {
		return output, fmt.Errorf("Error opening file: %v", err)
	}
	defer f.Close()
	d := xml.NewDecoder(f)
	for {
		t, tokenErr := d.Token()
		if tokenErr != nil {
			if tokenErr == io.EOF {
				break
			}
			return output, tokenErr
		}

		switch t := t.(type) {
		case xml.StartElement:
			if t.Name.Local == "GROUP" {
				var attrGroup xmlAttributeGroup
				if err := d.DecodeElement(&attrGroup, &t); err != nil {
					return output, err
				}
				group := &MedicationTypeAttributes{
					Id:    attrGroup.Id,
					Name:  strings.TrimSpace(attrGroup.Name),
					Usage: attrGroup.Usage,
				}
				for _, v := range attrGroup.Attributes {
					switch v.Cardinality {
					case "required":
						group.RequiredAttributes = append(group.RequiredAttributes, v.Name)
					case "optional":
						group.OptionalAttributes = append(group.OptionalAttributes, v.Name)
					}
				}
				output = append(output, group)
			}
		}
	}
	return output, nil
}

func readMedicationTypes(path string) ([]*MedicationTypeMapping, error) {
	var output []*MedicationTypeMapping
	f, err := os.Open(path)
	if err != nil {
		return output, fmt.Errorf("Error opening file: %v", err)
	}
	defer f.Close()
	d := xml.NewDecoder(f)
	for {
		t, tokenErr := d.Token()
		if tokenErr != nil {
			if tokenErr == io.EOF {
				break
			}
			return output, tokenErr
		}

		switch t := t.(type) {
		case xml.StartElement:
			if t.Name.Local == "medication_type" {
				var medicationType xmlMedicationType
				if err := d.DecodeElement(&medicationType, &t); err != nil {
					return output, err
				}
				mapping := &MedicationTypeMapping{
					Id:   medicationType.Id,
					Name: strings.TrimSpace(medicationType.Name),
				}
				for _, v := range medicationType.RouteFormPairs {
					mapping.FormRoutePairs = append(mapping.FormRoutePairs, &MedicationTypeMapping_FormRoutePair{
						Form:  v.Form.Form,
						Route: v.Route.Route,
					})
				}
				output = append(output, mapping)
			}
		}
	}
	return output, nil
}
