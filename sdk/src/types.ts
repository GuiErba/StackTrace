export type LogLevel = "info" | "warn" | "error";

export interface StackTraceConfig {
  apiKey: string;
  service: string;
  baseUrl?: string;
  environment?: string;
  batchSize?: number;
  flushIntervalMs?: number;
  maxRetries?: number;
  captureExceptions?: boolean;
  debug?: boolean;
}

export interface LogEntry {
  level: LogLevel;
  message: string;
  service: string;
  timestamp?: string;
  trace_id?: string;
  metadata?: Record<string, unknown>;
}

export interface BatchResponse {
  status: string;
  accepted: number;
  dropped: number;
}

export interface LogOptions {
  traceId?: string;
  metadata?: Record<string, unknown>;
}
