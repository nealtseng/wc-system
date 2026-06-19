package worldcup26

import (
	"testing"
	"time"
)

func TestStadiumLocation_MexicoCity(t *testing.T) {
	loc := StadiumLocation("Estadio Azteca")
	kickoff, err := time.ParseInLocation("01/02/2006 15:04", "06/11/2026 13:00", loc)
	if err != nil {
		t.Fatal(err)
	}
	utc := kickoff.UTC()
	if utc.Hour() != 19 || utc.Minute() != 0 {
		t.Fatalf("expected 19:00 UTC for Mexico City 13:00 in June, got %v", utc)
	}
}

func TestStadiumLocation_LoadsInMinimalEnv(t *testing.T) {
	loc := StadiumLocation("Lumen Field")
	if loc.String() == "UTC" {
		t.Fatal("expected America/Los_Angeles, got UTC fallback — import _ \"time/tzdata\" in main")
	}
}
