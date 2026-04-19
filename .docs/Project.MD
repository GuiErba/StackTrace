# StackTrace — Contexto completo do projeto

> Este arquivo é a fonte de verdade para a IA de codagem entender a estrutura, decisões e estado do projeto.

---

## O que é o StackTrace

StackTrace é uma plataforma de observabilidade open source, self-hosted, voltada para founders de Micro SaaS e desenvolvedores solo que não têm conhecimento técnico profundo sobre infraestrutura. O objetivo é ser o "Datadog para humanos" — uma alternativa simples, acessível e sem jargão ao Datadog, New Relic e similares.

O usuário-alvo é alguém que quer saber "meu app está funcionando?" e "o que quebrou agora?" sem precisar entender o que é um stack trace ou um percentil de latência.

Quando o projeto do usuário escalar e precisar de features avançadas, ele migra para o Datadog. O StackTrace é o ponto de entrada.

---

## Objetivo do projeto

- Receber logs de qualquer backend (Node.js, Python, Go) via SDK ou HTTP direto
- Exibir logs em linguagem humana, não técnica
- Detectar incidentes automaticamente (muitos erros em pouco tempo)
- Alertar o dono do projeto por email e/ou WhatsApp quando algo quebra
- Expor uma página de status pública para os clientes do usuário verem se o serviço está no ar
- Ser 100% gratuito para hospedar no free tier dos serviços listados abaixo

---

## Modelo de negócio

Open source, self-hosted. O usuário baixa e hospeda na própria infra. Não há plano pago nem SaaS no momento.

---

## Repositório (monorepo)

```
stacktrace/
├── api/          # Backend em Go (Gin) — API de ingestão e query
├── dashboard/    # Frontend em React + Vite — interface do usuário
├── sdk/          # SDK Node.js (TypeScript) — pacote npm que os usuários instalam
└── docs/         # Documentação
```

---

## Stack técnica

### Backend (`api/`)
- **Linguagem:** Go 1.22+
- **Framework HTTP:** Gin (`github.com/gin-gonic/gin`)
- **Banco de dados:** PostgreSQL com extensão TimescaleDB (via Neon no free tier)
- **Driver Go para Postgres:** `database/sql` + `github.com/lib/pq`
- **Cache e rate limit:** Redis via Upstash (`github.com/redis/go-redis/v9`)
- **Migrations:** `golang-migrate`
- **Email:** Resend (SDK Go ou HTTP direto)
- **Env vars:** `github.com/joho/godotenv`

### Frontend (`dashboard/`)
- **Framework:** React 18 + Vite + TypeScript
- **Estilo:** Tailwind CSS
- **Roteamento:** React Router v6
- **Data fetching:** React Query (TanStack Query)
- **Gráficos:** Recharts
- **Build:** Vite

### SDK (`sdk/`)
- **Linguagem:** TypeScript
- **Build:** tsup
- **Publicação:** npm (`stacktrace-sdk`)
- **Compatibilidade:** Node.js 18+, suporte a Express e Fastify via middlewares prontos

---

## Serviços externos (todos com plano gratuito)

| Serviço | Uso | Limite free |
|---|---|---|
| Neon | PostgreSQL + TimescaleDB | 0,5 GB · 100 CU-h/mês |
| Upstash | Redis (rate limit + cache) | 500k comandos/mês |
| Render | Hospedagem do backend Go | 750h/mês compute |
| Cloudflare Pages | Hospedagem do frontend | Requests ilimitados |
| Resend | Alertas por email | 3.000 emails/mês |
| BetterStack | Uptime monitoring + keep-alive | 10 monitores |
| GitHub Actions | CI/CD | 2.000 min/mês |

> **Atenção Render:** o free tier dorme após 15 min sem tráfego. O BetterStack pinga `/health` a cada 5 min para manter o servidor acordado.

---

## Arquitetura do sistema

```
[App do usuário]
      │
      ▼ (HTTP POST /logs via SDK)
[API Go — Gin]
      │
      ├── Middleware auth (API key → project_id via Redis cache)
      ├── Rate limit (Redis INCR, 10k logs/min por projeto)
      │
      ▼
[Go channel — buffer 10.000]
      │
      ├──▶ [Worker de persistência — goroutine pool x5]
      │           └──▶ PostgreSQL + TimescaleDB (Neon)
      │
      └──▶ [Worker de alertas — sliding window 60s]
                  └──▶ Se threshold ultrapassado:
                            ├──▶ Resend (email)
                            └──▶ Tabela incidents no Postgres
                                 └──▶ GET /status (status page pública)
```

### Por que Go channel em vez de salvar direto no banco

