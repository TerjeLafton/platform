CREATE SCHEMA IF NOT EXISTS id;

CREATE TABLE id.users (
  id UUID PRIMARY KEY DEFAULT uuidv7(),
  email TEXT NOT NULL UNIQUE CHECK (LENGTH(email) <= 255),
  password_hash TEXT NOT NULL,
  name TEXT NOT NULL CHECK (LENGTH(name) >= 1 AND LENGTH(name) <= 100),
  avatar BYTEA,
  avatar_content_type TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
