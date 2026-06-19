package config

import (
	"os"
	"strings"
)

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	LLMAPIKey       string
	LLMBaseURL      string
	LLMModel        string
	AivenDBURL      string
	APIFootballKey  string
	TheOddsAPIKey   string
	FifaCSVURL      string
	FMCSVDir        string
	FMCSVFile       string
	FBrefXGCsv      string
	BackendPort     string
	PostgresDB      string
	PostgresUser    string
	PostgresPass    string
	PostgresPort    string
}

// Load reads environment variables (already loaded by godotenv in main) and
// returns a populated Config. Missing optional fields are left as empty string.
func Load() *Config {
	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = "8080"
	}
	fmCSVFile := strings.TrimSpace(os.Getenv("FM_CSV_FILE"))
	if fmCSVFile == "" {
		fmCSVFile = defaultFMCSVFile()
	}
	return &Config{
		LLMAPIKey:      firstNonEmpty(os.Getenv("LLM_API_KEY"), os.Getenv("DEEPSEEK_API_KEY")),
		LLMBaseURL:     os.Getenv("LLM_BASE_URL"),
		LLMModel:       os.Getenv("LLM_MODEL"),
		AivenDBURL:     os.Getenv("AIVEN_DB_URL"),
		APIFootballKey: os.Getenv("API_FOOTBALL_KEY"),
		TheOddsAPIKey:  os.Getenv("THE_ODDS_API_KEY"),
		FifaCSVURL:     os.Getenv("FIFA_CSV_URL"),
		FMCSVDir:       os.Getenv("FM_CSV_DIR"),
		FMCSVFile:      fmCSVFile,
		FBrefXGCsv:     os.Getenv("FBREF_XG_CSV"),
		BackendPort:    port,
		PostgresDB:     os.Getenv("POSTGRES_DB"),
		PostgresUser:   os.Getenv("POSTGRES_USER"),
		PostgresPass:   os.Getenv("POSTGRES_PASSWORD"),
		PostgresPort:   os.Getenv("POSTGRES_PORT"),
	}
}

func defaultFMCSVFile() string {
	const path = "data/fm/players.csv"
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}