O Go channel desacopla ingestão de persistência. Se o banco travar por 2 segundos, os requests do usuário não são afetados — os logs ficam no buffer. Sem o channel, 1.000 requests simultâneos = 1.000 writes simultâneos no banco.

### Por que TimescaleDB em vez de MongoDB

Logs são dados de série temporal. O TimescaleDB cria hypertables com particionamento automático por tempo, reduzindo o espaço em até 90% com compressão e tornando queries de janela temporal (últimas 2h, ontem) muito mais rápidas. É uma extension do Postgres — SQL normal funciona sem precisar aprender nada novo.

### Por que Redis com sliding window para alertas

Em vez de rodar `COUNT(*)` no banco a cada segundo, o worker mantém um mapa em memória `{project_id → []timestamp}` e desliza uma janela de 60 segundos. É O(1) e não toca o banco para decidir se dispara alerta.

---

## Schema do banco de dados

```sql
-- Cada cliente do StackTrace tem um projeto
CREATE TABLE projects (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT NOT NULL,
  api_key     TEXT NOT NULL UNIQUE DEFAULT gen_random_uuid()::text,
  owner_email TEXT NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tabela principal de logs (hypertable TimescaleDB, particionada por timestamp)
CREATE TABLE logs (
  id         BIGSERIAL,
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  timestamp  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  level      TEXT NOT NULL CHECK (level IN ('info', 'warn', 'error')),
  message    TEXT NOT NULL,
  service    TEXT NOT NULL DEFAULT 'default',
  trace_id   TEXT,
  metadata   JSONB,
  PRIMARY KEY (id, timestamp)
);
SELECT create_hypertable('logs', 'timestamp');

-- Índices para as queries mais comuns do dashboard
CREATE INDEX idx_logs_project_level   ON logs (project_id, level, timestamp DESC);
CREATE INDEX idx_logs_project_service ON logs (project_id, service, timestamp DESC);
CREATE INDEX idx_logs_trace_id        ON logs (trace_id) WHERE trace_id IS NOT NULL;

-- Incidentes detectados automaticamente pelo worker de alertas
CREATE TABLE incidents (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  title       TEXT NOT NULL,
  description TEXT,
  status      TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'resolved')),
  started_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  resolved_at TIMESTAMPTZ
);

-- Regras de alerta configuradas pelo usuário
CREATE TABLE alert_rules (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id     UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  condition      TEXT NOT NULL,   -- ex: 'error_count', 'latency_p95'
  threshold      INTEGER NOT NULL,
  window_seconds INTEGER NOT NULL DEFAULT 60,
  channel        TEXT NOT NULL CHECK (channel IN ('email', 'whatsapp')),
  destination    TEXT NOT NULL,   -- email ou número de telefone
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

## Estrutura de pastas do backend Go (`api/`)

```
api/
├── cmd/
│   └── main.go                  # Entrypoint — inicializa DB, Redis, router, workers
├── internal/
│   ├── handlers/
│   │   ├── logs.go              # POST /logs, GET /logs
│   │   ├── incidents.go         # GET /incidents
│   │   ├── alert_rules.go       # CRUD /alert-rules
│   │   └── status.go            # GET /status/:slug (público, sem auth)
│   ├── services/
│   │   ├── ingest.go            # Lógica de receber log e colocar no channel
│   │   ├── alert.go             # Worker de alertas (sliding window)
│   │   └── query.go             # Queries de logs com filtros
│   ├── repository/
│   │   ├── logs.go              # SQL queries na tabela logs
│   │   ├── projects.go          # SQL queries na tabela projects
│   │   └── incidents.go         # SQL queries na tabela incidents
│   ├── middleware/
│   │   ├── auth.go              # Valida API key, injeta project_id no contexto
│   │   └── ratelimit.go         # Rate limit por projeto via Redis
│   ├── cache/
│   │   └── redis.go             # Cliente Redis (Upstash)
│   └── models/
│       ├── log.go               # Struct LogEntry
│       ├── project.go           # Struct Project
│       ├── incident.go          # Struct Incident
│       └── alert_rule.go        # Struct AlertRule
├── pkg/
│   └── notify/
│       ├── email.go             # Integração Resend
│       └── whatsapp.go          # Integração Evolution API
├── migrations/
│   ├── 001_create_projects.sql
│   ├── 002_create_logs.sql
│   ├── 003_create_incidents.sql
│   └── 004_create_alert_rules.sql
├── .env                         # Variáveis de ambiente (não commitar)
├── .env.example                 # Template sem valores reais (commitar)
├── go.mod
└── go.sum
```

---

## Estrutura de pastas do frontend React (`dashboard/`)

```
dashboard/
├── src/
│   ├── pages/
│   │   ├── Overview.tsx         # Visão geral — métricas + gráfico + incidentes
│   │   ├── Logs.tsx             # Lista de logs com filtros e busca
│   │   ├── Incidents.tsx        # Lista de incidentes ativos e histórico
│   │   ├── AlertRules.tsx       # Configurar regras de alerta
│   │   ├── StatusPage.tsx       # Página pública (sem auth, rota /status/:slug)
│   │   └── Login.tsx            # Login por email + API key
│   ├── components/
│   │   ├── Sidebar.tsx          # Navegação lateral
│   │   ├── MetricCard.tsx       # Card de métrica (requests, erros, latência)
│   │   ├── LogTable.tsx         # Tabela de logs com badges de nível
│   │   ├── IncidentItem.tsx     # Item de incidente com ícone de severidade
│   │   └── RequestsChart.tsx    # Gráfico de requests por hora (Recharts)
│   ├── hooks/
│   │   ├── useLogs.ts           # React Query para buscar logs
│   │   └── useMetrics.ts        # React Query para buscar métricas
│   ├── lib/
│   │   └── api.ts               # Fetch wrapper com Authorization header automático
│   └── main.tsx
├── .env                         # VITE_API_URL=https://sua-api.onrender.com
└── vite.config.ts
```

---

## Estrutura do SDK Node.js (`sdk/`)

```
sdk/
├── src/
│   ├── client.ts                # Classe StackTrace — fila, batch, retry, flush
│   ├── middleware.ts            # Middlewares para Express e Fastify
│   └── index.ts                 # Exports públicos
├── package.json                 # name: "stacktrace-sdk"
└── tsconfig.json
```

### Como o SDK funciona

```typescript
import { StackTrace } from 'stacktrace-sdk'

