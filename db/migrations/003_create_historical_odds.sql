CREATE TABLE IF NOT EXISTS historical_odds (
    id          BIGSERIAL    PRIMARY KEY,
    match_id    BIGINT       REFERENCES matches(id),
    source      VARCHAR(50)  NOT NULL,
    recorded_at TIMESTAMPTZ  NOT NULL,
    home_odds   FLOAT,
    draw_odds   FLOAT,
    away_odds   FLOAT
);
