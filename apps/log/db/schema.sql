CREATE SCHEMA IF NOT EXISTS log;

CREATE TABLE log.entries (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    timestamp TIMESTAMPTZ NOT NULL,
    level TEXT NOT NULL CHECK (level IN ('INFO', 'WARN', 'ERROR')),
    service TEXT NOT NULL CHECK (length(service) BETWEEN 1 AND 50),
    module TEXT NOT NULL DEFAULT '',
    correlation_id TEXT NOT NULL DEFAULT '',
    message TEXT NOT NULL,
    attrs JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_log_entries_timestamp ON log.entries (timestamp DESC);
CREATE INDEX idx_log_entries_service ON log.entries (service);
CREATE INDEX idx_log_entries_level ON log.entries (level);
CREATE INDEX idx_log_entries_correlation_id ON log.entries (correlation_id) WHERE correlation_id != '';
