package catalog

// fbrefSlugs maps FIFA team ID to FBref national team URL suffix.
var fbrefSlugs = map[string]string{
	"ARG": "ARG/Argentina-Men-Stats",
	"AUS": "AUS/Australia-Men-Stats",
	"AUT": "AUT/Austria-Men-Stats",
	"ALG": "ALG/Algeria-Men-Stats",
	"BIH": "BIH/Bosnia-and-Herzegovina-Men-Stats",
	"BEL": "BEL/Belgium-Men-Stats",
	"BRA": "BRA/Brazil-Men-Stats",
	"CAN": "CAN/Canada-Men-Stats",
	"CIV": "CIV/Ivory-Coast-Men-Stats",
	"COL": "COL/Colombia-Men-Stats",
	"CPV": "CPV/Cape-Verde-Men-Stats",
	"COD": "COD/DR-Congo-Men-Stats",
	"CRO": "CRO/Croatia-Men-Stats",
	"CUW": "CUW/Curacao-Men-Stats",
	"CZE": "CZE/Czech-Republic-Men-Stats",
	"ECU": "ECU/Ecuador-Men-Stats",
	"EGY": "EGY/Egypt-Men-Stats",
	"ENG": "ENG/England-Men-Stats",
	"ESP": "ESP/Spain-Men-Stats",
	"FRA": "FRA/France-Men-Stats",
	"GER": "GER/Germany-Men-Stats",
	"GHA": "GHA/Ghana-Men-Stats",
	"HAI": "HAI/Haiti-Men-Stats",
	"IRN": "IRN/Iran-Men-Stats",
	"IRQ": "IRQ/Iraq-Men-Stats",
	"JOR": "JOR/Jordan-Men-Stats",
	"JPN": "JPN/Japan-Men-Stats",
	"KOR": "KOR/Korea-Republic-Men-Stats",
	"KSA": "KSA/Saudi-Arabia-Men-Stats",
	"MAR": "MAR/Morocco-Men-Stats",
	"MEX": "MEX/Mexico-Men-Stats",
	"NED": "NED/Netherlands-Men-Stats",
	"NOR": "NOR/Norway-Men-Stats",
	"NZL": "NZL/New-Zealand-Men-Stats",
	"PAN": "PAN/Panama-Men-Stats",
	"PAR": "PAR/Paraguay-Men-Stats",
	"POR": "POR/Portugal-Men-Stats",
	"QAT": "QAT/Qatar-Men-Stats",
	"RSA": "RSA/South-Africa-Men-Stats",
	"SCO": "SCO/Scotland-Men-Stats",
	"SEN": "SEN/Senegal-Men-Stats",
	"SUI": "SUI/Switzerland-Men-Stats",
	"SWE": "SWE/Sweden-Men-Stats",
	"TUN": "TUN/Tunisia-Men-Stats",
	"TUR": "TUR/Turkey-Men-Stats",
	"URU": "URU/Uruguay-Men-Stats",
	"USA": "USA/United-States-Men-Stats",
	"UZB": "UZB/Uzbekistan-Men-Stats",
}

// FBrefSlug returns the FBref URL suffix for a catalog team ID.
func FBrefSlug(teamID string) (string, bool) {
	s, ok := fbrefSlugs[teamID]
	return s, ok
}
