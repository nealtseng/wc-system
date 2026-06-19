# WC System — Local Development Setup

## 1. Obtain API Credentials

All credentials below are stored in the root `.env` file (copy from `.env.example`).

| Credential | Where to Get | Free Tier Limit | Required? |
|------------|-------------|-----------------|-----------|
| `DEEPSEEK_API_KEY` | https://platform.deepseek.com/api_keys | Pay-per-token | For `/api/narrative` only |
| `AIVEN_DB_URL` | https://console.aiven.io/ → PostgreSQL | Free plan available | Yes (or local Docker Postgres) |
| `THE_ODDS_API_KEY` | https://the-odds-api.com/#get-access | 500 requests/month | For odds + EV signals |
| `API_FOOTBALL_KEY` | https://dashboard.api-football.com/register | 100 requests/day | **Optional** — unused; live scores come from worldcup2026 CDN |

> **Odds auto-sync:** With `THE_ODDS_API_KEY` set, the backend scheduler pulls The Odds API at **12h / 2h / 15m** before each match kickoff. Manual refresh: `POST /api/pipeline/sync/odds` or **立即更新賠率** on the Pipeline page.

> **Note:** `DEEPSEEK_API_KEY` is required for the `/api/narrative` endpoint. If it is empty the endpoint returns a fallback message; the rest of the system functions normally.

## 2. Configure Environment Variables

```bash
# From the project root:
cp .env.example .env
# Edit .env with your actual values.
```

## 3. Local Development (Docker)

The `docker-compose.yml` runs the Go backend on port 8080 and a local PostgreSQL on port 5432.

```bash
# Build and start backend + local postgres:
docker compose up --build

# Stop:
docker compose down
```

The backend applies SQL migrations automatically on startup.

## 4. Frontend Development

```bash
cd frontend
npm install
npm run dev
# Vite dev server starts on http://localhost:3000
# /api/* requests are proxied to http://localhost:8080
```

## 5. Switching to Aiven (Production PostgreSQL)

Set `AIVEN_DB_URL` in `.env` to your Aiven connection string:
```
AIVEN_DB_URL=postgresql://avnadmin:YOURPASS@yourhost.aivencloud.com:PORT/defaultdb?sslmode=require
```

For production, run only the backend container (skip the local postgres service):
```bash
docker compose up backend
```

## 6. Legacy Notes

- `GEMINI_API_KEY` in `frontend/.env.example` is a leftover from the AI Studio scaffold. It is not used by the current system.
- `@google/genai` and `express` in `frontend/package.json` are AI Studio residue. They are safe to leave as unused dependencies.
- `Bankroll.tsx` is imported in `App.tsx` but not routed. It is reserved for a future Bankroll Management view.

## 7. Data Sync

On startup the backend syncs data for **48 teams** from four adapters plus the WorldCup2026 fixture source:

| Scraper | Source |
|---------|--------|
| World Bank (GDP) | World Bank Open Data API |
| Wikimedia (Squad Meta) | Wikipedia REST API |
| Kaggle Hist. Results | martj42 international results CSV |
| Kaggle/FIFA Players | SoFIFA player dataset |
| WorldCup2026 (Fixtures) | GitHub raw CDN (`rezarahiminia/worldcup2026`) |

**WorldCup2026 adapter** — On startup the backend automatically fetches teams, stadiums, and 104 fixtures from the public GitHub raw CDN and writes them to the database. No additional environment variables are required. The first startup may take ~25 seconds longer than before due to World Bank rate-limit sleep (~200 ms × 48 teams ≈ 10 s) plus the CDN fetch.
