# StackTrace

Open-source observability platform for Micro SaaS founders. Get logs, incident detection, and email alerts without the complexity of Datadog or New Relic.

**"Is my app working? What broke?"** — StackTrace answers these questions in plain language, with zero infrastructure knowledge required.

---

## Features

- **Log Ingestion** — Receive logs from any backend via SDK or HTTP
- **Automatic Incident Detection** — Sliding window algorithm detects error spikes in real-time
- **Email Alerts** — Get notified when something breaks (via Resend)
- **Public Status Page** — Your users can check if your service is up at `/status/:slug`
- **Dashboard** — Beautiful React UI with real-time metrics, log viewer, and incident management

## Architecture

```
[Your App] → SDK (batch + retry) → [Go API]
                                       │
                                       ├── Redis (auth cache + rate limit)
                                       ├── PostgreSQL/TimescaleDB (logs + incidents)
                                       └── Alert Worker (in-memory sliding window)
                                              └── Email notification (Resend)
```

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Backend | Go + Gin |
| Database | PostgreSQL + TimescaleDB (Neon) |
| Cache | Redis (Upstash) |
| Dashboard | React + Vite + Tailwind CSS v4 |
| SDK | TypeScript (npm: `stacktrace-sdk`) |
| Email | Resend |

---

## Quick Start (Self-Hosting)

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- PostgreSQL database (we recommend [Neon](https://neon.tech) free tier)
- Redis instance (we recommend [Upstash](https://upstash.com) free tier)
- [Resend](https://resend.com) account for email alerts

### 1. Clone the repository

```bash
git clone https://github.com/GuiErba/StackTrace.git
cd StackTrace
```

### 2. Configure environment variables

```bash
cp api/.env.example api/.env
```

Edit `api/.env` with your credentials:

```bash
# PostgreSQL (Neon)
DATABASE_URL=postgresql://user:pass@host/db?sslmode=require

# Redis (Upstash)
UPSTASH_REDIS_URL=rediss://default:token@host:6379

# Email alerts (Resend)
RESEND_API_KEY=re_xxxxxxxxxxxx
RESEND_FROM=alerts@yourdomain.com

# Auth
JWT_SECRET=your-secret-key

# CORS (your dashboard URL)
CORS_ORIGINS=http://localhost:5173
```

### 3. Run the backend

```bash
# With Docker
docker build -t stacktrace-api ./api
docker run -p 8080:8080 --env-file api/.env stacktrace-api

# Or with Go directly
cd api && go run .
```

### 4. Run the dashboard

```bash
cd dashboard
npm install
npm run dev
```

Open [http://localhost:5173](http://localhost:5173) in your browser.

### 5. Install the SDK in your app

```bash
npm install stacktrace-sdk
```

```typescript
import { StackTrace } from 'stacktrace-sdk'

const tracker = new StackTrace({
  apiKey: 'your-api-key',
  service: 'my-app',
})

// Manual logging
tracker.info('User signed up', { userId: 123 })
tracker.error('Payment failed', { orderId: 456 })

// Express middleware (auto-captures requests)
app.use(tracker.expressMiddleware())
```

---

## Deployment (Free Tier)

StackTrace is designed to run entirely on free-tier services:

| Service | Purpose | Free Tier |
|---------|---------|-----------|
| [Render](https://render.com) | Backend hosting | 750h/month |
| [Cloudflare Pages](https://pages.cloudflare.com) | Dashboard hosting | Unlimited requests |
| [Neon](https://neon.tech) | PostgreSQL + TimescaleDB | 0.5 GB |
| [Upstash](https://upstash.com) | Redis | 500k commands/month |
| [Resend](https://resend.com) | Email alerts | 3,000 emails/month |
| [BetterStack](https://betterstack.com) | Uptime monitoring | 10 monitors |

### Deploy to Render

1. Fork this repo
2. Go to [Render Dashboard](https://dashboard.render.com) → **New** → **Blueprint**
3. Connect your GitHub repo — Render will read the `render.yaml` file
4. Set the environment variables marked as `sync: false` in the Render dashboard
5. Deploy

### Deploy Dashboard to Cloudflare Pages

1. Go to [Cloudflare Dashboard](https://dash.cloudflare.com) → **Workers & Pages** → **Create**
2. Connect your GitHub repo
3. Set build settings:
   - **Build command:** `cd dashboard && npm ci && npm run build`
   - **Build output directory:** `dashboard/dist`
4. Add environment variable: `VITE_API_URL=https://stacktrace-api.onrender.com`
5. Deploy

### Keep Render Awake

Render's free tier sleeps after 15 minutes of inactivity. Use [BetterStack](https://betterstack.com) to ping your `/health` endpoint every 5 minutes:

- **URL:** `https://stacktrace-api.onrender.com/health`
- **Check interval:** 5 minutes

---

## API Endpoints

| Method | Route | Auth | Description |
|--------|-------|------|-------------|
| `POST` | `/logs` | API Key | Ingest a single log |
| `POST` | `/logs/batch` | API Key | Ingest multiple logs |
| `GET` | `/dashboard/logs` | JWT | Query logs with filters |
| `GET` | `/dashboard/incidents` | JWT | List incidents |
| `PATCH` | `/dashboard/incidents/:id/resolve` | JWT | Resolve an incident |
| `GET` | `/dashboard/alert-rules` | JWT | List alert rules |
| `POST` | `/dashboard/alert-rules` | JWT | Create alert rule |
| `DELETE` | `/dashboard/alert-rules/:id` | JWT | Delete alert rule |
| `GET` | `/dashboard/metrics/overview` | JWT | Get overview metrics |
| `GET` | `/status/:slug` | None | Public status page |
| `GET` | `/health` | None | Health check |

---

## Configuring Alerts

1. Log in to the dashboard
2. Go to **Alert Rules**
3. Click **Create Rule**
4. Set:
   - **Condition:** Error count
   - **Threshold:** Number of errors to trigger (e.g., 5)
   - **Window:** Time window in seconds (e.g., 60)
   - **Channel:** Email
   - **Destination:** Your email address
5. When errors exceed the threshold within the window, you'll receive an email and an incident will be created automatically

---

## License

MIT
