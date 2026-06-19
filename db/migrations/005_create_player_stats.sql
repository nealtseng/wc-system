CREATE TABLE IF NOT EXISTS player_stats (
    id          BIGSERIAL    PRIMARY KEY,
    team_id     VARCHAR(3)   REFERENCES teams(id),
    name        VARCHAR(200) NOT NULL,
    age         SMALLINT,
    position    VARCHAR(50),
    nationality VARCHAR(100),
    overall     FLOAT,
    source      VARCHAR(20)  NOT NULL DEFAULT 'fmcsv',
    imported_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
