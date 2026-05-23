import { create } from "zustand";

export interface OpenCodeSession {
  sessionId: string;
  trackId: string;
  toolName: string;
  prompt: string;
  title: string;
  status: string;
  createdAt: string;
  age: string;
}

interface OpenCodeSessionStore {
  sessions: OpenCodeSession[];
  loading: boolean;
  load: () => Promise<void>;
  deleteSession: (sessionId: string, trackId: string) => Promise<void>;
  cleanExpired: () => Promise<void>;
}

export const useOpenCodeSessionStore = create<OpenCodeSessionStore>((set, get) => ({
  sessions: [],
  loading: false,

  load: async () => {
    set({ loading: true });
    try {
      const res = await fetch("/api/opencode/sessions");
      if (!res.ok) {
        set({ sessions: [], loading: false });
        return;
      }
      const data: OpenCodeSession[] = await res.json();
      set({ sessions: data, loading: false });
    } catch {
      set({ sessions: [], loading: false });
    }
  },

  deleteSession: async (sessionId: string, trackId: string) => {
    try {
      await fetch(`/api/opencode/sessions/${encodeURIComponent(sessionId)}`, {
        method: "DELETE",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ trackId }),
      });
      set({ sessions: get().sessions.filter((s) => s.sessionId !== sessionId) });
    } catch {
      // ignore
    }
  },

  cleanExpired: async () => {
    try {
      await fetch("/api/opencode/sessions", { method: "DELETE" });
      get().load();
    } catch {
      // ignore
    }
  },
}));
