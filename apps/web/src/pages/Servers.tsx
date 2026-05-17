import { useState } from "react";
import { Plus } from "lucide-react";
import { useMcpServerStore } from "@/store/mcpServerStore";
import McpServerCard from "@/components/McpServerCard";
import McpServerForm from "@/components/McpServerForm";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";

export default function Servers() {
  const { servers, addServer, removeServer, toggleServer, toggleTool } =
    useMcpServerStore();
  const [formOpen, setFormOpen] = useState(false);

  const connectedCount = servers.filter((s) => s.status === "connected").length;

  return (
    <div className="max-w-2xl space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Servidores MCP</h1>
          <p className="text-sm text-muted-foreground mt-1">
            {servers.length} servidores registrados · {connectedCount} conectados
          </p>
        </div>
        <Button onClick={() => setFormOpen(true)}>
          <Plus className="h-4 w-4 mr-1" />
          Agregar
        </Button>
      </div>

      <McpServerForm
        open={formOpen}
        onClose={() => setFormOpen(false)}
        onAdd={addServer}
      />

      {servers.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-center text-muted-foreground">
          <p className="text-sm">No hay servidores MCP registrados.</p>
          <p className="text-xs mt-1">
            Usa el boton Agregar para registrar tu primer servidor.
          </p>
        </div>
      ) : (
        <div className="space-y-3">
          {servers.map((server) => (
            <McpServerCard
              key={server.id}
              server={server}
              onToggle={() => toggleServer(server.id)}
              onRemove={() => removeServer(server.id)}
              onToggleTool={(toolName) => toggleTool(server.id, toolName)}
            />
          ))}
        </div>
      )}

      <div className="flex items-center gap-3">
        <Separator className="flex-1" />
        <span className="text-xs text-muted-foreground">demo</span>
        <Separator className="flex-1" />
      </div>
      <p className="text-xs text-muted-foreground text-center">
        Servidores mock para validar la UI. El backend Go sincronizara el
        estado real de cada servidor MCP.
      </p>
    </div>
  );
}
