import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import { useProject } from "../context/ProjectContext";
import {
  LayoutDashboard,
  ScrollText,
  AlertTriangle,
  Bell,
  FolderOpen,
  LogOut,
  FolderPlus,
  Shield,
} from "lucide-react";
import { useState } from "react";
import { useTitle } from "../hooks/useTitle";
import { useMutation } from "@tanstack/react-query";
import { api } from "../lib/api";
import { Copy, Check, Key } from "lucide-react";

const navItems = [
  { to: "/", icon: LayoutDashboard, label: "Overview" },
  { to: "/logs", icon: ScrollText, label: "Logs" },
  { to: "/incidents", icon: AlertTriangle, label: "Incidents" },
  { to: "/alert-rules", icon: Bell, label: "Alert Rules" },
  { to: "/projects", icon: FolderOpen, label: "Projects" },
];

function OnboardingScreen() {
  useTitle("Welcome");
  const { refreshProjects } = useProject();
  const { user, logout } = useAuth();
  const navigate = useNavigate();
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
    <div className="min-h-screen flex items-center justify-center p-5">
      <div className="w-full max-w-[480px]">
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-14 h-14 bg-(--color-accent) rounded-[14px] mb-4">
            <Shield size={28} color="white" />
          </div>
          <h1 className="text-2xl font-bold mb-2">Welcome to StackTrace</h1>
          <p className="text-(--color-text-secondary) text-sm">
            Create your first project to start monitoring your application.
          </p>
        </div>

        {newApiKey ? (
          <div className="card">
            <div className="flex items-center gap-2 mb-2">
              <Key size={16} color="#fbbf24" />
              <h3 className="text-sm font-semibold text-amber-300">
                Your API Key
              </h3>
            </div>
            <p className="text-xs text-(--color-text-secondary) mb-3">
              Copy and save this key now. You will not be able to see it again.
            </p>
            <div className="flex items-center gap-2 p-2.5 bg-(--color-bg-primary) rounded-md font-mono text-xs break-all mb-4">
              <code className="flex-1">{newApiKey}</code>
              <button
                className="btn btn-ghost py-1.5 px-2.5 shrink-0"
                onClick={copyKey}
              >
                {copied ? (
                  <Check size={14} color="var(--color-success)" />
                ) : (
                  <Copy size={14} />
                )}
              </button>
            </div>
            <button
              className="btn btn-primary w-full"
              onClick={() => {
                refreshProjects().then(() => navigate("/"));
              }}
            >
              Go to Dashboard
            </button>
          </div>
        ) : (
          <div className="card">
            <form
              onSubmit={handleCreate}
              className="flex flex-col gap-3.5"
            >
              <div>
                <label className="form-label">
                  Project name
                </label>
                <input
                  className="input"
                  placeholder="My SaaS"
                  value={name}
                  onChange={(e) => {
                    setName(e.target.value);
                    setSlug(
                      e.target.value
                        .toLowerCase()
                        .replace(/[^a-z0-9]+/g, "-")
                        .replace(/(^-|-$)/g, "")
                    );
                  }}
                  required
                  autoFocus
                />
              </div>
              <div>
                <label className="form-label">
                  Slug
                  <span className="font-normal text-(--color-text-muted) ml-1.5">
                    (for your public status page)
                  </span>
                </label>
                <input
                  className="input"
                  placeholder="my-saas"
                  value={slug}
                  onChange={(e) =>
                    setSlug(
                      e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, "")
                    )
                  }
                  required
                />
              </div>
              <button
                className="btn btn-primary w-full mt-1"
                type="submit"
                disabled={createMutation.isPending || !name || !slug}
              >
                <FolderPlus size={16} />
                {createMutation.isPending
                  ? "Creating..."
                  : "Create Project"}
              </button>
            </form>
          </div>
        )}

        <div className="text-center mt-5">
          <button
            onClick={() => {
              logout();
              navigate("/login");
            }}
            className="bg-transparent border-none text-(--color-text-muted) text-[13px] cursor-pointer"
          >
            Sign out ({user?.email})
          </button>
        </div>
      </div>
    </div>
  );
}

export default function Layout() {
  const { user, logout } = useAuth();
  const { projects, isLoading, hasLoaded } = useProject();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate("/login");
  };

  if (isLoading && !hasLoaded) {
    return (
      <div className="min-h-screen flex items-center justify-center text-(--color-text-secondary)">
        Loading...
      </div>
    );
  }

  if (hasLoaded && projects.length === 0) {
    return <OnboardingScreen />;
  }

  return (
    <div className="flex min-h-screen">
      <aside className="w-60 bg-(--color-bg-secondary) border-r border-(--color-border) flex flex-col fixed top-0 left-0 bottom-0 z-50">
        <div className="p-5 border-b border-(--color-border)">
          <h1 className="text-lg font-bold">StackTrace</h1>
          <p className="text-xs text-(--color-text-muted) mt-1">
            Observability Platform
          </p>
        </div>

        <nav className="flex-1 p-3">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.to === "/"}
              className={({ isActive }) =>
                `flex items-center gap-2.5 px-3 py-2.5 rounded-lg text-sm no-underline mb-1 transition-all duration-150 ${
                  isActive
                    ? "text-(--color-text-primary) bg-(--color-bg-hover) font-medium"
                    : "text-(--color-text-secondary) font-normal hover:bg-(--color-bg-hover)"
                }`
              }
            >
              <item.icon size={18} />
              {item.label}
            </NavLink>
          ))}
        </nav>

        <div className="p-4 border-t border-(--color-border)">
          <p className="text-[13px] text-(--color-text-secondary) mb-2 overflow-hidden text-ellipsis whitespace-nowrap">
            {user?.email}
          </p>
          <button
            onClick={handleLogout}
            className="btn btn-ghost w-full text-[13px] py-2"
          >
            <LogOut size={14} />
            Sign out
          </button>
        </div>
      </aside>

      <main className="flex-1 ml-60 p-8 min-h-screen">
        <Outlet />
      </main>
    </div>
  );
}
