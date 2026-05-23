import { useState } from "react";
import { Plug, Copy, Check, LoaderCircle } from "lucide-react";
import { Input } from "./ui/input";
import { Button } from "./ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "./ui/card";
import { useMcpIntegrationStore } from "@/store/mcpIntegrationStore";

export default function McpEndpointCard() {
  const { status, endpoint, setEndpoint, connect } =
    useMcpIntegrationStore();
  const [copied, setCopied] = useState(false);
  const [connecting, setConnecting] = useState(false);

  const handleCopy = async () => {
    if (!endpoint) return;
    await navigator.clipboard.writeText(endpoint);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleConnect = async () => {
    if (!endpoint) return;
    setConnecting(true);
    await connect();
    setConnecting(false);
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <Plug className="h-4 w-4 text-muted-foreground" />
          MCP Endpoint
        </CardTitle>
        <CardDescription>
          Pega la URL de Custom Services de XiaoZhi para iniciar la integracion
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="flex gap-2">
          <Input
            value={endpoint}
            onChange={(e) => setEndpoint(e.target.value)}
            placeholder="wss://api.xiaozhi.me/mcp/?token=..."
            disabled={status === "connected"}
            className="font-mono text-xs"
          />
          <Button
            variant="outline"
            size="icon"
            onClick={handleCopy}
            disabled={!endpoint}
          >
            {copied ? (
              <Check className="h-4 w-4 text-emerald-400" />
            ) : (
              <Copy className="h-4 w-4" />
            )}
          </Button>
        </div>
        <Button
          onClick={handleConnect}
          disabled={!endpoint || connecting || status === "connected"}
          className="w-full"
        >
          {connecting ? (
            <>
              <LoaderCircle className="h-4 w-4 animate-spin" />
              Conectando...
            </>
          ) : status === "connected" ? (
            "Conectado"
          ) : (
            "Conectar servidor MCP local"
          )}
        </Button>
      </CardContent>
    </Card>
  );
}
