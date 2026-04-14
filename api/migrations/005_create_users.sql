-- Phase 5: Auth system migration
-- Run this on Neon SQL console

-- Users table
CREATE TABLE IF NOT EXISTS users (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email      TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add user_id and hashed API key columns to projects
ALTER TABLE projects ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id);
ALTER TABLE projects ADD COLUMN IF NOT EXISTS api_key_hash TEXT;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS api_key_prefix TEXT;
