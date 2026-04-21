import { useState } from "react";
import { useTitle } from "../hooks/useTitle";
import { useMutation } from "@tanstack/react-query";
import { useProject } from "../context/ProjectContext";
import { api } from "../lib/api";
import { Copy, Check, FolderPlus, RotateCw, Key } from "lucide-react";

export default function Projects() {
  useTitle("Projects");
  const { projects, refreshProjects, selectProject, selectedProject } = useProject();
  const [showCreate, setShowCreate] = useState(false);
  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [newApiKey, setNewApiKey] = useState("");
  const [copied, setCopied] = useState(false);

  const createMutation = useMutation({
    mutationFn: (data: { name: string; slug: string }) =>
      api.post<{ id: string; name: string; slug: string; api_key: string }>(
        "/projects",
        data
      ),
    onSuccess: (data) => {
      setNewApiKey(data.api_key);
      setShowCreate(false);
      setName("");
      setSlug("");
      refreshProjects();
    },
  });

  const rotateMutation = useMutation({
    mutationFn: (projectId: string) =>
      api.post<{ api_key: string }>(`/projects/${projectId}/rotate-key`, {}),
    onSuccess: (data) => {
      setNewApiKey(data.api_key);
      refreshProjects();
    },
  });

  const copyKey = () => {
    navigator.clipboard.writeText(newApiKey);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault();
    createMutation.mutate({ name, slug });
  };

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-[22px] font-bold">Projects</h1>
        <button className="btn btn-primary" onClick={() => setShowCreate(!showCreate)}>
          <FolderPlus size={16} />
          New Project
        </button>
      </div>

      {newApiKey && (
        <div className="card mb-6 border-amber-500! bg-(--color-warn-bg)">
          <div className="flex items-center gap-2 mb-2">
            <Key size={16} color="#fbbf24" />
            <h3 className="text-sm font-semibold text-amber-300">
              Save your API Key now
            </h3>
          </div>
          <p className="text-xs text-(--color-text-secondary) mb-3">
            This key will only be shown once. Store it securely.
          </p>
          <div className="flex items-center gap-2 p-2.5 bg-(--color-bg-primary) rounded-md font-mono text-[13px] break-all">
            <code className="flex-1">{newApiKey}</code>
            <button
              className="btn btn-ghost py-1.5 px-2.5 shrink-0"
              onClick={copyKey}
            >
              {copied ? <Check size={14} color="var(--color-success)" /> : <Copy size={14} />}
            </button>
          </div>
        </div>
      )}

      {showCreate && (
        <div className="card mb-6">
          <h2 className="section-title">Create Project</h2>
          <form onSubmit={handleCreate} className="flex flex-col gap-3">
            <div>
              <label className="form-label">Project name</label>
              <input
                className="input"
                placeholder="My SaaS"
                value={name}
                onChange={(e) => {
                  setName(e.target.value);
                  setSlug(e.target.value.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/(^-|-$)/g, ""));
                }}
                required
              />
            </div>
            <div>
              <label className="form-label">Slug (for status page URL)</label>
              <input
                className="input"
                placeholder="my-saas"
                value={slug}
                onChange={(e) => setSlug(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, ""))}
                required
              />
            </div>
            <div className="flex gap-2">
              <button className="btn btn-primary" type="submit" disabled={createMutation.isPending}>
                {createMutation.isPending ? "Creating..." : "Create Project"}
              </button>
              <button className="btn btn-ghost" type="button" onClick={() => setShowCreate(false)}>
                Cancel
              </button>
            </div>
          </form>
        </div>
      )}

      <div className="flex flex-col gap-3">
        {projects.map((project) => (
          <div
            key={project.id}
            className={`card flex items-center justify-between cursor-pointer ${
              selectedProject?.id === project.id
                ? "border-[var(--color-accent)]!"
                : ""
            }`}
            onClick={() => selectProject(project)}
          >
            <div>
              <p className="text-[15px] font-medium mb-1">
                {project.name}
                {selectedProject?.id === project.id && (
                  <span className="text-[11px] text-(--color-accent) ml-2">
                    ACTIVE
                  </span>
                )}
              </p>
              <p className="text-xs text-(--color-text-muted)">
                {project.api_key_prefix && `Key: ${project.api_key_prefix}...`}
                {project.slug && ` · Status: /status/${project.slug}`}
              </p>
            </div>
            <button
              className="btn btn-ghost text-xs py-1.5 px-3"
              onClick={(e) => {
                e.stopPropagation();
                if (confirm("This will invalidate the current API key. Continue?")) {
                  rotateMutation.mutate(project.id);
                }
              }}
            >
              <RotateCw size={14} />
              Rotate Key
            </button>
          </div>
        ))}
      </div>

      {projects.length === 0 && !showCreate && (
        <div className="card text-center !p-10">
          <FolderPlus size={32} className="text-(--color-text-muted) mb-3 mx-auto" />
          <p className="text-(--color-text-secondary) mb-4">
            Create your first project to start sending logs.
          </p>
          <button className="btn btn-primary" onClick={() => setShowCreate(true)}>
            Create Project
          </button>
        </div>
      )}
    </div>
  );
}
