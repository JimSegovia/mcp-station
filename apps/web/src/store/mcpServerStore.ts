import { create } from "zustand";

export type McpServerType = "stdio" | "websocket";
export type McpServerStatus = "connected" | "disconnected" | "error";

export interface McpServerTool {
  name: string;
  description: string;
  enabled: boolean;
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
}

interface McpServerStore {
  servers: McpServer[];

  addServer: (server: Omit<McpServer, "id" | "status" | "tools" | "lastConnected">) => void;
  removeServer: (id: string) => void;
  toggleServer: (id: string) => void;
  toggleTool: (serverId: string, toolName: string) => void;
  setServerStatus: (id: string, status: McpServerStatus) => void;
}

const defaultServers: McpServer[] = [
  {
    id: "mcp-1",
    name: "OpenCode Local",
    type: "stdio",
    endpoint: "opencode serve",
    enabled: true,
    status: "connected",
    tools: [
      { name: "bash", description: "Ejecuta comandos en la terminal", enabled: true },
      { name: "read", description: "Lee archivos del sistema", enabled: true },
      { name: "edit", description: "Edita archivos de texto", enabled: true },
    ],
    lastConnected: new Date().toISOString(),
  },
  {
    id: "mcp-2",
    name: "File System Proxy",
    type: "websocket",
    endpoint: "ws://localhost:8081/mcp",
    enabled: false,
    status: "disconnected",
    tools: [
      { name: "list_files", description: "Lista archivos en un directorio", enabled: true },
      { name: "move_file", description: "Mueve archivos entre directorios", enabled: false },
    ],
    lastConnected: null,
  },
];

let nextId = 3;

export const useMcpServerStore = create<McpServerStore>((set) => ({
  servers: defaultServers,

  addServer: (server) =>
    set((state) => ({
      servers: [
        ...state.servers,
        {
          ...server,
          id: `mcp-${nextId++}`,
          status: "disconnected" as McpServerStatus,
          tools: [],
          lastConnected: null,
        },
      ],
    })),

  removeServer: (id: string) =>
    set((state) => ({
      servers: state.servers.filter((s) => s.id !== id),
    })),

  toggleServer: (id: string) =>
    set((state) => ({
      servers: state.servers.map((s) =>
        s.id === id
          ? {
              ...s,
              enabled: !s.enabled,
              status: s.enabled ? ("disconnected" as McpServerStatus) : ("connected" as McpServerStatus),
            }
          : s
      ),
    })),

  toggleTool: (serverId: string, toolName: string) =>
    set((state) => ({
      servers: state.servers.map((s) =>
        s.id === serverId
          ? {
              ...s,
              tools: s.tools.map((t) =>
                t.name === toolName ? { ...t, enabled: !t.enabled } : t
              ),
            }
          : s
      ),
    })),

  setServerStatus: (id: string, status: McpServerStatus) =>
    set((state) => ({
      servers: state.servers.map((s) =>
        s.id === id ? { ...s, status, lastConnected: status === "connected" ? new Date().toISOString() : s.lastConnected } : s
      ),
    })),
}));
