package worldcup26

import "time"

// StadiumLocation returns the IANA timezone for a venue name from football.stadiums.json.
// local_date in the CDN is wall-clock time at the host stadium, not UTC.
func StadiumLocation(stadiumName string) *time.Location {
	tz, ok := stadiumTimezones[stadiumName]
	if !ok {
		return time.UTC
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.UTC
	}
	return loc
}

var stadiumTimezones = map[string]string{
	"Estadio Azteca":                   "America/Mexico_City",
	"Estadio Akron":                    "America/Mexico_City",
	"Estadio BBVA":                     "America/Monterrey",
	"AT&T Stadium":                     "America/Chicago",
	"NRG Stadium":                      "America/Chicago",
	"GEHA Field at Arrowhead Stadium":  "America/Chicago",
	"Mercedes-Benz Stadium":            "America/New_York",
	"Hard Rock Stadium":                "America/New_York",
	"Gillette Stadium":                 "America/New_York",
	"Lincoln Financial Field":          "America/New_York",
	"MetLife Stadium":                  "America/New_York",
	"BMO Field":                        "America/Toronto",
	"BC Place":                         "America/Vancouver",
	"Lumen Field":                      "America/Los_Angeles",
	"Levi's Stadium":                   "America/Los_Angeles",
	"SoFi Stadium":                     "America/Los_Angeles",
}
