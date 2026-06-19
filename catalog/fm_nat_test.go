package catalog_test

import (
	"testing"

	"wc-system/catalog"
)

func TestFMNatMatches(t *testing.T) {
	tests := []struct {
		teamID string
		nat    string
		want   bool
	}{
		{"FRA", "FRA", true},
		{"FRA", "fra", true},
		{"KOR", "Korea Republic", true},
		{"KOR", "SKO", true},
		{"USA", "United States", true},
		{"FRA", "BRA", false},
		{"ENG", "England,Republic of Ireland", false},
	}
	for _, tt := range tests {
		got := catalog.FMNatMatches(tt.teamID, tt.nat)
		if got != tt.want {
			t.Errorf("FMNatMatches(%q, %q) = %v, want %v", tt.teamID, tt.nat, got, tt.want)
		}
	}
}

func TestFMNatMatchesAny(t *testing.T) {
	if !catalog.FMNatMatchesAny("ENG", "England,Republic of Ireland") {
		t.Error("expected ENG to match England in dual nationality")
	}
	if !catalog.FMNatMatchesAny("FRA", "France,Mauritania") {
		t.Error("expected FRA to match France in dual nationality")
	}
	if catalog.FMNatMatchesAny("BRA", "England,Republic of Ireland") {
		t.Error("expected BRA not to match England")
	}
}
