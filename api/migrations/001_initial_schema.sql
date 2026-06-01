-- =============================================
-- StackTrace: Schema completo para Aiven (PostgreSQL puro)
-- Migração de Neon (TimescaleDB) → Aiven
-- =============================================
-- INSTRUÇÕES:
--   1. Crie o serviço PostgreSQL Free no Aiven
--   2. Conecte via psql ou pelo SQL console do Aiven
--   3. Execute este script inteiro
-- =============================================

-- 1. Tabela de projetos
CREATE TABLE IF NOT EXISTS projects (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name           TEXT NOT NULL,
  slug           TEXT UNIQUE,
  api_key        TEXT NOT NULL UNIQUE DEFAULT gen_random_uuid()::text,
  owner_email    TEXT NOT NULL,
  user_id        UUID,
  api_key_hash   TEXT,
  api_key_prefix TEXT,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Tabela de usuários (auth)
CREATE TABLE IF NOT EXISTS users (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email      TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- FK: projects.user_id → users.id
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'fk_projects_user_id'
  ) THEN
    ALTER TABLE projects
      ADD CONSTRAINT fk_projects_user_id
      FOREIGN KEY (user_id) REFERENCES users(id);
  END IF;
END $$;

-- 3. Tabela de logs (PostgreSQL puro — sem hypertable)
-- Nota: No Neon, esta tabela usava composite PK (id, timestamp) para hypertable.
--       No Aiven, usamos BIGSERIAL simples como PK.
CREATE TABLE IF NOT EXISTS logs (
  id         BIGSERIAL PRIMARY KEY,
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  timestamp  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  level      TEXT NOT NULL CHECK (level IN ('info', 'warn', 'error')),
  message    TEXT NOT NULL,
  service    TEXT NOT NULL DEFAULT 'default',
  trace_id   TEXT,
  metadata   JSONB
);

-- Índices para performance das queries do dashboard
CREATE INDEX IF NOT EXISTS idx_logs_project_level   ON logs (project_id, level, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_logs_project_service ON logs (project_id, service, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_logs_trace_id        ON logs (trace_id) WHERE trace_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_logs_project_ts      ON logs (project_id, timestamp DESC);

-- 4. Tabela de incidentes
CREATE TABLE IF NOT EXISTS incidents (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  title       TEXT NOT NULL,
  description TEXT,
  status      TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'resolved')),
  started_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  resolved_at TIMESTAMPTZ
);

-- 5. Tabela de regras de alerta
CREATE TABLE IF NOT EXISTS alert_rules (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id     UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  condition      TEXT NOT NULL,
  threshold      INTEGER NOT NULL,
  window_seconds INTEGER NOT NULL DEFAULT 60,
  channel        TEXT NOT NULL CHECK (channel IN ('email', 'whatsapp')),
  destination    TEXT NOT NULL,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
