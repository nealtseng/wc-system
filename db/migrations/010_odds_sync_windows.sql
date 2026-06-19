-- Tracks per-match odds sync windows (12h / 2h / 15m before kickoff).
CREATE TABLE IF NOT EXISTS odds_sync_windows (
    match_id   BIGINT NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    window_key TEXT NOT NULL,
    synced_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (match_id, window_key)
);

CREATE INDEX IF NOT EXISTS idx_odds_sync_windows_synced_at ON odds_sync_windows (synced_at);
