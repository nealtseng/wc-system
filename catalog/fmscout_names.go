package catalog

// fmscoutNames maps FIFA team ID to FMScout national team name string.
var fmscoutNames = map[string]string{
	"ARG": "Argentina",
	"AUS": "Australia",
	"AUT": "Austria",
	"ALG": "Algeria",
	"BIH": "Bosnia and Herzegovina",
	"BEL": "Belgium",
	"BRA": "Brazil",
	"CAN": "Canada",
	"CIV": "Ivory Coast",
	"COL": "Colombia",
	"CPV": "Cape Verde",
	"COD": "DR Congo",
	"CRO": "Croatia",
	"CUW": "Curacao",
	"CZE": "Czech Republic",
	"ECU": "Ecuador",
	"EGY": "Egypt",
	"ENG": "England",
	"ESP": "Spain",
	"FRA": "France",
	"GER": "Germany",
	"GHA": "Ghana",
	"HAI": "Haiti",
	"IRN": "Iran",
	"IRQ": "Iraq",
	"JOR": "Jordan",
	"JPN": "Japan",
	"KOR": "South Korea",
	"KSA": "Saudi Arabia",
	"MAR": "Morocco",
	"MEX": "Mexico",
	"NED": "Netherlands",
	"NOR": "Norway",
	"NZL": "New Zealand",
	"PAN": "Panama",
	"PAR": "Paraguay",
	"POR": "Portugal",
	"QAT": "Qatar",
	"RSA": "South Africa",
	"SCO": "Scotland",
	"SEN": "Senegal",
	"SUI": "Switzerland",
	"SWE": "Sweden",
	"TUN": "Tunisia",
	"TUR": "Turkiye",
	"URU": "Uruguay",
	"USA": "United States",
	"UZB": "Uzbekistan",
}

// FMScoutName returns the FMScout-recognized national team name for a catalog team ID.
func FMScoutName(teamID string) (string, bool) {
	n, ok := fmscoutNames[teamID]
	return n, ok
}
