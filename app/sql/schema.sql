CREATE TABLE IF NOT EXISTS deliveries (
  id            INTEGER PRIMARY KEY,
  date          TIMESTAMP NOT NULL,
  schedule      TEXT NOT NULL,
  location_type TEXT NOT NULL,
  location_name TEXT NOT NULL,
  created_at    TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_deliveries_date ON deliveries(date);

CREATE INDEX IF NOT EXISTS idx_deliveries_location ON deliveries(location_name);

CREATE TABLE IF NOT EXISTS imports (
  id           INTEGER PRIMARY KEY,
  file_path    TEXT NOT NULL,
  file_hash    INTEGER UNIQUE NOT NULL,
  created_at   TIMESTAMP NOT NULL,
  completed_at TIMESTAMP DEFAULT NULL,
  failed_at    TIMESTAMP DEFAULT NULL,
  runs         INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_imports_completed_at ON imports(completed_at);
