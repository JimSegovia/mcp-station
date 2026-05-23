import { create } from "zustand";

export type IntegrationStatus =
  | "empty"
  | "connecting"
  | "connected"
  | "disconnected"
  | "error";

export type HealthLevel =
  | "healthy"
  | "degraded"
  | "unhealthy"
  | "pending"
  | "info"
  | "unknown";

export interface HealthCheck {
  label: string;
  level: HealthLevel;
  detail: string;
}

export interface McpTool {
  name: string;
  description: string;
  enabled: boolean;
}

export interface ConnectionError {
  message: string;
  timestamp: string;
}

interface IntegrationApiResponse {
  status?: string;
  endpoint?: string;
  tools?: McpTool[];
  lastConnected?: string | null;
  lastError?: string | null;
  serverPort?: number;
  protocolVersion?: string;
  uptime?: number;
  latency?: number;
  healthChecks?: HealthCheck[];
  updatedAt?: string;
}

function normalizeStatus(status?: string): IntegrationStatus {
  switch (status) {
    case "connecting":
    case "connected":
    case "disconnected":
    case "error":
    case "empty":
      return status;
    default:
      return "empty";
  }
}

function normalizeError(data: IntegrationApiResponse): ConnectionError | null {
  if (!data.lastError) return null;
  return {
    message: data.lastError,
    timestamp: data.updatedAt ?? new Date().toISOString(),
  };
}

interface McpIntegrationStore {
  status: IntegrationStatus;
  endpoint: string;
  tools: McpTool[];
  lastConnected: string | null;
  lastError: ConnectionError | null;

  serverPort: number;
  protocolVersion: string;
  uptime: number;
  latency: number;
  healthChecks: HealthCheck[];

  setEndpoint: (url: string) => void;
  connect: () => Promise<void>;
  disconnect: () => Promise<void>;
  toggleTool: (toolName: string) => Promise<void>;
  load: () => Promise<void>;
}

export const useMcpIntegrationStore = create<McpIntegrationStore>((set, get) => ({
  status: "empty",
  endpoint: "",
  tools: [],
  lastConnected: null,
  lastError: null,

  serverPort: 8090,
  protocolVersion: "MCP/JSON-RPC 2.0",
  uptime: 0,
  latency: 0,
  healthChecks: [],

  setEndpoint: (url: string) => set({ endpoint: url }),

  load: async () => {
    try {
      const res = await fetch("/api/integration");
      if (!res.ok) return;
      const data: IntegrationApiResponse = await res.json();
      set({
        status: normalizeStatus(data.status),
        endpoint: data.endpoint ?? "",
        tools: data.tools ?? [],
        lastConnected: data.lastConnected ?? null,
        lastError: normalizeError(data),
        serverPort: data.serverPort ?? 8090,
        protocolVersion: data.protocolVersion ?? "MCP/JSON-RPC 2.0",
        uptime: data.uptime ?? 0,
        latency: data.latency ?? 0,
        healthChecks: data.healthChecks ?? [],
      });
    } catch {
      // backend not available
    }
  },

  connect: async () => {
    const { endpoint, load } = get();
    if (!endpoint) {
      set({ status: "error", lastError: { message: "No endpoint configured", timestamp: new Date().toISOString() } });
      return;
    }
    set({ status: "connecting" });
    try {
      const res = await fetch("/api/integration/connect", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ endpoint }),
      });
      if (!res.ok) {
        let message = "Connection failed";
        try {
          const data = await res.json();
          if (typeof data?.error === "string" && data.error) {
            message = data.error;
          }
        } catch {
          // ignore parse errors
        }
        set({ status: "error", lastError: { message, timestamp: new Date().toISOString() } });
        return;
      }
      await load();
    } catch {
      set({ status: "error", lastError: { message: "Network error", timestamp: new Date().toISOString() } });
    }
  },

  disconnect: async () => {
    const { load } = get();
    try {
      const res = await fetch("/api/integration/disconnect", { method: "POST" });
      if (!res.ok) return;
      await load();
    } catch {
      // ignore
    }
  },

  toggleTool: async (toolName: string) => {
    const { tools, load } = get();
    const tool = tools.find((t) => t.name === toolName);
    if (!tool) return;

    const newEnabled = !tool.enabled;
    try {
      const res = await fetch(`/api/integration/tools/${encodeURIComponent(toolName)}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ enabled: newEnabled }),
      });
      if (!res.ok) return;
      const data = await res.json();
      set({
        tools: data.tools ?? tools.map((t) => (t.name === toolName ? { ...t, enabled: newEnabled } : t)),
      });
      await load();
    } catch {
      // ignore
    }
  },
}));
