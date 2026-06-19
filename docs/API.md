# WC System Backend API Reference

Base URL (local dev): `http://localhost:8080`

All endpoints return `Content-Type: application/json`.
CORS is open for `http://localhost:3000` (Vite dev server).

---

## GET /api/health

Health check.

**Response 200:**
```json
{ "status": "ok" }
```

---

## GET /api/teams

Returns all teams with their ELO ratings and World Cup 2026 metadata.

**Response 200:**
```json
[
  {
    "id": "FRA",
    "name": "France",
    "elo": 2110,
    "wc_group": "I",
    "iso2": "fr",
    "gdp_per_capita": 42000,
    "win_rate": 0.65,
    "momentum": 0.12
  },
  {
    "id": "ESP",
    "name": "Spain",
    "elo": 2045,
    "wc_group": "H",
    "iso2": "es"
  }
]
```

| Field | Type | Description |
|-------|------|-------------|
| `wc_group` | string | World Cup 2026 group letter (`A`–`L`) |
| `iso2` | string | ISO 3166-1 alpha-2 code for flag rendering (e.g. `gb-eng` for England) |

---

## GET /api/matches

Returns all World Cup 2026 fixtures synced from the worldcup26 GitHub CDN, ordered by kickoff time.

**Response 200:**
```json
[
  {
    "id": 1,
    "wc_match_id": "1",
    "home_id": "MEX",
    "away_id": "RSA",
    "home_iso2": "mx",
    "away_iso2": "za",
    "home_name": "Mexico",
    "away_name": "South Africa",
    "stadium_name": "Estadio Azteca",
    "stage": "group",
    "matchday": 1,
    "local_date": "06/11/2026 13:00",
    "kickoff": "2026-06-11T13:00:00Z",
    "home_score": null,
    "away_score": null,
    "finished": false,
    "time_elapsed": "notstarted",
    "status": "scheduled"
  }
]
```

| Field | Type | Description |
|-------|------|-------------|
| `wc_match_id` | string | Fixture ID from worldcup26 dataset |
| `home_id` / `away_id` | string | FIFA catalog team codes |
| `home_iso2` / `away_iso2` | string | ISO codes joined from `teams` table |
| `stage` | string | Match type: `group`, `r32`, `r16`, `qf`, `sf`, `final` |
| `local_date` | string | Local kickoff string (`MM/DD/YYYY HH:MM`) |
| `kickoff` | string | UTC kickoff in RFC3339 format |
| `finished` | boolean | Whether the match has ended (from worldcup26 CDN) |
| `time_elapsed` | string | Raw status from CDN (`notstarted`, `45'`, etc.) |
| `status` | string | Normalized: `scheduled`, `live`, `finished` |

**Live scores source:** `rezarahiminia/worldcup2026` GitHub raw CDN (`football.matches.json`). Re-sync via `POST /api/pipeline/sync` to refresh scores during the tournament. No API key required.

**Error 500:** database query failure.

---

## POST /api/predict

Runs ELO + Poisson Monte Carlo prediction for a match.

**Request:**
```json
{ "home_team_id": "FRA", "away_team_id": "ESP" }
```

