package medicine

import (
	"errors"
	"math"
)

var defaultFrequencies = map[string]Frequency{
	"PER_HOUR":           Frequency{286551008, []string{"/hour", "/hr", "/h", "every-hour"}, func(dose float64) float64 { return dose * 24.0 }},
	"TWELVE_TIMES_DAILY": Frequency{396106003, []string{"12/day", "12/d", "twelve-times-daily"}, func(dose float64) float64 { return dose * 12.0 }},
	"TEN_TIMES_DAILY":    Frequency{396105004, []string{"10/day", "10/d", "ten-times-daily"}, func(dose float64) float64 { return dose * 10.0 }},
	"NINE_TIMES_DAILY":   Frequency{396115005, []string{"9/day", "9/d", "nine-times-daily"}, func(dose float64) float64 { return dose * 9.0 }},
	"EIGHT_TIMES_DAILY":  Frequency{307443002, []string{"8/day", "8/d", "eight-times-daily"}, func(dose float64) float64 { return dose * 8.0 }},
	"SEVEN_TIMES_DAILY":  Frequency{307442007, []string{"7/day", "7/d", "seven-times-daily"}, func(dose float64) float64 { return dose * 7.0 }},
	"SIX_TIMES_DAILY":    Frequency{307441000, []string{"6/day", "6/d", "six-times-daily"}, func(dose float64) float64 { return dose * 6.0 }},
	"FIVE_TIMES_DAILY":   Frequency{307440004, []string{"5/day", "5/d", "five-times-daily"}, func(dose float64) float64 { return dose * 5.0 }},
	"FOUR_TIMES_DAILY":   Frequency{307439001, []string{"qds", "4/day", "4/d", "four-times-daily"}, func(dose float64) float64 { return dose * 4.0 }},
	"THREE_TIMES_DAILY":  Frequency{229798009, []string{"tds", "tid", "3/day", "3/d", "three-times-daily"}, func(dose float64) float64 { return dose * 3.0 }},
	"TWICE_DAILY":        Frequency{229799001, []string{"bd", "bid", "2/day", "2/d", "twice-daily", "two-times-daily"}, func(dose float64) float64 { return dose * 2.0 }},
	"ONCE_DAILY":         Frequency{229797004, []string{"od", "1/day", "1/d", "once-daily", "one-time-daily"}, func(dose float64) float64 { return dose }},
	"ALTERNATE_DAYS":     Frequency{225760004, []string{"altdays", "alt", "alternate-days"}, func(dose float64) float64 { return math.Ceil((dose/2)*100) / 100 }},
	"ONCE_WEEKLY":        Frequency{225769003, []string{"/week", "/w", "/wk", "1/w", "once-every-week"}, func(dose float64) float64 { return math.Ceil((dose/7)*100) / 100 }},
	"ONCE_TWO_WEEKLY":    Frequency{20050000, []string{"/2weeks", "/2w", "/2wk", "once-every-two-weeks"}, func(dose float64) float64 { return math.Ceil((dose/14)*100) / 100 }},
	"ONCE_MONTHLY":       Frequency{307450003, []string{"/month", "/m", "/mo", "1/m", "once-every-month"}, func(dose float64) float64 { return math.Ceil((dose/30)*10) / 10 }},
	"ONCE_TWO_MONTHLY":   Frequency{445547001, []string{"/2months", "/2m", "/2mo", "once-every-two-months"}, func(dose float64) float64 { return math.Ceil((dose/60)*10) / 10 }},
	"ONCE_THREE_MONTHLY": Frequency{396129006, []string{"/3months", "/3m", "/3mo", "once-every-three-months"}, func(dose float64) float64 { return math.Ceil((dose/90)*10) / 10 }},
	"ONCE_YEARLY":        Frequency{307455008, []string{"/year", "/y", "/yr", "once-every-year"}, func(dose float64) float64 { return math.Ceil((dose/368)*10) / 10 }},
}

func FrequencyByName(search string) (Frequency, error) {
	for _, v := range defaultFrequencies {
		for _, name := range v.Names() {
			if search == name {
				return v, nil
			}
		}
	}
	return Frequency{}, errors.New("No Matching Frequency")
}

var defaultUnits = map[string]Units{
	"MICROGRAM":   Units{DoseBased, 258685003, []string{"mcg", "micrograms"}, 0.00001},
	"MILLIGRAM":   Units{DoseBased, 258684004, []string{"mg"}, 0.001},
	"MILLILITRES": Units{ProductBased, 258773002, []string{"ml"}, 0.001},
	"GRAM":        Units{DoseBased, 258682000, []string{"g", "gram"}, 1},
	"UNITS":       Units{ProductBased, 408102007, []string{"units", "u"}, 1},
	"TABLETS":     Units{ProductBased, 385055001, []string{"tablets", "tab", "t"}, 1},
	"PUFFS":       Units{ProductBased, 415215001, []string{"puffs", "puff", "p"}, 1},
	"NONE":        Units{ProductBased, 408102007, []string{""}, 1},
}

func UnitsByAbbreviation(search string) (Units, error) {
	for _, v := range defaultUnits {
		for _, abbreviation := range v.Abbreviations() {
			if search == abbreviation {
				return v, nil
			}
		}
	}
	return Units{}, errors.New("No Matching Units")
}

var defaultRoutes = map[string]Route{
	"ORAL":          Route{26643006, "po"},
	"INTRAVENOUS":   Route{47625008, "iv"},
	"SUBCUTANEOUS":  Route{34206005, "sc"},
	"INTRAMUSCULAR": Route{78421000, "im"},
	"INTRATHECAL":   Route{72607000, "intrathecal"},
	"INHALED":       Route{2764101000001108, "inh"},
	"TOPICAL":       Route{2762601000001108, "top"},
}

func RouteByAbbreviation(search string) (Route, error) {
	for _, v := range defaultRoutes {
		if search == v.Abbreviation() {
			return v, nil
		}
	}
	return Route{}, errors.New("No Matching Route")
}
