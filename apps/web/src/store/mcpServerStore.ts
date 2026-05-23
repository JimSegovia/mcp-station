import { create } from "zustand";

export type McpServerType = "stdio" | "websocket" | "virtual";
export type McpServerStatus = "connected" | "disconnected" | "error";

export interface McpServerTool {
  name: string;
  description: string;
  enabled: boolean;
}

interface McpServerStore {
  servers: McpServer[];

  addServer: (server: { name: string; type?: string; endpoint?: string }) => Promise<void>;
  removeServer: (id: string) => Promise<void>;
  toggleServer: (id: string) => Promise<void>;
  toggleTool: (serverId: string, toolName: string) => Promise<void>;
  discoverTools: (id: string) => Promise<void>;
  load: () => Promise<void>;
}

export interface McpServer {
  id: string;
  name: string;
  type: McpServerType;
  endpoint: string;
  enabled: boolean;
  status: McpServerStatus;
  tools: McpServerTool[];
  lastConnected: string | null;
  createdAt?: string;
  updatedAt?: string;
}

interface McpServerStore {
  servers: McpServer[];

  addServer: (server: { name: string; type?: string; endpoint?: string }) => Promise<void>;
  removeServer: (id: string) => Promise<void>;
  toggleServer: (id: string) => Promise<void>;
  toggleTool: (serverId: string, toolName: string) => Promise<void>;
  load: () => Promise<void>;
}

export const useMcpServerStore = create<McpServerStore>((set, get) => ({
  servers: [],

  load: async () => {
    try {
      const res = await fetch("/api/servers");
      if (!res.ok) return;
      const data: McpServer[] = await res.json();
      set({ servers: data });
    } catch {
      // backend not available
    }
  },

  addServer: async (server) => {
    try {
      const res = await fetch("/api/servers", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: server.name,
          type: server.type || "websocket",
          endpoint: server.endpoint || "",
        }),
      });
      if (!res.ok) return;
      const data: McpServer = await res.json();
      set((state) => ({ servers: [...state.servers, data] }));
    } catch {
      // ignore
    }
  },

  removeServer: async (id: string) => {
    try {
      await fetch(`/api/servers/${encodeURIComponent(id)}`, { method: "DELETE" });
      set((state) => ({ servers: state.servers.filter((s) => s.id !== id) }));
    } catch {
      // ignore
    }
  },

  toggleServer: async (id: string) => {
    const { servers } = get();
    const server = servers.find((s) => s.id === id);
    if (!server) return;

    const newEnabled = !server.enabled;
    try {
      const res = await fetch(`/api/servers/${encodeURIComponent(id)}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ enabled: newEnabled }),
      });
      if (!res.ok) return;
      const data: McpServer = await res.json();
      set({
        servers: servers.map((s) => (s.id === id ? { ...s, ...data } : s)),
      });
    } catch {
      // ignore
    }
  },

  toggleTool: async (serverId: string, toolName: string) => {
    const { servers } = get();
    const server = servers.find((s) => s.id === serverId);
    if (!server) return;
    const tool = server.tools.find((t) => t.name === toolName);
    if (!tool) return;

    const newEnabled = !tool.enabled;
    try {
      const res = await fetch(`/api/servers/${encodeURIComponent(serverId)}/tools/${encodeURIComponent(toolName)}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ enabled: newEnabled }),
      });
      if (!res.ok) return;
      const data: McpServer = await res.json();
      set({
        servers: servers.map((s) => (s.id === serverId ? { ...s, ...data } : s)),
      });
    } catch {
      // ignore
    }
  },

  discoverTools: async (id: string) => {
    try {
      const res = await fetch(`/api/servers/${encodeURIComponent(id)}/discover`, { method: "POST" });
      if (!res.ok) return;
      get().load();
    } catch {
      // ignore
    }
  },
}));
