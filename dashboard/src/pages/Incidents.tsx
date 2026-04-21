import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useProject } from "../context/ProjectContext";
import { api } from "../lib/api";
import { CheckCircle } from "lucide-react";
import type { Incident } from "../types/incidents";
import { useTitle } from "../hooks/useTitle";

export default function Incidents() {
  useTitle("Incidents");
  const { selectedProject } = useProject();
  const queryClient = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ["incidents", selectedProject?.id],
    queryFn: () =>
      api.get<{ data: Incident[] }>(
        `/dashboard/incidents?project_id=${selectedProject!.id}`
      ),
    enabled: !!selectedProject,
  });

  const resolveMutation = useMutation({
    mutationFn: (id: string) =>
      api.patch(`/dashboard/incidents/${id}/resolve?project_id=${selectedProject!.id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["incidents"] });
    },
  });

  const incidents = data?.data || [];
  const openIncidents = incidents.filter((i) => i.status === "open");
  const resolvedIncidents = incidents.filter((i) => i.status === "resolved");

  return (
    <div>
      <h1 className="page-title">Incidents</h1>

      <div className="card mb-6">
        <h2 className="section-title">Open ({openIncidents.length})</h2>
        {isLoading ? (
          <p className="text-muted">Loading...</p>
        ) : openIncidents.length === 0 ? (
          <p className="text-muted">
            No open incidents — all systems operational. ✅
          </p>
        ) : (
          <div className="flex flex-col gap-2.5">
            {openIncidents.map((inc) => (
              <div
                key={inc.id}
                className="flex items-center justify-between p-3.5 bg-(--color-bg-hover) rounded-lg border border-(--color-border)"
              >
                <div>
                  <div className="flex items-center gap-2.5 mb-1">
                    <span className="badge badge-open">open</span>
                    <span className="text-sm font-medium">{inc.title}</span>
                  </div>
                  <p className="text-xs text-(--color-text-muted)">
                    {inc.description} · Started {new Date(inc.started_at).toLocaleString()}
                  </p>
                </div>
                <button
                  className="btn btn-ghost text-[13px]"
                  onClick={() => resolveMutation.mutate(inc.id)}
                  disabled={resolveMutation.isPending}
                >
                  <CheckCircle size={14} />
                  Resolve
                </button>
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="card">
        <h2 className="section-title">Resolved ({resolvedIncidents.length})</h2>
        {resolvedIncidents.length === 0 ? (
          <p className="text-muted">No incident history yet.</p>
        ) : (
          <div className="flex flex-col gap-2">
            {resolvedIncidents.map((inc) => (
              <div
                key={inc.id}
                className="flex items-center justify-between p-3 bg-(--color-bg-hover) rounded-lg"
              >
                <div className="flex items-center gap-2.5">
                  <span className="badge badge-resolved">resolved</span>
                  <span className="text-sm">{inc.title}</span>
                </div>
                <span className="text-xs text-(--color-text-muted)">
                  {new Date(inc.started_at).toLocaleDateString()} →{" "}
                  {inc.resolved_at && new Date(inc.resolved_at).toLocaleDateString()}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
