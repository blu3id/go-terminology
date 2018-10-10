package medicine

import (
	"errors"
	"math"
)

type FrequencyConst struct {
	Frequency
	equivalentDose func(float64) float64
}

func (f *FrequencyConst) DailyEquivalentDose(dose float64) float64 {
	return f.equivalentDose(dose)
}

var defaultFrequencies = map[string]FrequencyConst{
	"PER_HOUR":           FrequencyConst{Frequency{ConceptId: 286551008, Names: []string{"/hour", "/hr", "/h", "every-hour"}}, func(dose float64) float64 { return dose * 24.0 }},
	"TWELVE_TIMES_DAILY": FrequencyConst{Frequency{ConceptId: 396106003, Names: []string{"12/day", "12/d", "twelve-times-daily"}}, func(dose float64) float64 { return dose * 12.0 }},
	"TEN_TIMES_DAILY":    FrequencyConst{Frequency{ConceptId: 396105004, Names: []string{"10/day", "10/d", "ten-times-daily"}}, func(dose float64) float64 { return dose * 10.0 }},
	"NINE_TIMES_DAILY":   FrequencyConst{Frequency{ConceptId: 396115005, Names: []string{"9/day", "9/d", "nine-times-daily"}}, func(dose float64) float64 { return dose * 9.0 }},
	"EIGHT_TIMES_DAILY":  FrequencyConst{Frequency{ConceptId: 307443002, Names: []string{"8/day", "8/d", "eight-times-daily"}}, func(dose float64) float64 { return dose * 8.0 }},
	"SEVEN_TIMES_DAILY":  FrequencyConst{Frequency{ConceptId: 307442007, Names: []string{"7/day", "7/d", "seven-times-daily"}}, func(dose float64) float64 { return dose * 7.0 }},
	"SIX_TIMES_DAILY":    FrequencyConst{Frequency{ConceptId: 307441000, Names: []string{"6/day", "6/d", "six-times-daily"}}, func(dose float64) float64 { return dose * 6.0 }},
	"FIVE_TIMES_DAILY":   FrequencyConst{Frequency{ConceptId: 307440004, Names: []string{"5/day", "5/d", "five-times-daily"}}, func(dose float64) float64 { return dose * 5.0 }},
	"FOUR_TIMES_DAILY":   FrequencyConst{Frequency{ConceptId: 307439001, Names: []string{"qds", "4/day", "4/d", "four-times-daily"}}, func(dose float64) float64 { return dose * 4.0 }},
	"THREE_TIMES_DAILY":  FrequencyConst{Frequency{ConceptId: 229798009, Names: []string{"tds", "tid", "3/day", "3/d", "three-times-daily"}}, func(dose float64) float64 { return dose * 3.0 }},
	"TWICE_DAILY":        FrequencyConst{Frequency{ConceptId: 229799001, Names: []string{"bd", "bid", "2/day", "2/d", "twice-daily", "two-times-daily"}}, func(dose float64) float64 { return dose * 2.0 }},
	"ONCE_DAILY":         FrequencyConst{Frequency{ConceptId: 229797004, Names: []string{"od", "1/day", "1/d", "once-daily", "one-time-daily"}}, func(dose float64) float64 { return dose }},
	"ALTERNATE_DAYS":     FrequencyConst{Frequency{ConceptId: 225760004, Names: []string{"altdays", "alt", "alternate-days"}}, func(dose float64) float64 { return math.Ceil((dose/2)*100) / 100 }},
	"ONCE_WEEKLY":        FrequencyConst{Frequency{ConceptId: 225769003, Names: []string{"/week", "/w", "/wk", "1/w", "once-every-week"}}, func(dose float64) float64 { return math.Ceil((dose/7)*100) / 100 }},
	"ONCE_TWO_WEEKLY":    FrequencyConst{Frequency{ConceptId: 20050000, Names: []string{"/2weeks", "/2w", "/2wk", "once-every-two-weeks"}}, func(dose float64) float64 { return math.Ceil((dose/14)*100) / 100 }},
	"ONCE_MONTHLY":       FrequencyConst{Frequency{ConceptId: 307450003, Names: []string{"/month", "/m", "/mo", "1/m", "once-every-month"}}, func(dose float64) float64 { return math.Ceil((dose/30)*10) / 10 }},
	"ONCE_TWO_MONTHLY":   FrequencyConst{Frequency{ConceptId: 445547001, Names: []string{"/2months", "/2m", "/2mo", "once-every-two-months"}}, func(dose float64) float64 { return math.Ceil((dose/60)*10) / 10 }},
	"ONCE_THREE_MONTHLY": FrequencyConst{Frequency{ConceptId: 396129006, Names: []string{"/3months", "/3m", "/3mo", "once-every-three-months"}}, func(dose float64) float64 { return math.Ceil((dose/90)*10) / 10 }},
	"ONCE_YEARLY":        FrequencyConst{Frequency{ConceptId: 307455008, Names: []string{"/year", "/y", "/yr", "once-every-year"}}, func(dose float64) float64 { return math.Ceil((dose/368)*10) / 10 }},
}

