CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO categories (name) VALUES
    ('Clinical Psychology'),
    ('Developmental Psychology'),
    ('Cognitive Psychology'),
    ('Social Psychology'),
    ('Educational Psychology'),
    ('Mental Health'),
    ('Research Methods')
ON CONFLICT (name) DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_categories_name ON categories (name);
