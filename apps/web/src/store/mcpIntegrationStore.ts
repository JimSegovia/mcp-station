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

export interface ToolTestResult {
  toolName: string;
  params: string;
  result: string;
  duration: number;
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
  toolTestResults: ToolTestResult[];

  setEndpoint: (url: string) => void;
  simulateConnect: () => void;
  disconnect: () => void;
  toggleTool: (toolName: string) => void;
  simulateError: (message: string) => void;
  reset: () => void;
  runToolTest: (toolName: string, params: string) => void;
  tickUptime: () => void;
  simulateHealthDegrade: () => void;
  restoreHealth: () => void;
}

const mockTools: McpTool[] = [
  { name: "open_application", description: "Abre una aplicacion del sistema", enabled: true },
  { name: "create_file", description: "Crea un archivo en una ruta permitida", enabled: true },
  { name: "list_directory", description: "Lista el contenido de un directorio", enabled: true },
  { name: "execute_command", description: "Ejecuta un comando controlado en la terminal", enabled: false },
  { name: "read_file", description: "Lee el contenido de un archivo de texto", enabled: true },
  { name: "opencode_task", description: "Envia una tarea a OpenCode para ejecucion inteligente", enabled: true },
];

const defaultHealth: HealthCheck[] = [
  { label: "Endpoint reachable", level: "healthy", detail: "200 OK · wss://api.xiaozhi.me/mcp" },
  { label: "Server runtime", level: "healthy", detail: "PID 48291 · uptime 2h 34m" },
  { label: "Tool registry", level: "healthy", detail: "6 tools loaded · 0 stale" },
  { label: "Message queue", level: "healthy", detail: "0 pending · avg 3ms" },
];

export const useMcpIntegrationStore = create<McpIntegrationStore>((set) => ({
  status: "empty",
  endpoint: "",
  tools: [],
  lastConnected: null,
  lastError: null,

  serverPort: 8090,
  protocolVersion: "MCP/JSON-RPC 2.0",
  uptime: 9240,
  latency: 12,
  healthChecks: [],
  toolTestResults: [],

  setEndpoint: (url: string) =>
    set({ endpoint: url }),

  simulateConnect: () =>
    set({
      status: "connected",
      tools: mockTools,
      lastConnected: new Date().toISOString(),
      lastError: null,
      healthChecks: defaultHealth,
      uptime: 0,
      latency: 12,
    }),

  disconnect: () =>
    set((state) => ({
      status: "empty",
      endpoint: state.endpoint,
      tools: [],
      lastConnected: null,
      healthChecks: [],
      toolTestResults: [],
    })),

  toggleTool: (toolName: string) =>
    set((state) => ({
      tools: state.tools.map((t) =>
        t.name === toolName ? { ...t, enabled: !t.enabled } : t
      ),
    })),

  simulateError: (message: string) =>
    set((state) => ({
      status: "error",
      tools: state.tools.length > 0 ? state.tools : mockTools,
      healthChecks: [
        { label: "Endpoint reachable", level: "unhealthy", detail: "Connection refused" },
        { label: "Server runtime", level: "degraded", detail: "PID 48291 · retrying" },
        { label: "Tool registry", level: "degraded", detail: "4/6 tools validated" },
        { label: "Message queue", level: "degraded", detail: "3 pending · timeout 30s" },
      ],
      lastError: {
        message,
        timestamp: new Date().toISOString(),
      },
    })),

  reset: () =>
    set({
      status: "empty",
      endpoint: "",
      tools: [],
      lastConnected: null,
      lastError: null,
      healthChecks: [],
      toolTestResults: [],
      uptime: 0,
    }),

  runToolTest: (toolName: string, params: string) => {
    const duration = Math.floor(Math.random() * 200) + 20;
    const results: Record<string, string> = {
      open_application: "VS Code abierto correctamente en la pantalla principal",
      create_file: "Archivo 'notas.txt' creado en ~/Documents/mcp-station/",
      list_directory: "12 archivos · 3 carpetas · 2.4 MB en total",
      execute_command: "bash: comando no permitido por politica de seguridad",
      read_file: "[OK] 142 lineas leidas · encoding: UTF-8",
      opencode_task: "Tarea delegada a OpenCode · en progreso",
    };
    set((state) => ({
      toolTestResults: [
        {
          toolName,
          params,
          result: results[toolName] ?? `Tool ${toolName} ejecutada sin novedades`,
          duration,
          timestamp: new Date().toISOString(),
        },
        ...state.toolTestResults.slice(0, 9),
      ],
    }));
  },

  tickUptime: () =>
    set((state) => ({
      uptime: state.status === "connected" ? state.uptime + 1 : state.uptime,
    })),

  simulateHealthDegrade: () =>
    set((state) => ({
      healthChecks: state.healthChecks.map((h, i) =>
        i === Math.floor(Math.random() * state.healthChecks.length)
          ? { ...h, level: "degraded" as HealthLevel, detail: h.detail + " · warning" }
          : h
      ),
    })),

  restoreHealth: () =>
    set((state) => ({
      healthChecks: state.healthChecks.map((h) =>
        h.level !== "healthy" ? { ...h, level: "healthy" as HealthLevel, detail: h.detail.replace(" · warning", "") } : h
      ),
    })),
}));