const pilot = new StackTrace({
  apiKey: process.env.STACKTRACE_KEY,
  service: 'minha-loja',
  environment: process.env.NODE_ENV,
})

// Manual
pilot.error('Falha no checkout', { userId: 123 })

// Automático via middleware (Express)
app.use(stacktraceMiddleware(pilot))

// Captura uncaughtException e unhandledRejection automaticamente
```

O SDK acumula logs em uma fila interna e envia em batch (50 logs ou a cada 2s). Fire-and-forget com retry exponencial (3 tentativas: 1s, 2s, 4s). Nunca bloqueia o app do usuário.

---

## Endpoints da API Go

| Método | Rota | Auth | Descrição |
|---|---|---|---|
| POST | `/logs` | API key | Recebe um ou múltiplos logs |
| POST | `/logs/batch` | API key | Recebe array de logs (usado pelo SDK) |
| GET | `/logs` | API key | Lista logs com filtros |
| GET | `/incidents` | API key | Lista incidentes |
| POST | `/alert-rules` | API key | Cria regra de alerta |
| GET | `/alert-rules` | API key | Lista regras de alerta |
| DELETE | `/alert-rules/:id` | API key | Remove regra de alerta |
| GET | `/status/:slug` | Nenhuma | Página de status pública |
| GET | `/health` | Nenhuma | Health check (usado pelo BetterStack) |

### Formato do payload de log

```json
{
  "level": "error",
  "message": "Falha ao processar pagamento",
  "service": "checkout",
  "timestamp": "2026-03-24T14:22:08Z",
  "trace_id": "abc-123-def",
  "metadata": {
    "userId": 456,
    "endpoint": "/checkout",
    "statusCode": 500
  }
}
```

### Autenticação

Todos os endpoints (exceto `/status/:slug` e `/health`) requerem o header:

```
Authorization: Bearer {api_key}
```

O middleware busca o `project_id` no Redis com TTL de 5 minutos antes de consultar o Postgres. Retorna `401` se inválido.

---

## Variáveis de ambiente do backend

```bash
# api/.env

PORT=8080
ENV=development

# Neon PostgreSQL
DATABASE_URL=postgresql://usuario:senha@ep-xxx.us-east-2.aws.neon.tech/neondb?sslmode=require

# Upstash Redis (usar URL rediss:// do painel "Redis Client", não REST API)
UPSTASH_REDIS_URL=rediss://default:senha@xxx-xxx.upstash.io:6379

# Resend (alertas por email)
RESEND_API_KEY=re_xxxxxxxxxxxx
RESEND_FROM=alertas@stacktrace.dev

# Evolution API (alertas WhatsApp, opcional)
EVOLUTION_API_URL=https://evolution.seudominio.com
EVOLUTION_API_KEY=xxxxxxxxxxxx
```

## Variáveis de ambiente do frontend

```bash
# dashboard/.env

