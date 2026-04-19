import { createContext, useContext, useState, useEffect, type ReactNode } from "react";
import { api } from "../lib/api";

export interface Project {
  id: string;
  name: string;
  slug?: string;
  api_key_prefix?: string;
  owner_email: string;
}

interface ProjectContextType {
  projects: Project[];
  selectedProject: Project | null;
  selectProject: (project: Project) => void;
  refreshProjects: () => Promise<void>;
  isLoading: boolean;
  hasLoaded: boolean;
}

const ProjectContext = createContext<ProjectContextType | null>(null);

export function ProjectProvider({ children }: { children: ReactNode }) {
  const [projects, setProjects] = useState<Project[]>([]);
  const [selectedProject, setSelectedProject] = useState<Project | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [hasLoaded, setHasLoaded] = useState(false);

  const refreshProjects = async () => {
    setIsLoading(true);
    try {
      const res = await api.get<{ data: Project[] }>("/projects");
      const data = res.data || [];
      setProjects(data);
      setHasLoaded(true);

      if (data.length > 0) {
        const saved = localStorage.getItem("stacktrace_project");
        const found = saved ? data.find((p) => p.id === saved) : null;
        setSelectedProject(found || data[0]);
      }
    } catch (err) {
      console.error("Failed to load projects:", err);
      // Don't reset projects on error — keep whatever we had
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    refreshProjects();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const selectProject = (project: Project) => {
    setSelectedProject(project);
    localStorage.setItem("stacktrace_project", project.id);
  };

  return (
    <ProjectContext.Provider
      value={{
        projects,
        selectedProject,
        selectProject,
        refreshProjects,
        isLoading,
        hasLoaded,
      }}
    >
      {children}
    </ProjectContext.Provider>
  );
}

export function useProject() {
  const context = useContext(ProjectContext);
  if (!context) {
    throw new Error("useProject must be used within a ProjectProvider");
  }
  return context;
}
