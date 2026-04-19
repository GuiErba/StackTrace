import { useQuery } from "@tanstack/react-query";
import { api } from "../lib/api";

export interface LogEntry {
  id: number;
  project_id: string;
  timestamp: string;
  level: string;
  message: string;
  service: string;
  trace_id?: string;
  metadata?: Record<string, unknown>;
}

export interface LogResponse {
  data: LogEntry[];
  next_cursor: string;
  has_more: boolean;
}

export interface LogFilters {
  level?: string;
  service?: string;
  from?: string;
  to?: string;
  cursor?: string;
  limit?: number;
}

export function useLogs(projectId: string | undefined, filters: LogFilters = {}) {
  const params = new URLSearchParams();
  if (projectId) params.set("project_id", projectId);
  if (filters.level) params.set("level", filters.level);
  if (filters.service) params.set("service", filters.service);
  if (filters.from) params.set("from", filters.from);
  if (filters.to) params.set("to", filters.to);
  if (filters.cursor) params.set("cursor", filters.cursor);
  if (filters.limit) params.set("limit", String(filters.limit));

  return useQuery<LogResponse>({
    queryKey: ["logs", projectId, filters],
    queryFn: () => api.get<LogResponse>(`/dashboard/logs?${params.toString()}`),
    enabled: !!projectId,
  });
}