func DailyEquivalentDose(concept int64, dose float64) float64 {
	for _, v := range defaultFrequencies {
		if concept == v.ConceptId {
			return v.DailyEquivalentDose(dose)
		}
	}
	return 0
}

func FrequencyByName(search string) (Frequency, error) {
	for _, v := range defaultFrequencies {
		for _, name := range v.Names {
			if search == name {
				return v.Frequency, nil
			}
		}
	}
	return Frequency{}, errors.New("No Matching Frequency")
}

type UnitsConst struct {
	Units
	conversion float64
}

var defaultUnits = map[string]UnitsConst{
	"MICROGRAM":   UnitsConst{Units{PrescribingType: Units_DOSE_BASED, ConceptId: 258685003, Abbreviations: []string{"mcg", "micrograms"}}, 0.00001},
	"MILLIGRAM":   UnitsConst{Units{PrescribingType: Units_DOSE_BASED, ConceptId: 258684004, Abbreviations: []string{"mg"}}, 0.001},
	"MILLILITRES": UnitsConst{Units{PrescribingType: Units_PRODUCT_BASED, ConceptId: 258773002, Abbreviations: []string{"ml"}}, 0.001},
	"GRAM":        UnitsConst{Units{PrescribingType: Units_DOSE_BASED, ConceptId: 258682000, Abbreviations: []string{"g", "gram"}}, 1},
	"UNITS":       UnitsConst{Units{PrescribingType: Units_PRODUCT_BASED, ConceptId: 408102007, Abbreviations: []string{"units", "u"}}, 1},
	"TABLETS":     UnitsConst{Units{PrescribingType: Units_PRODUCT_BASED, ConceptId: 385055001, Abbreviations: []string{"tablets", "tab", "t"}}, 1},
	"PUFFS":       UnitsConst{Units{PrescribingType: Units_PRODUCT_BASED, ConceptId: 415215001, Abbreviations: []string{"puffs", "puff", "p"}}, 1},
	"NONE":        UnitsConst{Units{PrescribingType: Units_PRODUCT_BASED, ConceptId: 408102007, Abbreviations: []string{""}}, 1},
}

func UnitsConversion(concept int64) float64 {
	for _, v := range defaultUnits {
		if concept == v.ConceptId {
			return v.conversion
		}
	}
	return 0
}

func UnitsByAbbreviation(search string) (Units, error) {
	for _, v := range defaultUnits {
		for _, abbreviation := range v.Abbreviations {
			if search == abbreviation {
				return v.Units, nil
			}
		}
	}
	return Units{}, errors.New("No Matching Units")
}

var defaultRoutes = map[string]Route{
	"ORAL":          Route{ConceptId: 26643006, Abbreviation: "po"},
	"INTRAVENOUS":   Route{ConceptId: 47625008, Abbreviation: "iv"},
	"SUBCUTANEOUS":  Route{ConceptId: 34206005, Abbreviation: "sc"},
	"INTRAMUSCULAR": Route{ConceptId: 78421000, Abbreviation: "im"},
	"INTRATHECAL":   Route{ConceptId: 72607000, Abbreviation: "intrathecal"},
	"INHALED":       Route{ConceptId: 2764101000001108, Abbreviation: "inh"},
	"TOPICAL":       Route{ConceptId: 2762601000001108, Abbreviation: "top"},
}

func RouteByAbbreviation(search string) (Route, error) {
	for _, v := range defaultRoutes {
		if search == v.Abbreviation {
			return v, nil
		}
	}
	return Route{}, errors.New("No Matching Route")
}
