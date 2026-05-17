import { Server, Trash2, ChevronDown, ChevronRight } from "lucide-react";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "./ui/card";
import { Badge } from "./ui/badge";
import { Switch } from "./ui/switch";
import { Separator } from "./ui/separator";
import { Button } from "./ui/button";
import { useState } from "react";
import type { McpServer } from "@/store/mcpServerStore";

const statusBadge: Record<string, { label: string; variant: "success" | "destructive" | "secondary" }> = {
  connected: { label: "Conectado", variant: "success" },
  disconnected: { label: "Desconectado", variant: "secondary" },
  error: { label: "Error", variant: "destructive" },
};

function formatDate(iso: string | null) {
  if (!iso) return "Nunca";
  return new Date(iso).toLocaleString("es-CL", {
    dateStyle: "medium",
    timeStyle: "short",
  });
}

interface McpServerCardProps {
  server: McpServer;
  onToggle: () => void;
  onRemove: () => void;
  onToggleTool: (toolName: string) => void;
}

export default function McpServerCard({
  server,
  onToggle,
  onRemove,
  onToggleTool,
}: McpServerCardProps) {
  const [expanded, setExpanded] = useState(false);
  const badge = statusBadge[server.status];

  return (
    <Card>
      <CardHeader>
        <div className="flex items-start justify-between gap-3">
          <div className="flex items-center gap-2 min-w-0">
            <Server className="h-4 w-4 text-muted-foreground shrink-0" />
            <div>
              <CardTitle className="text-base flex items-center gap-2">
                {server.name}
                <Badge variant="outline" className="text-[10px]">
                  {server.type}
                </Badge>
              </CardTitle>
              <div className="flex items-center gap-2 mt-1">
                <Badge variant={badge.variant}>{badge.label}</Badge>
                <span className="text-xs text-muted-foreground">
                  Ult. conexion: {formatDate(server.lastConnected)}
                </span>
              </div>
            </div>
          </div>
          <Switch checked={server.enabled} onCheckedChange={onToggle} />
        </div>
      </CardHeader>
      <CardContent className="space-y-2">
        <code className="text-xs font-mono text-muted-foreground block truncate">
          {server.endpoint}
        </code>
        <Separator />
        <div className="flex items-center justify-between">
          <span className="text-xs text-muted-foreground">
            {server.tools.filter((t) => t.enabled).length}/{server.tools.length} tools activas
          </span>
          <div className="flex gap-2">
            {server.tools.length > 0 && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setExpanded(!expanded)}
                className="text-xs"
              >
                {expanded ? (
                  <ChevronDown className="h-3 w-3 mr-1" />
                ) : (
                  <ChevronRight className="h-3 w-3 mr-1" />
                )}
                Tools
              </Button>
            )}
            <Button
              variant="ghost"
              size="sm"
              onClick={onRemove}
              className="text-xs text-destructive hover:text-destructive"
            >
              <Trash2 className="h-3 w-3 mr-1" />
              Eliminar
            </Button>
          </div>
        </div>
        {expanded && server.tools.length > 0 && (
          <div className="space-y-1.5 pt-2">
            {server.tools.map((tool) => (
              <div key={tool.name} className="flex items-center justify-between gap-2 py-1">
                <div className="min-w-0">
                  <code className="text-xs font-mono text-foreground">{tool.name}</code>
                  <p className="text-xs text-muted-foreground truncate">{tool.description}</p>
                </div>
                <Switch
                  checked={tool.enabled}
                  onCheckedChange={() => onToggleTool(tool.name)}
                />
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
