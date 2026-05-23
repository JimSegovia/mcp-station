import { useEffect } from "react";
import { Server } from "lucide-react";
import { useMcpServerStore } from "@/store/mcpServerStore";
import McpServerCard from "@/components/McpServerCard";
import McpServerForm from "@/components/McpServerForm";
import {
  Card,
  CardContent,
} from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";

export default function Servers() {
  const { servers, load, toggleServer, toggleTool, removeServer, discoverTools } =
    useMcpServerStore();

  useEffect(() => {
    load();
  }, [load]);

  const virtualServers = servers.filter((s) => s.type === "virtual");
  const externalServers = servers.filter((s) => s.type !== "virtual");
  const connectedCount = servers.filter((s) => s.status === "connected").length;

  return (
    <div className="max-w-2xl space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Servidores MCP</h1>
          <p className="text-sm text-muted-foreground mt-1">
            {servers.length} servidores · {connectedCount} conectados
            {virtualServers.length > 0 && ` · ${virtualServers.length} locales`}
          </p>
        </div>
        <McpServerForm />
      </div>

      {externalServers.length === 0 && (
        <Card>
          <CardContent className="py-12 text-center">
            <Server className="h-8 w-8 text-muted-foreground mx-auto mb-3" />
            <p className="text-sm text-muted-foreground">
              No hay servidores externos registrados
            </p>
            <p className="text-xs text-muted-foreground mt-1">
              Agrega un servidor MCP externo para descubrir sus tools
            </p>
          </CardContent>
        </Card>
      )}

      {externalServers.map((s) => (
        <McpServerCard
          key={s.id}
          server={s}
          onToggle={() => toggleServer(s.id)}
          onToggleTool={(toolName) => toggleTool(s.id, toolName)}
          onRemove={() => removeServer(s.id)}
          onDiscover={
            s.type === "websocket"
              ? () => discoverTools(s.id)
              : undefined
          }
        />
      ))}

      <div className="flex items-center gap-3">
        <Separator className="flex-1" />
        <span className="text-xs text-muted-foreground">servidores locales</span>
        <Separator className="flex-1" />
      </div>

      {virtualServers.map((s) => (
        <McpServerCard
          key={s.id}
          server={s}
          onToggle={() => toggleServer(s.id)}
          onToggleTool={(toolName) => toggleTool(s.id, toolName)}
          isVirtual
        />
      ))}
    </div>
  );
}
