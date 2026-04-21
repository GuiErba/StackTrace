import { useMetrics } from "../hooks/useMetrics";
import { useProject } from "../context/ProjectContext";
import { useQuery } from "@tanstack/react-query";
import { api } from "../lib/api";
import {
  AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer,
} from "recharts";
import { Activity, AlertTriangle, ScrollText } from "lucide-react";
import type { Incident } from "../types/incidents";
import { useTitle } from "../hooks/useTitle";

export default function Overview() {
  useTitle("Overview");
  const { selectedProject } = useProject();
  const { data: metrics, isLoading } = useMetrics(selectedProject?.id);

  const { data: incidents } = useQuery({
    queryKey: ["incidents", selectedProject?.id],
    queryFn: () =>
      api.get<{ data: Incident[] }>(
        `/dashboard/incidents?project_id=${selectedProject!.id}`
      ),
    enabled: !!selectedProject,
  });

  if (!selectedProject) {
    return (
      <div className="text-center py-15 px-5 text-(--color-text-secondary)">
        <p>Select or create a project to get started.</p>
      </div>
    );
  }

  const errorRate = metrics && metrics.total_logs > 0
    ? ((metrics.error_count / metrics.total_logs) * 100).toFixed(1)
    : "0.0";

  const chartData = (metrics?.logs_per_hour || []).map((d) => ({
    hour: new Date(d.hour).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" }),
    count: d.count,
  }));

  const openIncidents = (incidents?.data || []).filter((i) => i.status === "open");

  return (
    <div>
      <h1 className="page-title">Overview</h1>

      <div className="grid grid-cols-3 gap-4 mb-8">
        <div className="card flex items-center gap-4">
          <div className="p-2.5 rounded-[10px] bg-(--color-info-bg)">
            <ScrollText size={20} color="#60a5fa" />
          </div>
          <div>
            <p className="text-xs text-(--color-text-secondary) mb-1">
              Total Logs (24h)
            </p>
            <p className="text-[28px] font-bold">
              {isLoading ? "..." : metrics?.total_logs.toLocaleString()}
            </p>
          </div>
        </div>

        <div className="card flex items-center gap-4">
          <div className="p-2.5 rounded-[10px] bg-(--color-error-bg)">
            <AlertTriangle size={20} color="#f87171" />
          </div>
          <div>
            <p className="text-xs text-(--color-text-secondary) mb-1">
              Errors (24h)
            </p>
            <p className="text-[28px] font-bold text-(--color-danger)">
              {isLoading ? "..." : metrics?.error_count.toLocaleString()}
            </p>
          </div>
        </div>

        <div className="card flex items-center gap-4">
          <div className="p-2.5 rounded-[10px] bg-(--color-warn-bg)">
            <Activity size={20} color="#fbbf24" />
          </div>
          <div>
            <p className="text-xs text-(--color-text-secondary) mb-1">
              Error Rate
            </p>
            <p className="text-[28px] font-bold text-(--color-warning)">
              {isLoading ? "..." : `${errorRate}%`}
            </p>
          </div>
        </div>
      </div>

      <div className="card mb-8">
        <h2 className="section-title">Logs per hour (24h)</h2>
        <div className="h-[300px]">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={chartData}>
              <defs>
                <linearGradient id="colorCount" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#6366f1" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="#6366f1" stopOpacity={0} />
                </linearGradient>
              </defs>
              <XAxis
                dataKey="hour"
                stroke="#5b5b73"
                fontSize={12}
                tickLine={false}
                axisLine={false}
              />
              <YAxis
                stroke="#5b5b73"
                fontSize={12}
                tickLine={false}
                axisLine={false}
                allowDecimals={false}
              />
              <Tooltip
                contentStyle={{
                  background: "#16161f",
                  border: "1px solid #2a2a3a",
                  borderRadius: "8px",
                  color: "#f1f1f6",
                  fontSize: "13px",
                }}
              />
              <Area
                type="monotone"
                dataKey="count"
                stroke="#6366f1"
                strokeWidth={2}
                fill="url(#colorCount)"
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      </div>

      <div className="card">
        <h2 className="section-title">
          Open Incidents ({openIncidents.length})
        </h2>
        {openIncidents.length === 0 ? (
          <p className="text-muted">
            No open incidents — all systems operational.
          </p>
        ) : (
          <div className="flex flex-col gap-2">
            {openIncidents.map((inc) => (
              <div
                key={inc.id}
                className="flex items-center justify-between p-3 bg-(--color-bg-hover) rounded-lg"
              >
                <div className="flex items-center gap-2.5">
                  <span className="badge badge-open">open</span>
                  <span className="text-sm">{inc.title}</span>
                </div>
                <span className="text-xs text-(--color-text-muted)">
                  {new Date(inc.started_at).toLocaleString()}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
