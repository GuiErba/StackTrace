import { useState } from "react";
import { useLogs } from "../hooks/useLogs";
import { useProject } from "../context/ProjectContext";
import { Search, ChevronDown, ChevronRight } from "lucide-react";

export default function Logs() {
  const { selectedProject } = useProject();
  const [level, setLevel] = useState("");
  const [service, setService] = useState("");
  const [cursor, setCursor] = useState("");
  const [expanded, setExpanded] = useState<number | null>(null);

  const { data, isLoading, refetch } = useLogs(selectedProject?.id, {
    level: level || undefined,
    service: service || undefined,
    cursor: cursor || undefined,
    limit: 50,
  });

  const handleFilter = () => {
    setCursor("");
    refetch();
  };

  return (
    <div>
      <h1 className="page-title">Logs</h1>

      <div className="card mb-4">
        <div className="flex gap-3 items-center flex-wrap">
          <select
            className="select w-[140px]"
            value={level}
            onChange={(e) => setLevel(e.target.value)}
          >
            <option value="">All levels</option>
            <option value="info">Info</option>
            <option value="warn">Warning</option>
            <option value="error">Error</option>
          </select>

          <input
            className="input w-[200px]"
            placeholder="Filter by service..."
            value={service}
            onChange={(e) => setService(e.target.value)}
          />

          <button className="btn btn-primary" onClick={handleFilter}>
            <Search size={14} />
            Filter
          </button>
        </div>
      </div>

      <div className="card !p-0 overflow-hidden">
        <table className="w-full border-collapse text-[13px]">
          <thead>
            <tr className="border-b border-(--color-border) text-(--color-text-secondary)">
              <th className="p-3 px-4 text-left font-medium"></th>
              <th className="p-3 px-4 text-left font-medium">Timestamp</th>
              <th className="p-3 px-4 text-left font-medium">Level</th>
              <th className="p-3 px-4 text-left font-medium">Service</th>
              <th className="p-3 px-4 text-left font-medium">Message</th>
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              <tr>
                <td colSpan={5} className="p-10 text-center text-(--color-text-secondary)">
                  Loading...
                </td>
              </tr>
            ) : !data?.data?.length ? (
              <tr>
                <td colSpan={5} className="p-10 text-center text-(--color-text-secondary)">
                  No logs found.
                </td>
              </tr>
            ) : (
              data.data.map((log) => (
                <>
                  <tr
                    key={log.id}
                    onClick={() => setExpanded(expanded === log.id ? null : log.id)}
                    className={`border-b border-(--color-border) cursor-pointer transition-colors duration-100 ${
                      expanded === log.id ? "bg-(--color-bg-hover)" : "hover:bg-(--color-bg-hover)"
                    }`}
                  >
                    <td className="py-2.5 px-4 w-[30px]">
                      {expanded === log.id ? (
                        <ChevronDown size={14} className="text-(--color-text-muted)" />
                      ) : (
                        <ChevronRight size={14} className="text-(--color-text-muted)" />
                      )}
                    </td>
                    <td className="py-2.5 px-4 text-(--color-text-secondary) whitespace-nowrap">
                      {new Date(log.timestamp).toLocaleString()}
                    </td>
                    <td className="py-2.5 px-4">
                      <span className={`badge badge-${log.level}`}>{log.level}</span>
                    </td>
                    <td className="py-2.5 px-4 text-(--color-text-secondary)">
                      {log.service}
                    </td>
                    <td className="py-2.5 px-4 max-w-[400px] overflow-hidden text-ellipsis whitespace-nowrap">
                      {log.message}
                    </td>
                  </tr>
                  {expanded === log.id && (
                    <tr key={`${log.id}-detail`}>
                      <td colSpan={5} className="p-4 bg-(--color-bg-hover)">
                        <div className="text-xs">
                          {log.trace_id && (
                            <p className="mb-2">
                              <strong className="text-(--color-text-secondary)">Trace ID:</strong>{" "}
                              <code className="text-(--color-accent)">{log.trace_id}</code>
                            </p>
                          )}
                          {log.metadata && (
                            <div>
                              <strong className="text-(--color-text-secondary)">Metadata:</strong>
                              <pre className="mt-1.5 p-2.5 bg-(--color-bg-primary) rounded-md text-xs overflow-auto">
                                {JSON.stringify(log.metadata, null, 2)}
                              </pre>
                            </div>
                          )}
                        </div>
                      </td>
                    </tr>
                  )}
                </>
              ))
            )}
          </tbody>
        </table>
      </div>

      {data?.has_more && (
        <div className="text-center mt-4">
          <button
            className="btn btn-ghost"
            onClick={() => setCursor(data.next_cursor)}
          >
            Load more
          </button>
        </div>
      )}
    </div>
  );
}
