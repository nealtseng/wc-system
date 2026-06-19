CREATE TABLE IF NOT EXISTS matches (
    id         BIGSERIAL    PRIMARY KEY,
    home_id    VARCHAR(3)   REFERENCES teams(id),
    away_id    VARCHAR(3)   REFERENCES teams(id),
    kickoff    TIMESTAMPTZ  NOT NULL,
    stadium    VARCHAR(200),
    stage      VARCHAR(50),
    home_score SMALLINT,
    away_score SMALLINT,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
