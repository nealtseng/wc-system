package catalog

import "strings"

// fmNatAliases maps team ID to extra FM26 "Nat" column values beyond the team ID itself.
var fmNatAliases = map[string][]string{
	"KOR": {"SKO", "Korea Republic", "South Korea"},
	"CIV": {"IVC", "Cote d'Ivoire", "Côte d'Ivoire", "Ivory Coast"},
	"USA": {"United States"},
	"TUR": {"Turkey", "Turkiye"},
	"CUW": {"Curacao", "Curaçao", "CUW"},
	"COD": {"DR Congo", "Congo DR", "COD"},
	"CPV": {"Cape Verde", "Cabo Verde"},
	"KSA": {"Saudi Arabia", "KSA"},
	"RSA": {"South Africa", "RSA"},
	"NZL": {"New Zealand", "NZL"},
	"BIH": {"Bosnia and Herzegovina", "Bosnia"},
	"SCO": {"Scotland", "SCO"},
	"ENG": {"England", "ENG"},
}

// FMNatMatchesAny reports whether any nationality token in nat belongs to teamID.
// nat may be a single value or comma-separated dual citizenship, e.g. "England,Republic of Ireland".
func FMNatMatchesAny(teamID, nat string) bool {
	nat = strings.TrimSpace(nat)
	if nat == "" {
		return false
	}
	if FMNatMatches(teamID, nat) {
		return true
	}
	if strings.Contains(nat, ",") {
		for _, part := range strings.Split(nat, ",") {
			if FMNatMatches(teamID, strings.TrimSpace(part)) {
				return true
			}
		}
	}
	return false
}

// FMNatMatches reports whether a FM26 export Nat value belongs to teamID.
func FMNatMatches(teamID, nat string) bool {
	team, ok := ByID(teamID)
	if !ok {
		return strings.EqualFold(strings.TrimSpace(teamID), strings.TrimSpace(nat))
	}

	nat = strings.TrimSpace(nat)
	if nat == "" {
		return false
	}

	candidates := []string{
		team.ID,
		team.ISO2,
		team.FIFACountryName,
		team.Name,
		team.KaggleName,
	}
	candidates = append(candidates, fmNatAliases[team.ID]...)

	natUpper := strings.ToUpper(nat)
	for _, c := range candidates {
		if strings.EqualFold(c, nat) || strings.ToUpper(c) == natUpper {
			return true
		}
	}
	return false
}
