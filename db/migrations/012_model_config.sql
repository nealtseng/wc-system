CREATE TABLE IF NOT EXISTS model_config (
    id         BIGSERIAL    PRIMARY KEY,
    key        VARCHAR(50)  UNIQUE NOT NULL,
    value      FLOAT        NOT NULL,
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
INSERT INTO model_config (key, value) VALUES
    ('w1', 0.30), ('w2', 0.30), ('w3', 0.40),
    ('clip_delta', 0.05), ('delta_max', 0.15), ('kelly_scale', 0.25)
ON CONFLICT (key) DO NOTHING;
