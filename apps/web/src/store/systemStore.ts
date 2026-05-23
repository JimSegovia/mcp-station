import { create } from "zustand";

export interface MemoryStats {
  alloc: number;
  totalAlloc: number;
  sys: number;
  heapAlloc: number;
  heapSys: number;
  numGC: number;
}

export interface SystemStats {
  pid: number;
  uptime: number;
  goroutines: number;
  numCPU: number;
  goVersion: string;
  memory: MemoryStats;
}

interface SystemStore {
  stats: SystemStats | null;
  load: () => Promise<void>;
}

export const useSystemStore = create<SystemStore>((set) => ({
  stats: null,

  load: async () => {
    try {
      const res = await fetch("/api/system");
      if (!res.ok) return;
      const data: SystemStats = await res.json();
      set({ stats: data });
    } catch {
      // backend not available
    }
  },
}));
