import type { StackTraceConfig, LogEntry, LogLevel, LogOptions } from "./types";

const DEFAULT_BASE_URL = "http://localhost:8080";
const DEFAULT_BATCH_SIZE = 50;
const DEFAULT_FLUSH_INTERVAL_MS = 2000;
const DEFAULT_MAX_RETRIES = 3;

export class StackTrace {
  private readonly apiKey: string;
  private readonly service: string;
  private readonly baseUrl: string;
  private readonly environment: string;
  private readonly batchSize: number;
  private readonly flushIntervalMs: number;
  private readonly maxRetries: number;
  private readonly debug: boolean;

  private queue: LogEntry[] = [];
  private timer: ReturnType<typeof setInterval> | null = null;
  private isFlushing = false;
  private isShutdown = false;

  constructor(config: StackTraceConfig) {
    this.apiKey = config.apiKey;
    this.service = config.service;
    this.baseUrl = (config.baseUrl ?? DEFAULT_BASE_URL).replace(/\/$/, "");
    this.environment = config.environment ?? "production";
    this.batchSize = config.batchSize ?? DEFAULT_BATCH_SIZE;
    this.flushIntervalMs = config.flushIntervalMs ?? DEFAULT_FLUSH_INTERVAL_MS;
    this.maxRetries = config.maxRetries ?? DEFAULT_MAX_RETRIES;
    this.debug = config.debug ?? false;

    this.startTimer();

    if (config.captureExceptions !== false) {
      this.setupExceptionHandlers();
    }

    this.log("debug", `StackTrace SDK initialized for service "${this.service}"`);
  }

  info(message: string, options?: LogOptions): void {
    this.enqueue("info", message, options);
  }

  warn(message: string, options?: LogOptions): void {
    this.enqueue("warn", message, options);
  }

  error(message: string, options?: LogOptions): void {
    this.enqueue("error", message, options);
  }

  async flush(): Promise<void> {
    if (this.queue.length === 0 || this.isFlushing) {
      return;
    }

    this.isFlushing = true;
    const batch = this.queue.splice(0, this.batchSize);

    try {
      await this.sendBatch(batch);
      this.log("debug", `Flushed ${batch.length} logs`);
    } catch {
      this.log("debug", `Failed to flush ${batch.length} logs after all retries, discarding`);
    } finally {
      this.isFlushing = false;
    }

    if (this.queue.length >= this.batchSize) {
      await this.flush();
    }
  }

  async shutdown(): Promise<void> {
    if (this.isShutdown) return;

    this.isShutdown = true;
    this.stopTimer();

    if (this.queue.length > 0) {
      await this.flush();
    }

    this.log("debug", "StackTrace SDK shut down");
  }

  private enqueue(level: LogLevel, message: string, options?: LogOptions): void {
    if (this.isShutdown) return;

    const entry: LogEntry = {
      level,
      message,
      service: this.service,
      timestamp: new Date().toISOString(),
      metadata: {
        ...options?.metadata,
        environment: this.environment,
      },
    };

    if (options?.traceId) {
      entry.trace_id = options.traceId;
    }

    this.queue.push(entry);

    if (this.queue.length >= this.batchSize) {
      void this.flush();
    }
  }

  private async sendBatch(batch: LogEntry[]): Promise<void> {
    let lastError: Error | null = null;

    for (let attempt = 0; attempt <= this.maxRetries; attempt++) {
      try {
        const response = await fetch(`${this.baseUrl}/logs/batch`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${this.apiKey}`,
          },
          body: JSON.stringify(batch),
        });

        if (response.ok) {
          return;
        }

        lastError = new Error(`HTTP ${response.status}: ${response.statusText}`);
      } catch (err) {
        lastError = err instanceof Error ? err : new Error(String(err));
      }

      if (attempt < this.maxRetries) {
        const delay = Math.pow(2, attempt) * 1000;
        this.log("debug", `Retry ${attempt + 1}/${this.maxRetries} in ${delay}ms`);
        await this.sleep(delay);
      }
    }

    throw lastError ?? new Error("Failed to send batch");
  }

  private startTimer(): void {
    this.timer = setInterval(() => {
      void this.flush();
    }, this.flushIntervalMs);

    if (this.timer && typeof this.timer === "object" && "unref" in this.timer) {
      this.timer.unref();
    }
  }

  private stopTimer(): void {
    if (this.timer) {
      clearInterval(this.timer);
      this.timer = null;
    }
  }

  private setupExceptionHandlers(): void {
    process.on("uncaughtException", (err) => {
      this.error(`Uncaught Exception: ${err.message}`, {
        metadata: { stack: err.stack, name: err.name },
      });
      void this.flush();
    });

    process.on("unhandledRejection", (reason) => {
      const message =
        reason instanceof Error ? reason.message : String(reason);
      const stack = reason instanceof Error ? reason.stack : undefined;

      this.error(`Unhandled Rejection: ${message}`, {
        metadata: { stack, reason: String(reason) },
      });
      void this.flush();
    });

    process.on("beforeExit", () => {
      void this.shutdown();
    });
  }

  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }

  private log(level: string, message: string): void {
    if (this.debug) {
      console.log(`[stacktrace-sdk] [${level}] ${message}`);
    }
  }
}
