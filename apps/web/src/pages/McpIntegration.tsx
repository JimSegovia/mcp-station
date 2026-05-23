import { useEffect } from "react";
import { LoaderCircle } from "lucide-react";
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
    load,
    disconnect,
    connect,
  } = useMcpIntegrationStore();

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    if (status !== "connecting" && status !== "connected") return;
    const interval = setInterval(load, 1500);
    return () => clearInterval(interval);
  }, [status, load]);

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
        </>
      )}

      {status === "connecting" && (
        <>
          <McpEndpointCard />
          <div className="flex flex-col items-center justify-center py-12 text-center space-y-3">
            <LoaderCircle className="h-8 w-8 animate-spin text-blue-400" />
            <p className="text-sm text-muted-foreground">
              Conectando con el endpoint de XiaoZhi...
            </p>
            <p className="text-xs text-muted-foreground">
              Esto puede tomar unos segundos mientras se realiza el handshake MCP
            </p>
          </div>
        </>
      )}

      {status === "connected" && (
        <>
          <ConnectionStatusCard />
          <McpServerRuntimeCard />
          <ToolsListCard />
          <ToolTestPanel />
          <McpEndpointCard />

          <div className="flex gap-3">
            <Button variant="outline" onClick={disconnect}>
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
            <span className="text-xs text-muted-foreground">debug</span>
            <Separator className="flex-1" />
          </div>
          <div className="flex gap-3">
            <Button className="flex-1" onClick={connect}>
              Reintentar conexion
            </Button>
            <Button variant="outline" className="flex-1" onClick={disconnect}>
              Reiniciar
            </Button>
          </div>
        </>
      )}
    </div>
  );
}
