// Package fmcsv parses Football Manager player export CSV files.
// Supports FM26PlayerExport v4+ (English, semicolon, star ratings) and
// Chinese FM scout exports (comma-delimited, CA/PA attributes).
package fmcsv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// ErrColumnNotFound is returned when a required column name is absent from the CSV header.
var ErrColumnNotFound = errors.New("fmcsv: required column not found in header")

// FMCSVConfig specifies the file path and optional column name overrides.
// Empty column names are resolved automatically from the CSV header.
type FMCSVConfig struct {
	FilePath    string
	NameCol     string
	AgeCol      string
	PositionCol string
	NatCol      string
	AbilityCol  string
}

// PlayerRecord is one player row normalised into application-level types.
type PlayerRecord struct {
	Name        string
	Age         int
	Position    string
	Nationality string
	Club        string
	// Overall is normalised ability on a 0–100 scale (CA/2 or stars×20).
	Overall    float64
	CA         int
	PA         int
	RCA        int
	Apps       int
	Goals      int
	Value      string
	Height     int
	Attributes map[string]int
	AttrGroups map[string]float64
}

var technicalAttrs = []string{
	"角球", "传中", "盘带", "射门", "接球", "任意球", "头球", "远射", "界外球",
	"盯人", "传球", "罚点球", "抢断", "技术",
}
var mentalAttrs = []string{
	"侵略性", "预判", "勇敢", "镇定", "集中", "决断", "意志力", "想象力", "领导力",
	"无球跑动", "防守站位", "团队合作", "视野", "工作投入",
}
var physicalAttrs = []string{
	"爆发力", "灵活", "平衡", "弹跳", "体质", "速度", "耐力", "强壮",
}
var goalkeeperAttrs = []string{
	"制空能力", "拦截传中", "沟通", "神经指数", "手控球", "大脚开球", "一对一",
	"反应", "出击", "击球倾向", "手抛球的能力",
}
var groupAttrs = []string{"防守", "身体", "速度", "创造", "进攻", "技术", "制空", "精神"}

// Parse opens the CSV at cfg.FilePath and returns normalised player records.
func Parse(cfg FMCSVConfig) ([]PlayerRecord, error) {
	f, err := os.Open(cfg.FilePath)
	if err != nil {
		return nil, fmt.Errorf("fmcsv: opening file %q: %w", cfg.FilePath, err)
	}
	defer f.Close()
	return ParseReader(f, cfg)
}

// ParseReader reads FM export rows from r.
func ParseReader(r io.Reader, cfg FMCSVConfig) ([]PlayerRecord, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("fmcsv: read: %w", err)
	}

	delim := detectDelimiter(string(data))
	reader := csv.NewReader(strings.NewReader(string(data)))
	reader.Comma = delim
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("fmcsv: reading header: %w", err)
	}

	resolved, format := resolveColumns(header, cfg)
	if err := validateColumns(header, resolved); err != nil {
		return nil, err
	}

	var records []PlayerRecord
	for {
		row, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("fmcsv: reading row: %w", err)
		}
		if len(row) == 0 {
			continue
		}
		rec := rowToRecord(row, header, resolved, format)
		if rec.Name == "" {
			continue
		}
		records = append(records, rec)
	}
	return records, nil
}

func detectDelimiter(sample string) rune {
	sample = strings.TrimPrefix(sample, "\ufeff")
	if strings.Contains(sample, "姓名") || strings.Contains(sample, ",ca,") {
		return ','
	}
	if strings.Count(sample, ";") >= strings.Count(sample, ",") {
		return ';'
	}
	return ','
}

func resolveColumns(header []string, cfg FMCSVConfig) (FMCSVConfig, string) {
	if cfg.NameCol != "" {
		return cfg, "custom"
	}
	if colIndex(header, "姓名") >= 0 || colIndex(header, "ca") >= 0 {
		return FMCSVConfig{
			NameCol: "姓名", AgeCol: "年龄", PositionCol: "位置",
			NatCol: "国籍", AbilityCol: "ca",
		}, "zh_ca"
	}
	return FMCSVConfig{
		NameCol: "Name", AgeCol: "Age", PositionCol: "Position",
		NatCol: "Nat", AbilityCol: "Ability",
	}, "en_stars"
}

