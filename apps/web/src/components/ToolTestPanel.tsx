import { useState } from "react";
import { Play, Beaker, Clock } from "lucide-react";
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
import { useMcpIntegrationStore } from "@/store/mcpIntegrationStore";

function formatDate(iso: string) {
  return new Date(iso).toLocaleString("es-CL", {
    timeStyle: "medium",
  });
}

export default function ToolTestPanel() {
  const { tools, runToolTest, toolTestResults } = useMcpIntegrationStore();
  const [selectedTool, setSelectedTool] = useState("");
  const [params, setParams] = useState("");

  const activeTools = tools.filter((t) => t.enabled);

  const handleRun = () => {
    if (!selectedTool) return;
    runToolTest(selectedTool, params);
  };

  if (activeTools.length === 0) return null;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <Beaker className="h-4 w-4 text-muted-foreground" />
          Probar tool
        </CardTitle>
        <CardDescription>
          Ejecuta una tool con parametros de prueba para validar su funcionamiento
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="flex gap-2">
          <div className="relative flex-1">
            <select
              value={selectedTool}
              onChange={(e) => {
                setSelectedTool(e.target.value);
                setParams("");
              }}
              className="flex h-9 w-full rounded-md border border-border bg-transparent px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring font-mono text-xs"
            >
              <option value="" disabled>
                Selecciona una tool...
              </option>
              {activeTools.map((t) => (
                <option key={t.name} value={t.name}>
                  {t.name}
                </option>
              ))}
            </select>
          </div>
          <Button
            onClick={handleRun}
            disabled={!selectedTool}
            size="sm"
          >
            <Play className="h-3.5 w-3.5 mr-1" />
            Ejecutar
          </Button>
        </div>

        {selectedTool && (
          <div>
            <Input
              value={params}
              onChange={(e) => setParams(e.target.value)}
              placeholder={
                selectedTool === "open_application"
                  ? 'ej: {"app": "VS Code"}'
                  : selectedTool === "create_file"
                  ? 'ej: {"path": "/tmp/notas.txt", "content": "..."}'
                  : 'ej: {"path": "/home"}'
              }
              className="font-mono text-xs"
            />
          </div>
        )}

        {toolTestResults.length > 0 && (
          <>
            <Separator />
            <p className="text-xs text-muted-foreground flex items-center gap-1">
              <Clock className="h-3 w-3" />
              Resultados recientes
            </p>
            <div className="space-y-2 max-h-48 overflow-y-auto">
              {toolTestResults.map((r, i) => (
                <div
                  key={i}
                  className="rounded-md bg-muted/30 px-3 py-2 text-xs"
                >
                  <div className="flex items-center justify-between mb-1">
                    <div className="flex items-center gap-1.5">
                      <code className="font-mono text-foreground">{r.toolName}</code>
                      <Badge variant="outline" className="text-[9px]">
                        {r.duration}ms
                      </Badge>
                    </div>
                    <span className="text-muted-foreground">
                      {formatDate(r.timestamp)}
                    </span>
                  </div>
                  {r.params && (
                    <p className="text-muted-foreground mb-0.5 font-mono">
                      params: {r.params}
                    </p>
                  )}
                  <p className="text-emerald-400/80">{r.result}</p>
                </div>
              ))}
            </div>
          </>
        )}
      </CardContent>
    </Card>
  );
}
