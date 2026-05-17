import { create } from "zustand";

export type LogType = "tool_call" | "connection" | "module" | "error" | "security";
export type LogResult = "success" | "error" | "blocked";

export interface LogEntry {
  id: string;
  timestamp: string;
  type: LogType;
  source: string;
  message: string;
  result: LogResult;
}

interface LogStore {
  logs: LogEntry[];

  addLog: (log: Omit<LogEntry, "id" | "timestamp">) => void;
  clearLogs: () => void;
}

let logId = 1;

const defaultLogs: LogEntry[] = [
  {
    id: "log-4",
    timestamp: new Date(Date.now() - 60000).toISOString(),
    type: "tool_call",
    source: "OpenCode Adapter",
    message: "Tarea completada: crear archivo de notas",
    result: "success",
  },
  {
    id: "log-3",
    timestamp: new Date(Date.now() - 300000).toISOString(),
    type: "connection",
    source: "MCP Integration",
    message: "Conexion establecida con Custom Services endpoint",
    result: "success",
  },
  {
    id: "log-2",
    timestamp: new Date(Date.now() - 600000).toISOString(),
    type: "security",
    source: "Policy Engine",
    message: "Accion bloqueada: execute_command fuera de lista blanca",
    result: "blocked",
  },
  {
    id: "log-1",
    timestamp: new Date(Date.now() - 900000).toISOString(),
    type: "error",
    source: "Terminal Runner",
    message: "Error al ejecutar comando: timeout despues de 30s",
    result: "error",
  },
];

export const useLogStore = create<LogStore>((set) => ({
  logs: defaultLogs,

  addLog: (log) =>
    set((state) => ({
      logs: [
        {
          ...log,
          id: `log-${logId++}`,
          timestamp: new Date().toISOString(),
        },
        ...state.logs,
      ],
    })),

  clearLogs: () => set({ logs: [] }),
}));
