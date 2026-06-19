ALTER TABLE historical_odds
    ADD COLUMN IF NOT EXISTS totals_line FLOAT,
    ADD COLUMN IF NOT EXISTS over_odds   FLOAT,
    ADD COLUMN IF NOT EXISTS under_odds  FLOAT;
