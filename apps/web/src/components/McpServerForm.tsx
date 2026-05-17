import { useState } from "react";
import { Dialog } from "./ui/dialog";
import { Input } from "./ui/input";
import { Button } from "./ui/button";
import type { McpServerType } from "@/store/mcpServerStore";

interface McpServerFormProps {
  open: boolean;
  onClose: () => void;
  onAdd: (server: { name: string; type: McpServerType; endpoint: string; enabled: boolean }) => void;
}

const serverTypes: { value: McpServerType; label: string }[] = [
  { value: "stdio", label: "STDIO" },
  { value: "websocket", label: "WebSocket" },
];

export default function McpServerForm({ open, onClose, onAdd }: McpServerFormProps) {
  const [name, setName] = useState("");
  const [type, setType] = useState<McpServerType>("websocket");
  const [endpoint, setEndpoint] = useState("");
  const [enabled, setEnabled] = useState(true);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name || !endpoint) return;
    onAdd({ name, type, endpoint, enabled });
    setName("");
    setEndpoint("");
    setType("websocket");
    setEnabled(true);
    onClose();
  };

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title="Registrar servidor MCP"
      description="Agrega un nuevo servidor MCP para expandir las capacidades del agente"
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="text-sm font-medium mb-1.5 block">Nombre</label>
          <Input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Ej: OpenCode Local"
            required
          />
        </div>
        <div>
          <label className="text-sm font-medium mb-1.5 block">Tipo</label>
          <div className="flex gap-2">
            {serverTypes.map((st) => (
              <button
                key={st.value}
                type="button"
                onClick={() => setType(st.value)}
                className={`px-3 py-1.5 rounded-md text-sm border transition-colors ${
                  type === st.value
                    ? "border-primary bg-primary/10 text-primary"
                    : "border-border text-muted-foreground hover:bg-accent"
                }`}
              >
                {st.label}
              </button>
            ))}
          </div>
        </div>
        <div>
          <label className="text-sm font-medium mb-1.5 block">
            {type === "stdio" ? "Comando" : "URL"}
          </label>
          <Input
            value={endpoint}
            onChange={(e) => setEndpoint(e.target.value)}
            placeholder={
              type === "stdio"
                ? "ej: opencode serve"
                : "ej: ws://localhost:8081/mcp"
            }
            className="font-mono text-xs"
            required
          />
        </div>
        <div className="flex items-center justify-between">
          <div>
            <span className="text-sm font-medium">Activar al registrar</span>
            <p className="text-xs text-muted-foreground mt-0.5">
              El servidor quedara habilitado inmediatamente
            </p>
          </div>
          <button
            type="button"
            role="switch"
            aria-checked={enabled}
            onClick={() => setEnabled(!enabled)}
            className={`inline-flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full border-2 border-transparent transition-colors ${
              enabled ? "bg-primary" : "bg-muted"
            }`}
          >
            <span
              className={`pointer-events-none block h-4 w-4 rounded-full bg-background shadow-sm ring-0 transition-transform ${
                enabled ? "translate-x-4" : "translate-x-0"
              }`}
            />
          </button>
        </div>
        <Button type="submit" className="w-full">
          Registrar servidor
        </Button>
      </form>
    </Dialog>
  );
}
