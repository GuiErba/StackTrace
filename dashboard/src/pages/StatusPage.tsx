import { useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { CheckCircle, AlertTriangle, Shield } from "lucide-react";
import { BASE_URL } from "../lib/api";
import { useTitle } from "../hooks/useTitle";

interface StatusData {
  project: string;
  status: string;
  incidents: {
    id: string;
    title: string;
    description: string;
    status: string;
    started_at: string;
    resolved_at?: string;
  }[];
}

export default function StatusPage() {
  const { slug } = useParams();

  const { data, isLoading, error } = useQuery<StatusData>({
    queryKey: ["status", slug],
    queryFn: async () => {
      const res = await fetch(`${BASE_URL}/status/${slug}`);
      if (!res.ok) throw new Error("Project not found");
      return res.json();
    },
    refetchInterval: 30000,
  });

  useTitle(data?.project ? `Status - ${data.project}` : "Status");

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <p className="text-(--color-text-secondary)">Loading...</p>
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <Shield size={40} className="text-(--color-text-muted) mb-4 mx-auto" />
          <h1 className="text-xl mb-2">Project not found</h1>
          <p className="text-(--color-text-secondary)">Check the URL and try again.</p>
        </div>
      </div>
    );
  }

  const isOperational = data.status === "operational";

  return (
    <div className="min-h-screen py-10 px-5">
      <div className="max-w-[640px] mx-auto">
        <div className="text-center mb-10">
          <h1 className="text-2xl font-bold mb-2">{data.project}</h1>
          <p className="text-(--color-text-secondary) text-sm">System Status</p>
        </div>

        <div
          className={`card text-center mb-8 ${
            isOperational ? "border-green-600!" : "border-[var(--color-danger)]!"
          }`}
        >
          <div className="flex items-center justify-center gap-3">
            {isOperational ? (
              <CheckCircle size={28} color="#22c55e" />
            ) : (
              <AlertTriangle size={28} color="#ef4444" />
            )}
            <span
              className={`text-[22px] font-bold ${
                isOperational ? "text-green-500" : "text-red-500"
              }`}
            >
              {isOperational ? "All Systems Operational" : "Active Incident"}
            </span>
          </div>
        </div>

        {data.incidents.length > 0 && (
          <div>
            <h2 className="text-base font-semibold mb-4">Recent Incidents</h2>
            <div className="flex flex-col gap-2.5">
              {data.incidents.map((inc) => (
                <div key={inc.id} className="card !p-4">
                  <div className="flex items-center gap-2.5 mb-1.5">
                    <span className={`badge badge-${inc.status}`}>
                      {inc.status}
                    </span>
                    <span className="font-medium text-sm">{inc.title}</span>
                  </div>
                  <p className="text-[13px] text-(--color-text-secondary) mb-1">
                    {inc.description}
                  </p>
                  <p className="text-xs text-(--color-text-muted)">
                    {new Date(inc.started_at).toLocaleString()}
                    {inc.resolved_at && ` — Resolved ${new Date(inc.resolved_at).toLocaleString()}`}
                  </p>
                </div>
              ))}
            </div>
          </div>
        )}

        <div className="text-center mt-12 text-(--color-text-muted) text-xs">
          Powered by <strong>StackTrace</strong>
        </div>
      </div>
    </div>
  );
}
