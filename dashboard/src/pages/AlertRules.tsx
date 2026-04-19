import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useProject } from "../context/ProjectContext";
import { api } from "../lib/api";
import { Plus, Trash2 } from "lucide-react";

interface AlertRule {
  id: string;
  condition: string;
  threshold: number;
  window_seconds: number;
  channel: string;
  destination: string;
}

export default function AlertRules() {
  const { selectedProject } = useProject();
  const queryClient = useQueryClient();
  const [showForm, setShowForm] = useState(false);
  const [condition, setCondition] = useState("error_count");
  const [threshold, setThreshold] = useState("5");
  const [windowSeconds, setWindowSeconds] = useState("60");
  const [channel, setChannel] = useState("email");
  const [destination, setDestination] = useState("");

  const { data, isLoading } = useQuery({
    queryKey: ["alert-rules", selectedProject?.id],
    queryFn: () =>
      api.get<{ data: AlertRule[] }>(
        `/dashboard/alert-rules?project_id=${selectedProject!.id}`
      ),
    enabled: !!selectedProject,
  });

  const createMutation = useMutation({
    mutationFn: (rule: Omit<AlertRule, "id">) =>
      api.post(`/dashboard/alert-rules?project_id=${selectedProject!.id}`, rule),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["alert-rules"] });
      setShowForm(false);
      setDestination("");
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) =>
      api.del(`/dashboard/alert-rules/${id}?project_id=${selectedProject!.id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["alert-rules"] });
    },
  });

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault();
    createMutation.mutate({
      condition,
      threshold: parseInt(threshold),
      window_seconds: parseInt(windowSeconds),
      channel,
      destination,
    });
  };

  const rules = data?.data || [];

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-[22px] font-bold">Alert Rules</h1>
        <button className="btn btn-primary" onClick={() => setShowForm(!showForm)}>
          <Plus size={16} />
          Create Rule
        </button>
      </div>

      {showForm && (
        <div className="card mb-6">
          <h2 className="section-title">New Alert Rule</h2>
          <form onSubmit={handleCreate} className="grid grid-cols-2 gap-3">
            <div>
              <label className="form-label">Condition</label>
              <select className="select w-full" value={condition} onChange={(e) => setCondition(e.target.value)}>
                <option value="error_count">Error count</option>
              </select>
            </div>
            <div>
              <label className="form-label">Threshold</label>
              <input className="input" type="number" min="1" value={threshold} onChange={(e) => setThreshold(e.target.value)} />
            </div>
            <div>
              <label className="form-label">Window (seconds)</label>
              <input className="input" type="number" min="10" max="3600" value={windowSeconds} onChange={(e) => setWindowSeconds(e.target.value)} />
            </div>
            <div>
              <label className="form-label">Channel</label>
              <select className="select w-full" value={channel} onChange={(e) => setChannel(e.target.value)}>
                <option value="email">Email</option>
              </select>
            </div>
            <div className="col-span-full">
              <label className="form-label">Destination (email)</label>
              <input className="input" type="email" placeholder="alerts@company.com" value={destination} onChange={(e) => setDestination(e.target.value)} required />
            </div>
            <div className="col-span-full flex gap-2">
              <button className="btn btn-primary" type="submit" disabled={createMutation.isPending}>
                {createMutation.isPending ? "Creating..." : "Create"}
              </button>
              <button className="btn btn-ghost" type="button" onClick={() => setShowForm(false)}>
                Cancel
              </button>
            </div>
          </form>
        </div>
      )}

      <div className="card">
        {isLoading ? (
          <p className="text-muted">Loading...</p>
        ) : rules.length === 0 ? (
          <p className="text-muted">
            No alert rules configured. Create one to get notified when errors spike.
          </p>
        ) : (
          <div className="flex flex-col gap-2.5">
            {rules.map((rule) => (
              <div
                key={rule.id}
                className="flex items-center justify-between p-3.5 bg-(--color-bg-hover) rounded-lg"
              >
                <div>
                  <p className="text-sm font-medium mb-1">
                    {rule.condition === "error_count" ? "Error count" : rule.condition} ≥ {rule.threshold} in {rule.window_seconds}s
                  </p>
                  <p className="text-xs text-(--color-text-muted)">
                    {rule.channel}: {rule.destination}
                  </p>
                </div>
                <button
                  className="btn btn-danger py-2 px-3"
                  onClick={() => deleteMutation.mutate(rule.id)}
                >
                  <Trash2 size={14} />
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
