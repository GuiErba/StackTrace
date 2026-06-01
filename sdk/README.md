# StackTrace SDK

[![npm version](https://img.shields.io/npm/v/@stacktrace-dev/sdk.svg)](https://www.npmjs.com/package/@stacktrace-dev/sdk)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Node.js](https://img.shields.io/badge/node-%3E%3D18.0.0-brightgreen.svg)](https://nodejs.org)

The official **StackTrace** SDK for Node.js and TypeScript. Capture structured logs, monitor HTTP requests, and automatically track unhandled exceptions — all sent to your [StackTrace Dashboard](https://devstacktrace.com/) in real time.

- 🔋 **Zero dependencies** — lightweight and fast
- 📦 **Batched delivery** — logs are queued and sent in batches for zero performance impact
- 🔄 **Automatic retries** — failed deliveries are retried with exponential backoff
- 🛡️ **Exception capture** — `uncaughtException` and `unhandledRejection` are caught automatically
- 🔌 **Framework plugins** — native middleware for Express.js and Fastify
- 🧩 **TypeScript-first** — full type definitions included out of the box
- 📡 **Dual format** — supports both CommonJS (`require`) and ESModules (`import`)

---

## Prerequisites

- **Node.js** >= 18.0.0
- A **StackTrace account** — [create one for free](https://devstacktrace.com/)

---

## Quick Start

Get up and running in under 5 minutes:

### 1. Create your account

Go to [devstacktrace.com](https://devstacktrace.com/) and sign up with your email.

### 2. Register your project

After logging in, create a new project in the dashboard. Give it a name that identifies your application or microservice (e.g., `payments-api`, `auth-service`).

### 3. Copy your API Key

Once your project is created, the dashboard will generate a unique **API Key** (starts with `sk_live_`). Copy it — you'll need it in the next step.

### 4. Install the SDK

```bash
npm install @stacktrace-dev/sdk
```

### 5. Initialize and send your first log

```typescript
import { StackTrace } from "@stacktrace-dev/sdk";

const logger = new StackTrace({
  apiKey: "sk_live_your_api_key_here",  // Paste your API Key from the dashboard
  service: "my-api",                    // Your application/microservice name
  baseUrl: "https://stacktrace-api.onrender.com",
});

logger.info("Hello from StackTrace! 🚀");
```

### 6. Check your dashboard

Go back to [devstacktrace.com](https://devstacktrace.com/) — your log should appear in real time. That's it! ✅

---

## Installation

```bash
npm install @stacktrace-dev/sdk
```

```bash
# or with yarn
yarn add @stacktrace-dev/sdk
```

```bash
# or with pnpm
pnpm add @stacktrace-dev/sdk
```

---

## Configuration

Create an instance of `StackTrace` with your configuration:

```typescript
import { StackTrace } from "@stacktrace-dev/sdk";

const logger = new StackTrace({
  apiKey: "sk_live_your_api_key_here",
  service: "payments-api",
  baseUrl: "https://stacktrace-api.onrender.com",
  environment: "production",
  debug: true,
});
```

### All Options

| Option | Type | Required | Default | Description |
|--------|------|:--------:|---------|-------------|
| `apiKey` | `string` | ✅ | — | Your project API Key from the [StackTrace Dashboard](https://devstacktrace.com/) |
| `service` | `string` | ✅ | — | Name of your application or microservice (e.g., `"payments-api"`) |
| `baseUrl` | `string` | — | `http://localhost:8080` | StackTrace backend URL. Use `https://stacktrace-api.onrender.com` for cloud |
| `environment` | `string` | — | `"production"` | Environment label attached to all logs (e.g., `"staging"`, `"development"`) |
| `debug` | `boolean` | — | `false` | When `true`, prints SDK internal messages to the console (useful for troubleshooting) |
| `batchSize` | `number` | — | `50` | Maximum number of logs to send per batch |
| `flushIntervalMs` | `number` | — | `2000` | How often (in ms) the SDK flushes the log queue to the server |
| `maxRetries` | `number` | — | `3` | Number of retry attempts for failed batch deliveries (uses exponential backoff) |
| `captureExceptions` | `boolean` | — | `true` | Automatically capture `uncaughtException` and `unhandledRejection` events |

> **💡 Tip:** For most use cases, you only need `apiKey`, `service`, and `baseUrl`. The defaults are sensible for production workloads.

---

## Sending Logs

The SDK provides three log levels: `info`, `warn`, and `error`. Logs are kept in an in-memory queue and delivered in **batches** (every 2 seconds by default), ensuring zero bottleneck in your application.

### Basic Logs

```typescript
logger.info("User logged in successfully");
logger.warn("Slow database query detected");
logger.error("Failed to process payment");
```

### Logs with Metadata

Attach any structured data as metadata — it will be indexed and searchable in your dashboard:

```typescript
logger.info("Order created", {
  metadata: {
    orderId: "ORD-12345",
    amount: 99.90,
    currency: "USD",
  },
});

logger.warn("Login attempt with incorrect password", {
  metadata: {
    userId: "12345",
    ip: "192.168.1.1",
    attempts: 3,
  },
});
```

### Logs with Trace ID

For distributed tracing across microservices, attach a `traceId` to correlate logs from the same request:

```typescript
logger.error("Failed to connect to the main database", {
  traceId: "req-998877",
  metadata: {
    latencyMs: 1500,
    dbHost: "aws-rds-prod-01",
  },
});
```

---

## HTTP Request Monitoring (Middlewares)

The SDK includes native plugins for automatic HTTP request logging. Every incoming request is logged with:

- **Method** and **Route**
- **Status Code**
- **Latency** (duration in ms)

The log level is automatically classified based on the status code:
- `2xx` → `info`
- `4xx` → `warn`
- `5xx` → `error`

### Express.js

```typescript
import express from "express";
import { StackTrace, stacktraceMiddleware } from "@stacktrace-dev/sdk";

const app = express();
const logger = new StackTrace({
  apiKey: "sk_live_...",
  service: "my-backend",
  baseUrl: "https://stacktrace-api.onrender.com",
});

// Register the middleware BEFORE your routes
app.use(stacktraceMiddleware(logger));

app.get("/", (req, res) => {
  res.send("Hello, world!");
});

app.listen(3000);
```

### Fastify

```typescript
import fastify from "fastify";
import { StackTrace, stacktraceFastifyPlugin } from "@stacktrace-dev/sdk";

const app = fastify();
const logger = new StackTrace({
  apiKey: "sk_live_...",
  service: "my-backend",
  baseUrl: "https://stacktrace-api.onrender.com",
});

// Register the StackTrace plugin
app.register(stacktraceFastifyPlugin(logger));

app.get("/", async (request, reply) => {
  return { hello: "world" };
});

app.listen({ port: 3000 });
```

---

## Automatic Exception Capture

By default, the SDK automatically captures:

- **`uncaughtException`** — synchronous errors that crash the process
- **`unhandledRejection`** — unhandled promise rejections

These are logged as `error` level with the full stack trace attached. No configuration needed — it works as soon as you create the `StackTrace` instance.

To disable this behavior:

```typescript
const logger = new StackTrace({
  apiKey: "sk_live_...",
  service: "my-backend",
  baseUrl: "https://stacktrace-api.onrender.com",
  captureExceptions: false, // Disable automatic exception capture
});
```

---

## Shutdown / Graceful Exit

When your application shuts down (deploy, restart, scaling, etc.), call `shutdown()` to force an immediate flush of any remaining logs in the queue:

```typescript
process.on("SIGTERM", async () => {
  await logger.shutdown(); // Flushes the log buffer before exiting
  process.exit(0);
});
```

> **Note:** The SDK already listens to the Node.js `beforeExit` event automatically, but you should call `shutdown()` manually in `SIGTERM`/`SIGINT` handlers for reliable behavior in containerized environments (Docker, Kubernetes, etc.).

---

## Troubleshooting

### Logs are not appearing in the dashboard

1. **Check your API Key** — make sure you copied the full key from the dashboard (starts with `sk_live_`).
2. **Check the `baseUrl`** — for cloud usage, it must be `https://stacktrace-api.onrender.com`.
3. **Enable debug mode** — set `debug: true` in your config to see SDK internal messages in the console.
4. **Check network connectivity** — ensure your server can reach the StackTrace backend.

### I see `[stacktrace-sdk] [debug] Flushed N logs` but nothing in the dashboard

This means the SDK is sending logs successfully. Check if you're looking at the correct project/service in the dashboard.

### Can I use the SDK in development?

Yes! Set `environment: "development"` to tag your logs. You can filter by environment in the dashboard.

---

## Compatibility

| Feature | Supported |
|---------|:---------:|
| CommonJS (`require`) | ✅ |
| ES Modules (`import`) | ✅ |
| TypeScript (types included) | ✅ |
| Node.js >= 18 | ✅ |
| External dependencies | **None** (zero-dependency) |

---

## License

[MIT](https://opensource.org/licenses/MIT)
