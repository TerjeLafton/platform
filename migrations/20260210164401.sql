-- Add new schema named "log"
CREATE SCHEMA "log";
-- Create "entries" table
CREATE TABLE "log"."entries" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "timestamp" timestamptz NOT NULL,
  "level" text NOT NULL,
  "service" text NOT NULL,
  "module" text NOT NULL DEFAULT '',
  "correlation_id" text NOT NULL DEFAULT '',
  "message" text NOT NULL,
  "attrs" jsonb NOT NULL DEFAULT '{}',
  PRIMARY KEY ("id"),
  CONSTRAINT "entries_level_check" CHECK (level = ANY (ARRAY['INFO'::text, 'WARN'::text, 'ERROR'::text])),
  CONSTRAINT "entries_service_check" CHECK ((length(service) >= 1) AND (length(service) <= 50))
);
-- Create index "idx_log_entries_correlation_id" to table: "entries"
CREATE INDEX "idx_log_entries_correlation_id" ON "log"."entries" ("correlation_id") WHERE (correlation_id <> ''::text);
-- Create index "idx_log_entries_level" to table: "entries"
CREATE INDEX "idx_log_entries_level" ON "log"."entries" ("level");
-- Create index "idx_log_entries_service" to table: "entries"
CREATE INDEX "idx_log_entries_service" ON "log"."entries" ("service");
-- Create index "idx_log_entries_timestamp" to table: "entries"
CREATE INDEX "idx_log_entries_timestamp" ON "log"."entries" ("timestamp" DESC);
