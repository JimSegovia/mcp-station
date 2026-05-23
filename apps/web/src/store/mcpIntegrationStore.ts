import { create } from "zustand";

export type IntegrationStatus = "empty" | "connecting" | "connected" | "error";

export type HealthLevel = "healthy" | "degraded" | "unhealthy";

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
      const data = await res.json();
      set({
        status: data.status ?? "empty",
        endpoint: data.endpoint ?? "",
        tools: data.tools ?? [],
        lastConnected: data.lastConnected ?? null,
        lastError: data.lastError ?? null,
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
        set({ status: "error", lastError: { message: "Connection failed", timestamp: new Date().toISOString() } });
        return;
      }
      load();
    } catch {
      set({ status: "error", lastError: { message: "Network error", timestamp: new Date().toISOString() } });
    }
  },

  disconnect: async () => {
    const { load } = get();
    try {
      const res = await fetch("/api/integration/disconnect", { method: "POST" });
      if (!res.ok) return;
      load();
    } catch {
      // ignore
    }
  },

  toggleTool: async (toolName: string) => {
    const { tools } = get();
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
    } catch {
      // ignore
    }
  },
}));
