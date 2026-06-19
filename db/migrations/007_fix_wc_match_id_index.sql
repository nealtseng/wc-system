-- Partial unique indexes require ON CONFLICT ... WHERE; use a full unique index instead.
DROP INDEX IF EXISTS matches_wc_id_idx;
CREATE UNIQUE INDEX IF NOT EXISTS matches_wc_id_idx ON matches (wc_match_id);
