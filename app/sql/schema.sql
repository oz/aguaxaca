-- deliveries stores information about the water delivery and their
-- schedules in time.
CREATE TABLE IF NOT EXISTS deliveries (
  id            INTEGER PRIMARY KEY,
  date          TIMESTAMP NOT NULL,
  schedule      TEXT NOT NULL,
  location_type TEXT NOT NULL,
  location_name TEXT NOT NULL,
  created_at    TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_deliveries_date ON deliveries(date);

-- imports is the "queue" for images with delivery data.
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

-- FTS on delivery locations
CREATE VIRTUAL TABLE IF NOT EXISTS deliveries_fts USING fts5(id UNINDEXED, location_name);

-- FTS updates
CREATE TRIGGER IF NOT EXISTS deliveries_ai AFTER INSERT ON deliveries BEGIN
  INSERT INTO deliveries_fts (id, location_name) VALUES (new.id, new.location_name);
END;

CREATE TRIGGER IF NOT EXISTS deliveries_ad AFTER DELETE ON deliveries BEGIN
  DELETE FROM deliveries_fts WHERE id = old.id;
END;

CREATE TRIGGER IF NOT EXISTS deliveries_au AFTER UPDATE ON deliveries BEGIN
  UPDATE deliveries_fts SET location_name = new.location_name WHERE id = old.id;
END;