VITE_API_URL=https://sua-api.onrender.com
```

---

## Uso do Redis (chaves e TTLs)

```
stacktrace:apikey:{chave_hash}    → project_id (string)     TTL: 5 min
stacktrace:ratelimit:{project_id} → contador (integer)       TTL: 60s (janela deslizante)
stacktrace:status:{project_id}    → JSON (status page cache) TTL: 30s
```

---

## Roadmap de desenvolvimento (etapas)

### Fase 1 — Setup inicial (Concluído)
- [x] Criar monorepo no GitHub com pastas `api/`, `dashboard/`, `sdk/`
- [x] Instalar Go 1.22+ e configurar ambiente local
- [x] Criar projeto Go com `go mod init`
- [x] Criar conta no Neon, provisionar banco, ativar TimescaleDB, rodar migrations
- [x] Criar conta no Upstash, provisionar Redis
- [x] Criar `.env` com todas as variáveis

### Fase 2 — Backend Go: ingestão de logs (Concluído)
- [x] Instalar Gin e dependências base
- [x] Criar modelo `LogEntry` e handler `POST /logs`
- [x] Implementar Go channel + goroutine pool de persistência
- [x] Criar middleware de auth por API key (com cache Redis)
- [x] Criar endpoint `GET /logs` com filtros e paginação por cursor
- [x] Implementar rate limit com Upstash Redis
- [x] Criar endpoint `GET /health`
- [x] Testar com curl / Insomnia

### Fase 3 — Backend Go: worker de alertas (Concluído)
- [x] Implementar alert worker com sliding window em memória
- [x] Criar CRUD de `alert_rules`
- [x] Integrar Resend para alertas por email
- [x] Criar endpoint `GET /status/:slug` com cache de 30s

### Fase 4 — SDK Node.js (Concluído)
- [x] Criar pacote npm com TypeScript e tsup
- [x] Implementar classe `StackTrace` com fila + batch + retry
- [x] Criar middlewares para Express e Fastify
- [x] Publicar no npm como `stacktrace-sdk`

### Fase 5 — Dashboard React (Concluído)
- [x] Criar projeto Vite + React + TypeScript
- [x] Instalar Tailwind, React Query, React Router, Recharts
- [x] Tela de login e fluxo de auth
- [x] Layout base com sidebar
- [x] Página de visão geral (métricas + gráfico + incidentes)
- [x] Página de logs com filtros e busca
- [x] Página de configuração de alertas
- [x] Página de status público (`/status/:slug`, sem auth)

### Fase 6 — Deploy e CI/CD (Pendente)
- [ ] Criar Dockerfile multi-stage para o backend Go
- [ ] Deploy do backend no Render
- [ ] Deploy do dashboard no Cloudflare Pages
- [ ] Configurar GitHub Actions para testes e deploy automático
- [ ] Configurar BetterStack para uptime + ping anti-sleep no Render
- [ ] Teste end-to-end completo em produção
- [ ] Escrever README e documentação de self-hosting

---

## Estado atual do projeto

- [x] Todo o núcleo back-end (ingestão, auth com middleware, resiliência do db, background workers) finalizado.
- [x] Lógica complexa do Worker de alertas (sliding window em memória local) finalizada, auditada e documentada.
- [x] SDK Node.js robusto testado e simulando tráfego real.
- [x] Front-end (dashboard) construído com React, React Query, Recharts, totalmente refatorado utilizando Tailwind CSS v4, com tipos padronizados e organizados.
- [ ] A última grande etapa que resta é **Fase 6: Deploy e CI/CD** (Dockerização e pipeline).

---

## Decisões técnicas e seus motivos

| Decisão | Alternativa descartada | Motivo |
|---|---|---|
| Go no backend | Node.js | Melhor performance para ingestão de alto volume, goroutines nativas, baixo consumo de memória |
| TimescaleDB | MongoDB | Logs são série temporal — Timescale é 10x mais eficiente em queries por janela de tempo e comprime até 90% |
| Go channel + workers | Salvar direto no banco | Desacopla ingestão de persistência — pico de requests não trava o banco |
| Redis sliding window | COUNT(*) periódico no banco | O(1) em memória, não toca o banco para decisão de alerta |
| Upstash Redis | Redis self-hosted | Free tier generoso (500k cmd/mês), sem servidor pra gerenciar, HTTP API funciona no Render free |
| Neon Postgres | Supabase | TimescaleDB disponível como extension, scale-to-zero nativo, 100 CU-h free |
| Render | Fly.io | Fly.io removeu o free tier em 2024 |
| Cloudflare Pages | Vercel | Requests e bandwidth ilimitados no free tier |
| Resend | SendGrid | API mais moderna, SDK Go disponível, 3k emails/mês free |