func validateColumns(header []string, cfg FMCSVConfig) error {
	for _, col := range []string{cfg.NameCol, cfg.AgeCol, cfg.PositionCol, cfg.NatCol, cfg.AbilityCol} {
		if colIndex(header, col) < 0 {
			return fmt.Errorf("%w: %q", ErrColumnNotFound, col)
		}
	}
	return nil
}

func trimHeader(s string) string {
	s = strings.TrimSpace(s)
	return strings.TrimPrefix(s, "\ufeff")
}

func colIndex(header []string, name string) int {
	name = trimHeader(name)
	for i, h := range header {
		if trimHeader(h) == name {
			return i
		}
	}
	return -1
}

func colIndexAfter(header []string, name string, after int) int {
	name = trimHeader(name)
	for i := after; i < len(header); i++ {
		if trimHeader(header[i]) == name {
			return i
		}
	}
	return -1
}

func cell(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

func parseInt(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, _ := strconv.Atoi(s)
	return v
}

func parseFloat(s string) float64 {
	s = strings.ReplaceAll(strings.TrimSpace(s), ",", ".")
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func rowToRecord(row []string, header []string, cfg FMCSVConfig, format string) PlayerRecord {
	rec := PlayerRecord{
		Attributes: make(map[string]int),
		AttrGroups: make(map[string]float64),
	}
	rec.Name = cell(row, colIndex(header, cfg.NameCol))
	rec.Age = parseInt(cell(row, colIndex(header, cfg.AgeCol)))
	rec.Position = cell(row, colIndex(header, cfg.PositionCol))
	rec.Nationality = cell(row, colIndex(header, cfg.NatCol))
	rec.Club = cell(row, colIndex(header, "俱乐部"))
	rec.Value = cell(row, colIndex(header, "身价（欧元）"))
	rec.Height = parseInt(cell(row, colIndex(header, "身高")))
	rec.Apps = parseInt(cell(row, colIndex(header, "国家队出场")))
	rec.Goals = parseInt(cell(row, colIndex(header, "国家队进球")))
	rec.CA = parseInt(cell(row, colIndex(header, "ca")))
	rec.PA = parseInt(cell(row, colIndex(header, "pa")))
	rec.RCA = parseInt(cell(row, colIndex(header, "RCA")))

	switch format {
	case "zh_ca":
		if rec.CA > 0 {
			rec.Overall = float64(rec.CA) / 2.0
		} else if rec.RCA > 0 {
			rec.Overall = float64(rec.RCA) / 2.0
		}
	default:
		starStr := strings.ReplaceAll(cell(row, colIndex(header, cfg.AbilityCol)), ",", ".")
		stars := parseFloat(starStr)
		rec.Overall = stars * 20
		if rec.CA == 0 && rec.Overall > 0 {
			rec.CA = int(rec.Overall * 2)
		}
	}

	for _, cols := range [][]string{technicalAttrs, mentalAttrs, physicalAttrs, goalkeeperAttrs} {
		for _, col := range cols {
			idx := colIndex(header, col)
			if idx < 0 {
				continue
			}
			if v := parseInt(cell(row, idx)); v > 0 {
				rec.Attributes[col] = v
			}
		}
	}

	uidIdx := colIndex(header, "UID")
	searchFrom := uidIdx
	if searchFrom < 0 {
		searchFrom = 0
	}
	for _, g := range groupAttrs {
		idx := colIndexAfter(header, g, searchFrom)
		if idx < 0 {
			continue
		}
		if v := parseFloat(cell(row, idx)); v > 0 {
			rec.AttrGroups[g] = v
		}
	}

	return rec
}
