import { useEffect } from "react";
import { useMcpIntegrationStore } from "@/store/mcpIntegrationStore";
import { useMcpServerStore } from "@/store/mcpServerStore";
import { useLogStore } from "@/store/logStore";
import DashboardSummary from "@/components/DashboardSummary";
import { Separator } from "@/components/ui/separator";

export default function Dashboard() {
  const loadIntegration = useMcpIntegrationStore((s) => s.load);
  const loadServers = useMcpServerStore((s) => s.load);
  const loadLogs = useLogStore((s) => s.load);

  useEffect(() => {
    loadIntegration();
    loadServers();
    loadLogs();
  }, [loadIntegration, loadServers, loadLogs]);

  return (
    <div className="max-w-5xl space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Resumen</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Estado general del sistema MCP Station
        </p>
      </div>

      <DashboardSummary />

      <Separator />

      <p className="text-xs text-muted-foreground text-center">
        Datos sincronizados en tiempo real desde el backend Go via API REST.
      </p>
    </div>
  );
}
