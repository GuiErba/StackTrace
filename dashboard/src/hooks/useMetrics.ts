import { useQuery } from "@tanstack/react-query";
import { api } from "../lib/api";

export interface OverviewMetrics {
  total_logs: number;
  error_count: number;
  warn_count: number;
  info_count: number;
  logs_per_hour: { hour: string; count: number }[];
}

export function useMetrics(projectId: string | undefined) {
  return useQuery<OverviewMetrics>({
    queryKey: ["metrics", projectId],
    queryFn: () =>
      api.get<OverviewMetrics>(
        `/dashboard/metrics/overview?project_id=${projectId}`
      ),
    enabled: !!projectId,
    refetchInterval: 30000,
  });
}
