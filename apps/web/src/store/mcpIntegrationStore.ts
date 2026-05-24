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
  startLiveUpdates: () => Promise<void>;
  stopLiveUpdates: () => void;
}

const STREAM_RETRY_MS = 3000;
const FALLBACK_POLL_MS = 10000;

let integrationStream: EventSource | null = null;
let reconnectTimer: number | null = null;
let fallbackPollTimer: number | null = null;
let liveConsumers = 0;
let manualStop = false;

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

function applyIntegrationData(
  set: (partial: Partial<McpIntegrationStore>) => void,
  data: IntegrationApiResponse
) {
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
}

function clearReconnectTimer() {
  if (reconnectTimer !== null) {
    window.clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
}

function stopFallbackPolling() {
  if (fallbackPollTimer !== null) {
    window.clearInterval(fallbackPollTimer);
    fallbackPollTimer = null;
  }
}

function closeStream() {
  if (integrationStream) {
    integrationStream.close();
    integrationStream = null;
  }
}

function startFallbackPolling(get: () => McpIntegrationStore) {
  if (fallbackPollTimer !== null) return;
  fallbackPollTimer = window.setInterval(() => {
    void get().load();
  }, FALLBACK_POLL_MS);
}

function scheduleReconnect(
  set: (partial: Partial<McpIntegrationStore>) => void,
  get: () => McpIntegrationStore
) {
  if (manualStop || reconnectTimer !== null || liveConsumers === 0) {
    return;
  }

  reconnectTimer = window.setTimeout(() => {
    reconnectTimer = null;
    if (!manualStop && liveConsumers > 0) {
      openIntegrationStream(set, get);
    }
  }, STREAM_RETRY_MS);
}

function openIntegrationStream(
  set: (partial: Partial<McpIntegrationStore>) => void,
  get: () => McpIntegrationStore
) {
  if (manualStop || integrationStream || typeof EventSource === "undefined") {
    if (typeof EventSource === "undefined") {
      startFallbackPolling(get);
    }
    return;
  }

  const stream = new EventSource("/api/integration/stream");
  integrationStream = stream;

  stream.onopen = () => {
    clearReconnectTimer();
    stopFallbackPolling();
  };

  stream.addEventListener("integration", (event) => {
    try {
      const payload = JSON.parse((event as MessageEvent<string>).data) as IntegrationApiResponse;
      applyIntegrationData(set, payload);
    } catch {
      // ignore malformed events
    }
  });

  stream.onerror = () => {
    closeStream();
    startFallbackPolling(get);
    void get().load();
    scheduleReconnect(set, get);
  };
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
      applyIntegrationData(set, data);
    } catch {
      // backend not available
    }
  },

  startLiveUpdates: async () => {
    liveConsumers++;
    manualStop = false;

    if (liveConsumers > 1) {
      return;
    }

    await get().load();
    openIntegrationStream(set, get);
    if (typeof EventSource === "undefined") {
      startFallbackPolling(get);
    }
  },

  stopLiveUpdates: () => {
    liveConsumers = Math.max(0, liveConsumers - 1);
    if (liveConsumers > 0) {
      return;
    }

    manualStop = true;
    clearReconnectTimer();
    stopFallbackPolling();
    closeStream();
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