**Response 200:**
```json
{
  "home_team_id": "FRA",
  "away_team_id": "ESP",
  "elo": { "home": 2110.0, "away": 2045.0 },
  "poisson": {
    "home_lambda": 1.24,
    "away_lambda": 0.91,
    "home_win": 0.42,
    "draw": 0.28,
    "away_win": 0.30,
    "top_scores": [
      { "home": 1, "away": 0, "prob": 0.123 },
      { "home": 2, "away": 1, "prob": 0.105 },
      { "home": 1, "away": 1, "prob": 0.098 },
      { "home": 0, "away": 0, "prob": 0.082 }
    ]
  },
  "weights": { "w1": 0.30, "w2": 0.30, "w3": 0.40, "clip": 0.05 },
  "p_final": { "home": 0.380, "draw": 0.290, "away": 0.330 },
  "sources": {
    "elo": "kaggle",
    "xg": "fbref",
    "gdp": "worldbank",
    "wiki": "wikimedia",
    "w1_delta": 0.02,
    "w3": "odds"
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `weights` | object | Dynamic weights from `model_config` DB table (not hardcoded) |
| `sources.w3` | string | `"odds"` when The Odds API data is available for this fixture; `"elo_fallback"` otherwise |
| `sources.xg` | string | `"fbref"` when FBref WC qualifier xG is synced; `"kaggle"` fallback |

**p_final computation:**
```
W3 = de-vigged implied probabilities from latest odds (or ELO fallback)
p_home = clip(w1_micro_delta) + poisson.home_win * w2 + w3_home * w3
p_draw = poisson.draw * w2 + w3_draw * w3
p_away = poisson.away_win * w2 + w3_away * w3 - clip(w1_micro_delta)
```
All three values are normalised to sum to 1.0. Weights (`w1`, `w2`, `w3`) are loaded from `model_config`.

**Error 400:** missing or invalid JSON body.

---

## POST /api/narrative

Calls DeepSeek V4 Flash to generate a Traditional Chinese match narrative
with a confidence float.  Always returns 200; uses a fallback message if
the API key is absent or the service is unavailable.

**Request:**
```json
{
  "home_team": "France",
  "away_team": "Spain",
  "home_win_prob": 0.380,
  "draw_prob": 0.290,
  "away_win_prob": 0.330,
  "home_elo": 2110,
  "away_elo": 2045,
  "home_gdp": 42000,
  "away_gdp": 30000,
  "home_lambda": 1.24,
  "away_lambda": 0.91,
  "w1": 0.30,
  "w2": 0.30,
  "w3": 0.40
}
```

**Response 200:**
```json
{
  "narrative": "法國憑藉商業資本強勢注入與ELO優勢，預計以穩健節奏主導比賽。",
  "confidence": 0.87
}
```

**Fallback (API unavailable):**
```json
{ "narrative": "分析服務暫時不可用", "confidence": 0.0 }
```

---

## GET /api/signals

Returns match signals built from **The Odds API** h2h odds cross-referenced with model probabilities.

Requires `THE_ODDS_API_KEY` in `.env`. On each `POST /api/pipeline/sync`, the backend fetches `soccer_fifa_world_cup` odds and stores them in `historical_odds`. If the key is unset or no odds are matched, returns `[]`.

**Response 200:** array of `MatchSignal` objects:

```json
[
  {
    "id": "12",
    "homeTeam": "Mexico",
    "awayTeam": "South Africa",
    "homeFlag": "mx",
    "awayFlag": "za",
    "kickoff": "2026-06-11T13:00:00Z",
    "bookmarkOdds": { "home": 1.85, "draw": 3.40, "away": 4.50 },
    "impliedProb": { "home": 0.52, "draw": 0.28, "away": 0.20 },
    "pFinal": { "home": 0.48, "draw": 0.27, "away": 0.25 },
    "ev": { "home": -0.11, "draw": -0.08, "away": 0.12 },
    "weights": { "w1": 0.30, "w2": 0.30, "w3": 0.40, "clip": 0.05 },
    "kelly_fraction": { "home": 0.0, "draw": 0.0, "away": 0.03 }
  }
]
```

| Field | Type | Description |
|-------|------|-------------|
| `kelly_fraction` | object | Fractional Kelly per outcome: `max(0, (p*b-q)/b) * kelly_scale` where `kelly_scale` comes from `model_config` |

| Data | Source |
|------|--------|
| Odds | [The Odds API](https://the-odds-api.com/) `soccer_fifa_world_cup` |
| Model probs | ELO + Poisson engine (same as `/api/predict`) |
| Match linkage | Team names + kickoff time window |

**Note:** worldcup2026 CDN does **not** include betting odds — only scores/fixtures.

---

## GET /api/pipeline/status

Returns data adapter health and **upcoming pre-match odds sync windows** (12h / 2h / 15m before each fixture kickoff).

**Response 200:**
```json
{
  "scrapers": [
    { "name": "World Bank (GDP)", "status": "ok", "last_fetch": "2026-06-19T01:00:00Z" },
    { "name": "Wikimedia (Squad Meta)", "status": "ok", "last_fetch": "2026-06-19T00:00:00Z" },
    { "name": "Kaggle Hist. Results", "status": "ok", "last_fetch": "2026-06-18T01:00:00Z" },
    { "name": "Kaggle/FIFA Players", "status": "ok", "last_fetch": "2026-06-18T01:00:00Z" },
    { "name": "FBref (xG)", "status": "ok", "last_fetch": "2026-06-19T01:00:00Z" },
    { "name": "FMScout (Player Attrs)", "status": "ok", "last_fetch": "2026-06-19T01:00:00Z" },
    { "name": "WorldCup2026 (Fixtures)", "status": "ok", "last_fetch": "2026-06-19T01:00:00Z" },
    { "name": "The Odds API", "status": "ok", "last_fetch": "2026-06-19T01:00:00Z" }
  ],
  "schedules": [
    {
      "label": "賽前 12 小時",
      "next_at": "2026-06-19T13:00:00Z",
      "endpoint": "The Odds API",
      "match_label": "USA vs Australia",
      "window_key": "12h",
      "wc_match_id": "M001",
      "synced": false
    }
  ],
  "updated_at": "2026-06-19T01:00:00Z"
}
```

Schedules are computed from `matches.kickoff` in PostgreSQL. A background scheduler runs every minute and calls The Odds API when a window is due (±10m grace). Completed windows are recorded in `odds_sync_windows` so each match/window syncs once.

---

## GET /api/pipeline/targets

Returns scraper name → endpoint label map for the Pipeline UI.

---

## POST /api/pipeline/sync

Full data pipeline sync (World Bank, Wikimedia, Kaggle, SoFIFA, worldcup2026 fixtures, The Odds API). **Does not** scrape FBref xG — use `POST /api/pipeline/sync/fbref` for that.

---

## POST /api/pipeline/sync/fbref

FBref xG-only sync. Persists `avg_xg_for` to PostgreSQL; on per-team failure **keeps stale DB/cache values**. Expect ~6–12 minutes for 48 teams (8–15s delay between requests).

**Response 200:**
```json
{
  "status": "ok",
  "teams_ok": 32,
  "teams_total": 48,
  "teams_fail": 16,
  "message": "32/48 teams updated (16 failed, stale cache kept)",
  "error": ""
}
```

`status` is `degraded` when zero teams updated but slugs exist (Cloudflare 403 etc.).

Manual seed fallback: `data/fbref_xg.csv` or `FBREF_XG_CSV` env (see Poisson xG fill chain in docs).

---

## POST /api/pipeline/sync/odds

Odds-only sync via The Odds API. Updates `historical_odds` without re-running other scrapers. Use for manual refresh from the Pipeline UI (**立即更新賠率**). Requires `THE_ODDS_API_KEY`.

**Response 200:**
```json
{ "status": "ok" }
```

On failure:
```json
{ "status": "error", "message": "THE_ODDS_API_KEY not set" }
```

---

## POST /api/pipeline/calibrate

Runs grid search calibration over all finished WC matches (Brier score, 5% weight steps). Saves best weights to `model_config`. Also runs automatically after each full `SyncAll`.

**Response 200:**
```json
{ "w1": 0.30, "w2": 0.35, "w3": 0.35, "brier_score": 0.481, "matches_used": 12 }
```

**Error 400:** fewer than 3 finished matches available (`insufficient data`).

**Error 500:** database query failure.
