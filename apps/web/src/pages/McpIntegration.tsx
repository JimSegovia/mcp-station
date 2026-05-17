import { useEffect } from "react";
import { useMcpIntegrationStore } from "@/store/mcpIntegrationStore";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import EmptyIntegrationState from "@/components/EmptyIntegrationState";
import CustomServicesGuideCard from "@/components/CustomServicesGuideCard";
import McpEndpointCard from "@/components/McpEndpointCard";
import ConnectionStatusCard from "@/components/ConnectionStatusCard";
import McpServerRuntimeCard from "@/components/McpServerRuntimeCard";
import ToolsListCard from "@/components/ToolsListCard";
import ToolTestPanel from "@/components/ToolTestPanel";
import RecentErrorCard from "@/components/RecentErrorCard";

export default function McpIntegration() {
  const {
    status,
    disconnect,
    simulateError,
    simulateHealthDegrade,
    restoreHealth,
    reset,
    tickUptime,
  } = useMcpIntegrationStore();

  useEffect(() => {
    if (status !== "connected") return;
    const interval = setInterval(tickUptime, 1000);
    return () => clearInterval(interval);
  }, [status, tickUptime]);

  return (
    <div className="max-w-2xl space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Integracion MCP</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Conecta el agente oficial de XiaoZhi a tu servidor MCP local via
          Custom Services
        </p>
      </div>

      {status === "empty" && (
        <>
          <EmptyIntegrationState />
          <CustomServicesGuideCard />
          <McpEndpointCard />

          <div className="flex items-center gap-3">
            <Separator className="flex-1" />
            <span className="text-xs text-muted-foreground">demo</span>
            <Separator className="flex-1" />
          </div>
          <p className="text-xs text-muted-foreground text-center">
            Pega un endpoint cualquiera y presiona Conectar para probar la UI.
            Los datos son mock locales.
          </p>
        </>
      )}

      {status === "connected" && (
        <>
          <ConnectionStatusCard />
          <McpServerRuntimeCard />
          <ToolsListCard />
          <ToolTestPanel />
          <McpEndpointCard />

          <div className="flex items-center gap-3">
            <Separator className="flex-1" />
            <span className="text-xs text-muted-foreground">demo</span>
            <Separator className="flex-1" />
          </div>
          <div className="flex gap-3 flex-wrap">
            <Button
              variant="destructive"
              onClick={() =>
                simulateError("Error de handshake MCP: el endpoint no respondio a initialize")
              }
            >
              Simular error
            </Button>
            <Button
              variant="secondary"
              onClick={simulateHealthDegrade}
            >
              Degradar health
            </Button>
            <Button
              variant="outline"
              onClick={restoreHealth}
            >
              Restaurar health
            </Button>
            <Button
              variant="ghost"
              onClick={disconnect}
            >
              Desconectar
            </Button>
          </div>
        </>
      )}

      {status === "error" && (
        <>
          <ConnectionStatusCard />
          <McpServerRuntimeCard />
          <RecentErrorCard />
          <ToolsListCard />

          <div className="flex items-center gap-3">
            <Separator className="flex-1" />
            <span className="text-xs text-muted-foreground">demo</span>
            <Separator className="flex-1" />
          </div>
          <div className="flex gap-3">
            <Button className="flex-1" onClick={disconnect}>
              Reconectar
            </Button>
            <Button variant="outline" className="flex-1" onClick={reset}>
              Reiniciar
            </Button>
          </div>
        </>
      )}
    </div>
  );
}
