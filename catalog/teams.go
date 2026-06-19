package catalog

// Team is a supported national side with identifiers for each data source.
type Team struct {
	ID              string
	Name            string
	ISO2            string
	WikiSlug        string
	KaggleName      string
	FIFACountryName string
	Group           string // "A" through "L"
}

// All returns the canonical team list used across adapters and APIs.
func All() []Team {
	return []Team{
		// Group A
		{ID: "MEX", Name: "Mexico", ISO2: "mx", Group: "A", WikiSlug: "Mexico_national_football_team", KaggleName: "Mexico", FIFACountryName: "Mexico"},
		{ID: "RSA", Name: "South Africa", ISO2: "za", Group: "A", WikiSlug: "South_Africa_national_football_team", KaggleName: "South Africa", FIFACountryName: "South Africa"},
		{ID: "KOR", Name: "South Korea", ISO2: "kr", Group: "A", WikiSlug: "South_Korea_national_football_team", KaggleName: "South Korea", FIFACountryName: "Korea Republic"},
		{ID: "CZE", Name: "Czech Republic", ISO2: "cz", Group: "A", WikiSlug: "Czech_Republic_national_football_team", KaggleName: "Czech Republic", FIFACountryName: "Czech Republic"},
		// Group B
		{ID: "CAN", Name: "Canada", ISO2: "ca", Group: "B", WikiSlug: "Canada_men%27s_national_soccer_team", KaggleName: "Canada", FIFACountryName: "Canada"},
		{ID: "SUI", Name: "Switzerland", ISO2: "ch", Group: "B", WikiSlug: "Switzerland_national_football_team", KaggleName: "Switzerland", FIFACountryName: "Switzerland"},
		{ID: "QAT", Name: "Qatar", ISO2: "qa", Group: "B", WikiSlug: "Qatar_national_football_team", KaggleName: "Qatar", FIFACountryName: "Qatar"},
		{ID: "BIH", Name: "Bosnia and Herzegovina", ISO2: "ba", Group: "B", WikiSlug: "Bosnia_and_Herzegovina_national_football_team", KaggleName: "Bosnia and Herzegovina", FIFACountryName: "Bosnia and Herzegovina"},
		// Group C
		{ID: "MAR", Name: "Morocco", ISO2: "ma", Group: "C", WikiSlug: "Morocco_national_football_team", KaggleName: "Morocco", FIFACountryName: "Morocco"},
		{ID: "HAI", Name: "Haiti", ISO2: "ht", Group: "C", WikiSlug: "Haiti_national_football_team", KaggleName: "Haiti", FIFACountryName: "Haiti"},
		{ID: "SCO", Name: "Scotland", ISO2: "gb-sct", Group: "C", WikiSlug: "Scotland_national_football_team", KaggleName: "Scotland", FIFACountryName: "Scotland"},
		{ID: "BRA", Name: "Brazil", ISO2: "BR", Group: "C", WikiSlug: "Brazil_national_football_team", KaggleName: "Brazil", FIFACountryName: "Brazil"},
		// Group D
		{ID: "PAR", Name: "Paraguay", ISO2: "py", Group: "D", WikiSlug: "Paraguay_national_football_team", KaggleName: "Paraguay", FIFACountryName: "Paraguay"},
		{ID: "AUS", Name: "Australia", ISO2: "au", Group: "D", WikiSlug: "Australia_men%27s_national_soccer_team", KaggleName: "Australia", FIFACountryName: "Australia"},
		{ID: "TUR", Name: "Turkiye", ISO2: "tr", Group: "D", WikiSlug: "Turkey_national_football_team", KaggleName: "Turkey", FIFACountryName: "Turkiye"},
		{ID: "USA", Name: "USA", ISO2: "US", Group: "D", WikiSlug: "United_States_men%27s_national_soccer_team", KaggleName: "United States", FIFACountryName: "United States"},
		// Group E
		{ID: "CUW", Name: "Curaçao", ISO2: "cw", Group: "E", WikiSlug: "Cura%C3%A7ao_national_football_team", KaggleName: "Curacao", FIFACountryName: "Curaçao"},
		{ID: "CIV", Name: "Ivory Coast", ISO2: "ci", Group: "E", WikiSlug: "Ivory_Coast_national_football_team", KaggleName: "Ivory Coast", FIFACountryName: "Ivory Coast"},
		{ID: "ECU", Name: "Ecuador", ISO2: "ec", Group: "E", WikiSlug: "Ecuador_national_football_team", KaggleName: "Ecuador", FIFACountryName: "Ecuador"},
		{ID: "GER", Name: "Germany", ISO2: "DE", Group: "E", WikiSlug: "Germany_national_football_team", KaggleName: "Germany", FIFACountryName: "Germany"},
		// Group F
		{ID: "TUN", Name: "Tunisia", ISO2: "tn", Group: "F", WikiSlug: "Tunisia_national_football_team", KaggleName: "Tunisia", FIFACountryName: "Tunisia"},
		{ID: "SWE", Name: "Sweden", ISO2: "se", Group: "F", WikiSlug: "Sweden_men%27s_national_football_team", KaggleName: "Sweden", FIFACountryName: "Sweden"},
		{ID: "NED", Name: "Netherlands", ISO2: "NL", Group: "F", WikiSlug: "Netherlands_national_football_team", KaggleName: "Netherlands", FIFACountryName: "Netherlands"},
		{ID: "JPN", Name: "Japan", ISO2: "JP", Group: "F", WikiSlug: "Japan_national_football_team", KaggleName: "Japan", FIFACountryName: "Japan"},
		// Group G
		{ID: "BEL", Name: "Belgium", ISO2: "be", Group: "G", WikiSlug: "Belgium_national_football_team", KaggleName: "Belgium", FIFACountryName: "Belgium"},
		{ID: "EGY", Name: "Egypt", ISO2: "eg", Group: "G", WikiSlug: "Egypt_national_football_team", KaggleName: "Egypt", FIFACountryName: "Egypt"},
		{ID: "IRN", Name: "Iran", ISO2: "ir", Group: "G", WikiSlug: "Iran_national_football_team", KaggleName: "Iran", FIFACountryName: "Iran"},
		{ID: "NZL", Name: "New Zealand", ISO2: "nz", Group: "G", WikiSlug: "New_Zealand_national_football_team", KaggleName: "New Zealand", FIFACountryName: "New Zealand"},
		// Group H
		{ID: "CPV", Name: "Cape Verde", ISO2: "cv", Group: "H", WikiSlug: "Cape_Verde_national_football_team", KaggleName: "Cape Verde", FIFACountryName: "Cape Verde"},
		{ID: "KSA", Name: "Saudi Arabia", ISO2: "sa", Group: "H", WikiSlug: "Saudi_Arabia_national_football_team", KaggleName: "Saudi Arabia", FIFACountryName: "Saudi Arabia"},
		{ID: "URU", Name: "Uruguay", ISO2: "uy", Group: "H", WikiSlug: "Uruguay_national_football_team", KaggleName: "Uruguay", FIFACountryName: "Uruguay"},
		{ID: "ESP", Name: "Spain", ISO2: "ES", Group: "H", WikiSlug: "Spain_national_football_team", KaggleName: "Spain", FIFACountryName: "Spain"},
		// Group I
		{ID: "SEN", Name: "Senegal", ISO2: "sn", Group: "I", WikiSlug: "Senegal_national_football_team", KaggleName: "Senegal", FIFACountryName: "Senegal"},
		{ID: "NOR", Name: "Norway", ISO2: "no", Group: "I", WikiSlug: "Norway_national_football_team", KaggleName: "Norway", FIFACountryName: "Norway"},
		{ID: "IRQ", Name: "Iraq", ISO2: "iq", Group: "I", WikiSlug: "Iraq_national_football_team", KaggleName: "Iraq", FIFACountryName: "Iraq"},
		{ID: "FRA", Name: "France", ISO2: "FR", Group: "I", WikiSlug: "France_national_football_team", KaggleName: "France", FIFACountryName: "France"},
		// Group J
		{ID: "ALG", Name: "Algeria", ISO2: "dz", Group: "J", WikiSlug: "Algeria_national_football_team", KaggleName: "Algeria", FIFACountryName: "Algeria"},
		{ID: "AUT", Name: "Austria", ISO2: "at", Group: "J", WikiSlug: "Austria_national_football_team", KaggleName: "Austria", FIFACountryName: "Austria"},
		{ID: "JOR", Name: "Jordan", ISO2: "jo", Group: "J", WikiSlug: "Jordan_national_football_team", KaggleName: "Jordan", FIFACountryName: "Jordan"},
		{ID: "ARG", Name: "Argentina", ISO2: "AR", Group: "J", WikiSlug: "Argentina_national_football_team", KaggleName: "Argentina", FIFACountryName: "Argentina"},
		// Group K
		{ID: "COL", Name: "Colombia", ISO2: "co", Group: "K", WikiSlug: "Colombia_national_football_team", KaggleName: "Colombia", FIFACountryName: "Colombia"},
		{ID: "UZB", Name: "Uzbekistan", ISO2: "uz", Group: "K", WikiSlug: "Uzbekistan_national_football_team", KaggleName: "Uzbekistan", FIFACountryName: "Uzbekistan"},
		{ID: "COD", Name: "Congo DR", ISO2: "cd", Group: "K", WikiSlug: "DR_Congo_national_football_team", KaggleName: "DR Congo", FIFACountryName: "DR Congo"},
		{ID: "POR", Name: "Portugal", ISO2: "PT", Group: "K", WikiSlug: "Portugal_national_football_team", KaggleName: "Portugal", FIFACountryName: "Portugal"},
		// Group L
		{ID: "CRO", Name: "Croatia", ISO2: "hr", Group: "L", WikiSlug: "Croatia_national_football_team", KaggleName: "Croatia", FIFACountryName: "Croatia"},
		{ID: "GHA", Name: "Ghana", ISO2: "gh", Group: "L", WikiSlug: "Ghana_national_football_team", KaggleName: "Ghana", FIFACountryName: "Ghana"},
		{ID: "PAN", Name: "Panama", ISO2: "pa", Group: "L", WikiSlug: "Panama_national_football_team", KaggleName: "Panama", FIFACountryName: "Panama"},
		{ID: "ENG", Name: "England", ISO2: "gb-eng", Group: "L", WikiSlug: "England_national_football_team", KaggleName: "England", FIFACountryName: "England"},
	}
}

// ByID returns the team for id or false.
func ByID(id string) (Team, bool) {
	for _, t := range All() {
		if t.ID == id {
			return t, true
		}
	}
	return Team{}, false
}

// KaggleNames returns the set of country names used in the martj42 results CSV.
func KaggleNames() map[string]struct{} {
	names := make(map[string]struct{}, len(All()))
	for _, t := range All() {
		names[t.KaggleName] = struct{}{}
	}
	return names
}

// FIFACountryByTeamID maps catalog team IDs to FIFA CSV country_name values.
func FIFACountryByTeamID() map[string]string {
	out := make(map[string]string, len(All()))
	for _, t := range All() {
		out[t.ID] = t.FIFACountryName
	}
	return out
}
