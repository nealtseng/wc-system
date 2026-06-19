package main

import (
	"context"
	"fmt"
	"log"
	_ "time/tzdata" // embed IANA zones for Alpine/Docker (LoadLocation)

	"wc-system/api"
	"wc-system/config"
	"wc-system/db"
	"wc-system/service/calibration"
	"wc-system/service/teamdata"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("main: .env not found — using environment variables directly")
	}

	cfg := config.Load()
	ctx := context.Background()

	pool := db.InitPool(ctx, cfg.AivenDBURL)
	db.RunMigrations(ctx, pool)

	store := teamdata.NewStore(pool, cfg.FifaCSVURL, cfg.FMCSVDir, cfg.FMCSVFile, cfg.TheOddsAPIKey, cfg.FBrefXGCsv)
	if err := store.LoadLatestOddsFromDB(ctx); err != nil {
		log.Printf("main: load odds from DB: %v", err)
	}
	if cfg.FMCSVFile != "" {
		log.Printf("main: FM CSV enabled → %s", cfg.FMCSVFile)
	} else if cfg.FMCSVDir != "" {
		log.Printf("main: FM CSV dir enabled → %s", cfg.FMCSVDir)
	}
	if err := store.BootstrapXG(ctx); err != nil {
		log.Printf("main: bootstrap xG: %v", err)
	}
	go func() {
		log.Println("main: syncing World Bank, Wikimedia, Kaggle, and FIFA data…")
		if err := store.SyncAll(ctx); err != nil {
			log.Printf("main: initial sync completed with warnings: %v", err)
		} else {
			log.Printf("main: initial sync complete (%d teams)", len(store.List()))
		}
		if pool != nil {
			go func() {
				result, err := calibration.Calibrate(ctx, pool, store)
				if err != nil {
					log.Printf("main: auto-calibrate: %v", err)
					return
				}
				if err := store.SaveWeights(ctx, teamdata.ModelWeights{
					W1: result.W1, W2: result.W2, W3: result.W3,
					ClipDelta:  store.Weights().ClipDelta,
					DeltaMax:   store.Weights().DeltaMax,
					KellyScale: store.Weights().KellyScale,
				}); err != nil {
					log.Printf("main: save calibrated weights: %v", err)
				} else {
					log.Printf("main: calibrated weights w1=%.2f w2=%.2f w3=%.2f (Brier=%.4f, n=%d)",
						result.W1, result.W2, result.W3, result.BrierScore, result.MatchesUsed)
				}
			}()
		}
	}()
	go store.RunOddsScheduler(ctx)

	router := api.NewRouter(pool, cfg, store)

	addr := fmt.Sprintf(":%s", cfg.BackendPort)
	log.Printf("main: wc-system backend listening on %s", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("main: server error: %v", err)
	}
}
