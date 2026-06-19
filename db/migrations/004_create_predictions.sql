CREATE TABLE IF NOT EXISTS predictions (
    id         BIGSERIAL    PRIMARY KEY,
    match_id   BIGINT       REFERENCES matches(id),
    w1         FLOAT        NOT NULL,
    w2         FLOAT        NOT NULL,
    w3         FLOAT        NOT NULL,
    clip_val   FLOAT        NOT NULL,
    p_home     FLOAT        NOT NULL,
    p_draw     FLOAT        NOT NULL,
    p_away     FLOAT        NOT NULL,
    narrative  TEXT,
    confidence FLOAT,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
