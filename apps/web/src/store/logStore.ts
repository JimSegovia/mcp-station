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

export interface LogQuery {
  type?: string;
  source?: string;
  result?: string;
  limit?: number;
}

interface LogStore {
  logs: LogEntry[];

  clearLogs: () => Promise<void>;
  load: (query?: LogQuery) => Promise<void>;
}

export const useLogStore = create<LogStore>((set) => ({
  logs: [],

  load: async (query?: LogQuery) => {
    try {
      const params = new URLSearchParams();
      if (query?.type) params.set("type", query.type);
      if (query?.source) params.set("source", query.source);
      if (query?.result) params.set("result", query.result);
      if (query?.limit) params.set("limit", String(query.limit));
      const qs = params.toString();
      const res = await fetch(`/api/logs${qs ? `?${qs}` : ""}`);
      if (!res.ok) return;
      const data: LogEntry[] = await res.json();
      set({ logs: data });
    } catch {
      // backend not available
    }
  },

  clearLogs: async () => {
    try {
      await fetch("/api/logs", { method: "DELETE" });
      set({ logs: [] });
    } catch {
      // ignore
    }
  },
}));
