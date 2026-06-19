CREATE TABLE IF NOT EXISTS teams (
    id             VARCHAR(3)   PRIMARY KEY,
    name           VARCHAR(100) NOT NULL,
    elo            FLOAT        NOT NULL DEFAULT 1500,
    gdp_per_capita FLOAT,
    wiki_extract   TEXT,
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
