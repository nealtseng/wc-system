CREATE UNIQUE INDEX IF NOT EXISTS historical_odds_match_source_idx
    ON historical_odds (match_id, source);
