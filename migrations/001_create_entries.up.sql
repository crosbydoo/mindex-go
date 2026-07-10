CREATE TABLE IF NOT EXISTS entries (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    abstract TEXT NOT NULL,
    category TEXT NOT NULL,
    year INTEGER NOT NULL,
    author TEXT NOT NULL,
    source TEXT NOT NULL,
    type TEXT NOT NULL,
    url TEXT NOT NULL DEFAULT '#',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_entries_year_id ON entries (year DESC, id DESC);
