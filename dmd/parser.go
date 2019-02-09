package dmd

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var dosingRegEx = regexp.MustCompile(`(\d+\.{0,1}\d*)\s*(mg|mcg|micrograms|g|u|units|unit|t|tab|tablets|puffs|p|puff)`)

func matchDosing(token string) (float64, Units, error) {
	matches := dosingRegEx.FindStringSubmatch(token)
	if matches == nil || len(matches) != 3 {
		return 0.0, Units{}, errors.New("No Matching Dosing")
	}

	dose, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0.0, Units{}, errors.New("No Matching Dosing")
	}
	units, err := UnitsByAbbreviation(matches[2])
	if err != nil {
		return 0.0, Units{}, errors.New("No Matching Dosing")
	}
	return dose, units, nil
}

func ParseMedicationString(text string) *ParsedMedication {
	var medication ParsedMedication
	var makingDrugName = true
	var makingNotes = false

	for _, token := range strings.Split(strings.ToLower(text), " ") {
		if makingNotes {
			medication.Notes = medication.Notes + token + " "
		} else {
			if token == "prn" {
				medication.AsRequired = true
				makingDrugName = false
			} else {
				frequency, err := FrequencyByName(token)
				if err == nil {
					medication.Frequency = &frequency
					makingDrugName = false
				} else {
					dose, units, err := matchDosing(token)
					if err == nil {
						medication.Dose = dose
						medication.Units = &units
						makingDrugName = false
					} else {
						route, err := RouteByAbbreviation(token)
						if err == nil {
							medication.Route = &route
							makingDrugName = false
						} else if makingDrugName {
							medication.DrugName = medication.DrugName + token + " "
						} else {
							makingNotes = true
							medication.Notes = medication.Notes + token + " "
						}
					}
				}
			}

		}
	}

	medication.DrugName = strings.Trim(medication.DrugName, " ")
	medication.Notes = strings.Trim(medication.Notes, " ")

	medication.DailyEquivalentDose = medication.dailyEquivalentDose()
	medication.EquivalentDose = medication.equivalentDose()
	medication.String_ = medication.BuildString()

	return &medication
}
