import { useEffect, useState } from "react";
import { Terminal, RefreshCw, ChevronDown, Cpu } from "lucide-react";
import { Button } from "./ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "./ui/card";
import { Badge } from "./ui/badge";

interface ProviderInfo {
  id: string;
  name: string;
  models: Record<string, unknown>;
}

interface ModelsData {
  providers: ProviderInfo[];
  default: Record<string, string>;
}

export default function OpenCodeTerminal() {
  const [logs, setLogs] = useState<string[]>([]);
  const [ready, setReady] = useState(false);
  const [expanded, setExpanded] = useState(true);
  const [providers, setProviders] = useState<ProviderInfo[]>([]);

  const fetchLogs = async () => {
    try {
      const res = await fetch("/api/opencode/log");
      if (!res.ok) return;
      const data = await res.json();
      setLogs(data.logs ?? []);
      setReady(data.ready ?? false);
    } catch {
      // ignore
    }
  };

  const fetchModels = async () => {
    try {
      const res = await fetch("/api/opencode/models");
      if (!res.ok) return;
      const data: ModelsData = await res.json();
      setProviders(data.providers ?? []);
    } catch {
      // ignore
    }
  };

  useEffect(() => {
    fetchLogs();
    const interval = setInterval(fetchLogs, 3000);
    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    if (ready) fetchModels();
  }, [ready]);

  const totalModels = providers.reduce((n, p) => n + Object.keys(p.models).length, 0);

  return (
    <Card>
      <CardHeader
        className="cursor-pointer select-none"
        onClick={() => setExpanded(!expanded)}
      >
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Terminal className="h-4 w-4 text-muted-foreground" />
            <CardTitle className="text-base">Terminal OpenCode</CardTitle>
            <Badge
              variant={ready ? "success" : "secondary"}
              className="text-[10px]"
            >
              {ready ? "running" : "stopped"}
            </Badge>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7"
              onClick={(e) => {
                e.stopPropagation();
                fetchLogs();
                fetchModels();
              }}
            >
              <RefreshCw className="h-3.5 w-3.5" />
            </Button>
            <ChevronDown
              className={`h-4 w-4 text-muted-foreground transition-transform ${
                expanded ? "rotate-180" : ""
              }`}
            />
          </div>
        </div>
        <CardDescription>
          {ready
            ? `${providers.length} providers · ${totalModels} modelos · ${logs.length} lineas`
            : "Iniciando..."}
        </CardDescription>
      </CardHeader>
      {expanded && (
        <CardContent className="space-y-3">
          {ready && providers.length > 0 && (
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Cpu className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-xs font-medium">Modelos disponibles</span>
              </div>
              <div className="grid gap-2 sm:grid-cols-2">
                {providers
                  .filter((p) => Object.keys(p.models).length > 0)
                  .map((p) => {
                    const modelIds = Object.keys(p.models);
                    return (
                      <div
                        key={p.id}
                        className="rounded-md border border-border bg-card/50 px-3 py-2"
                      >
                        <span className="text-xs font-medium text-foreground">
                          {p.name}
                        </span>
                        <div className="flex flex-wrap gap-1 mt-1">
                          {modelIds.slice(0, 8).map((m) => (
                            <span
                              key={m}
                              className="text-[10px] px-1.5 py-0.5 rounded font-mono bg-muted text-muted-foreground"
                            >
                              {m}
                            </span>
                          ))}
                          {modelIds.length > 8 && (
                            <span className="text-[10px] text-muted-foreground">
                              +{modelIds.length - 8}
                            </span>
                          )}
                        </div>
                      </div>
                    );
                  })}
              </div>
            </div>
          )}

          <div className="bg-black/90 rounded-md border border-border overflow-hidden">
            <div className="max-h-80 overflow-y-auto p-3 font-mono text-xs leading-relaxed">
              {logs.length === 0 ? (
                <span className="text-muted-foreground">
                  Esperando salida del proceso...
                </span>
              ) : (
                logs.map((line, i) => (
                  <div
                    key={i}
                    className={
                      line.toLowerCase().includes("error") || line.toLowerCase().includes("warn")
                        ? "text-red-400"
                        : line.startsWith(">")
                        ? "text-cyan-400"
                        : "text-green-400/80"
                    }
                  >
                    {line}
                  </div>
                ))
              )}
            </div>
          </div>
        </CardContent>
      )}
    </Card>
  );
}
