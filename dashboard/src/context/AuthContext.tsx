import { createContext, useContext, useState, useEffect, type ReactNode } from "react";
import { api } from "../lib/api";

export interface User {
  id: string;
  email: string;
}

interface AuthContextType {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  sendCode: (email: string) => Promise<void>;
  verifyCode: (email: string, code: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const savedToken = localStorage.getItem("stacktrace_token");
    const savedUser = localStorage.getItem("stacktrace_user");

    if (savedToken && savedUser) {
      try {
        setToken(savedToken);
        setUser(JSON.parse(savedUser));
      } catch {
        localStorage.removeItem("stacktrace_token");
        localStorage.removeItem("stacktrace_user");
      }
    }

    setIsLoading(false);
  }, []);

  const sendCode = async (email: string) => {
    await api.post("/auth/send-code", { email });
  };

  const verifyCode = async (email: string, code: string) => {
    const response = await api.post<{ token: string; user: User }>(
      "/auth/verify-code",
      { email, code }
    );

    setToken(response.token);
    setUser(response.user);
    localStorage.setItem("stacktrace_token", response.token);
    localStorage.setItem("stacktrace_user", JSON.stringify(response.user));
  };

  const logout = () => {
    setToken(null);
    setUser(null);
    localStorage.removeItem("stacktrace_token");
    localStorage.removeItem("stacktrace_user");
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        isAuthenticated: !!token,
        isLoading,
        sendCode,
        verifyCode,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
