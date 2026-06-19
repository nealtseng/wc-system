package fmcsv_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"wc-system/adapter/fmcsv"
)

const sampleCSV = `Name;Age;Position;Nat;Ability
Kylian Mbappé;27;AMR, ST;FRA;4,5
Antoine Griezmann;34;AMC;FRA;4,0
Test Player;25;DC;FRA;3,5
`

func TestParseReader(t *testing.T) {
	records, err := fmcsv.ParseReader(strings.NewReader(sampleCSV), fmcsv.FMCSVConfig{})
	if err != nil {
		t.Fatalf("ParseReader: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("got %d records, want 3", len(records))
	}
	if records[0].Name != "Kylian Mbappé" {
		t.Fatalf("name = %q", records[0].Name)
	}
	if records[0].Overall != 90 {
		t.Fatalf("overall = %v, want 90", records[0].Overall)
	}
	if records[1].Overall != 80 {
		t.Fatalf("overall = %v, want 80", records[1].Overall)
	}
}

func TestParseFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "FRA.csv")
	if err := os.WriteFile(path, []byte(sampleCSV), 0o644); err != nil {
		t.Fatal(err)
	}

	records, err := fmcsv.Parse(fmcsv.FMCSVConfig{FilePath: path})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("got %d records, want 3", len(records))
	}
}

func TestParseMissingColumn(t *testing.T) {
	_, err := fmcsv.ParseReader(strings.NewReader("Name;Age\nA;1"), fmcsv.FMCSVConfig{
		NameCol: "Name", AgeCol: "Age", PositionCol: "Position", NatCol: "Nat", AbilityCol: "Ability",
	})
	if !errors.Is(err, fmcsv.ErrColumnNotFound) {
		t.Fatalf("expected ErrColumnNotFound, got: %v", err)
	}
}

const sampleZhCSV = `姓名,位置,年龄,ca,pa,国籍,俱乐部,角球,传中,盘带,射门,接球,技术,UID,防守,身体,速度,创造,进攻,技术,制空,精神
Kylian Mbappé,AM/ST RL,26,191,197,France,R. Madrid,13,13,18,18,18,17,85139014,4.00,14.75,19.00,16.33,17.33,17.67,8.50,14.83
Harry Kane,ST C,32,188,190,"England,Republic of Ireland",FC Bayern,11,12,14,19,18,17,28049320,8.33,15.50,12.50,16.00,18.00,16.33,16.50,17.00
`

func TestParseChineseCSV(t *testing.T) {
	records, err := fmcsv.ParseReader(strings.NewReader(sampleZhCSV), fmcsv.FMCSVConfig{})
	if err != nil {
		t.Fatalf("ParseReader: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("got %d records, want 2", len(records))
	}
	if records[0].CA != 191 {
		t.Fatalf("CA = %d, want 191", records[0].CA)
	}
	if records[0].Overall != 95.5 {
		t.Fatalf("Overall = %v, want 95.5", records[0].Overall)
	}
	if records[0].Attributes["盘带"] != 18 {
		t.Fatalf("盘带 = %d, want 18", records[0].Attributes["盘带"])
	}
	if records[0].AttrGroups["速度"] != 19 {
		t.Fatalf("速度 group = %v, want 19", records[0].AttrGroups["速度"])
	}
}

func TestParsePlayersFile(t *testing.T) {
	path := filepath.Join("..", "..", "data", "fm", "players.csv")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("players.csv not present")
	}
	records, err := fmcsv.Parse(fmcsv.FMCSVConfig{FilePath: path})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(records) < 100 {
		t.Fatalf("expected 100+ players, got %d", len(records))
	}
}
