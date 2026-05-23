import { useState, useEffect, useCallback } from "react";
import { Play, Beaker, Clock, LoaderCircle } from "lucide-react";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "./ui/card";
import { Badge } from "./ui/badge";
import { Separator } from "./ui/separator";

interface ToolDef {
  name: string;
  description: string;
  origin: string;
  enabled: boolean;
}

interface ToolGroup {
  origin: string;
  tools: ToolDef[];
}

interface ToolTestResult {
  toolName: string;
  params: string;
  result: string;
  duration: number;
  timestamp: string;
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleString("es-CL", {
    timeStyle: "medium",
  });
}

export default function ToolTestPanel() {
  const [groups, setGroups] = useState<ToolGroup[]>([]);
  const [selectedTool, setSelectedTool] = useState("");
  const [params, setParams] = useState("");
  const [results, setResults] = useState<ToolTestResult[]>([]);
  const [running, setRunning] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch("/api/tools")
      .then((r) => {
        if (!r.ok) throw new Error("Failed to load tools");
        return r.json();
      })
      .then((data: ToolGroup[]) => setGroups(data))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const allTools = groups.flatMap((g) =>
    g.tools.filter((t) => t.enabled).map((t) => ({ ...t, origin: g.origin }))
  );

  const handleRun = useCallback(() => {
    if (!selectedTool || running) return;

    const startTime = performance.now();
    setRunning(true);

    const protocol = location.protocol === "https:" ? "wss" : "ws";
    const ws = new WebSocket(`${protocol}://${location.host}/ws`);
    const initId = 1;
    const callId = 2;

    ws.onopen = () => {
      ws.send(
        JSON.stringify({
          jsonrpc: "2.0",
          id: initId,
          method: "initialize",
          params: {},
        })
      );
    };

    ws.onmessage = (event) => {
      const msg = JSON.parse(event.data);

      if (msg.id === initId && msg.result?.protocolVersion) {
        ws.send(
          JSON.stringify({
            jsonrpc: "2.0",
            method: "notifications/initialized",
          })
        );

        let parsedParams: Record<string, unknown> = {};
        try {
          parsedParams = JSON.parse(params || "{}");
        } catch {
          parsedParams = {};
        }

        ws.send(
          JSON.stringify({
            jsonrpc: "2.0",
            id: callId,
            method: "tools/call",
            params: { name: selectedTool, arguments: parsedParams },
          })
        );
        return;
      }

      if (msg.id === callId) {
        const duration = Math.round(performance.now() - startTime);
        let resultText = "";
        if (msg.result?.content?.[0]?.text) {
          resultText = msg.result.content[0].text;
        } else if (msg.error) {
          resultText = `Error: ${msg.error.message}`;
        } else {
          resultText = JSON.stringify(msg.result);
        }

        setResults((prev) => [
          {
            toolName: selectedTool,
            params: params || "{}",
            result: resultText,
            duration,
            timestamp: new Date().toISOString(),
          },
          ...prev.slice(0, 9),
        ]);
        setRunning(false);
        ws.close();
      }
    };

    ws.onerror = () => {
      setResults((prev) => [
        {
          toolName: selectedTool,
          params: params || "{}",
          result: "Error: WebSocket connection failed",
          duration: Math.round(performance.now() - startTime),
          timestamp: new Date().toISOString(),
        },
        ...prev.slice(0, 9),
      ]);
      setRunning(false);
    };

    ws.onclose = () => setRunning(false);

    setTimeout(() => {
      if (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING) {
        ws.close();
        setRunning(false);
      }
    }, 30000);
  }, [selectedTool, params, running]);

  if (loading) return null;
  if (allTools.length === 0) return null;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <Beaker className="h-4 w-4 text-muted-foreground" />
          Probar tool
        </CardTitle>
        <CardDescription>
          Ejecuta cualquier tool habilitada via WebSocket MCP
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="flex gap-2">
          <select
            value={selectedTool}
            onChange={(e) => {
              setSelectedTool(e.target.value);
              setParams("");
            }}
            className="flex h-9 w-full rounded-md border border-border bg-background text-foreground px-3 py-1 text-sm shadow-sm font-mono text-xs"
          >
            <option value="" disabled>
              Selecciona una tool...
            </option>
            {groups.map((g) => (
              <optgroup key={g.origin} label={`-- ${g.origin}`}>
                {g.tools
                  .filter((t) => t.enabled)
                  .map((t) => (
                    <option key={t.name} value={t.name}>
                      {t.name}
                    </option>
                  ))}
              </optgroup>
            ))}
          </select>
          <Button
            onClick={handleRun}
            disabled={!selectedTool || running}
            size="sm"
          >
            {running ? (
              <LoaderCircle className="h-3.5 w-3.5 animate-spin" />
            ) : (
              <Play className="h-3.5 w-3.5 mr-1" />
            )}
            Ejecutar
          </Button>
        </div>

        {selectedTool && (
          <Input
            value={params}
            onChange={(e) => setParams(e.target.value)}
            placeholder='ej: {"path": "."}'
            className="font-mono text-xs"
          />
        )}

        {results.length > 0 && (
          <>
            <Separator />
            <p className="text-xs text-muted-foreground flex items-center gap-1 mb-2">
              <Clock className="h-3 w-3" />
              Resultados recientes
            </p>
            <div className="space-y-2 max-h-48 overflow-y-auto">
              {results.map((r, i) => (
                <div
                  key={i}
                  className="rounded-md border border-border bg-card px-3 py-2"
                >
                  <div className="flex items-center justify-between mb-1">
                    <div className="flex items-center gap-1.5">
                      <code className="text-xs font-mono text-blue-400">
                        {r.toolName}
                      </code>
                      <Badge variant="outline" className="text-[9px]">
                        {r.duration}ms
                      </Badge>
                    </div>
                    <span className="text-xs text-muted-foreground">
                      {formatDate(r.timestamp)}
                    </span>
                  </div>
                  {r.params !== "{}" && (
                    <p className="text-xs text-muted-foreground mb-0.5 font-mono">
                      params: {r.params}
                    </p>
                  )}
                  <p className="text-xs text-foreground whitespace-pre-wrap break-all leading-relaxed">
                    {r.result.length > 300
                      ? r.result.slice(0, 300) + "..."
                      : r.result}
                  </p>
                </div>
              ))}
            </div>
          </>
        )}
      </CardContent>
    </Card>
  );
}
