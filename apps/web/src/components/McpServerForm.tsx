import { useState } from "react";
import { Plus } from "lucide-react";
import { Dialog } from "./ui/dialog";
import { Input } from "./ui/input";
import { Button } from "./ui/button";
import { useMcpServerStore } from "@/store/mcpServerStore";

const serverTypes = [
  { value: "websocket" as const, label: "WebSocket" },
  { value: "stdio" as const, label: "STDIO" },
];

export default function McpServerForm() {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [type, setType] = useState<"websocket" | "stdio">("websocket");
  const [endpoint, setEndpoint] = useState("");
  const { addServer } = useMcpServerStore();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name || !endpoint) return;
    addServer({ name, type, endpoint });
    setName("");
    setEndpoint("");
    setType("websocket");
    setOpen(false);
  };

  return (
    <>
      <Button onClick={() => setOpen(true)}>
        <Plus className="h-4 w-4 mr-1" />
        Agregar
      </Button>
      <Dialog
        open={open}
        onClose={() => setOpen(false)}
        title="Registrar servidor MCP"
        description="Agrega un nuevo servidor MCP externo para descubrir sus tools"
      >
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="text-sm font-medium mb-1.5 block">Nombre</label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Ej: File System Proxy"
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
          <Button type="submit" className="w-full">
            Registrar servidor
          </Button>
        </form>
      </Dialog>
    </>
  );
}
