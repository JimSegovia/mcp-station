import { Plug, Server, ScrollText, Activity, Timer } from "lucide-react";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "./ui/card";
import { Badge } from "./ui/badge";
import { useMcpIntegrationStore } from "@/store/mcpIntegrationStore";
import { useMcpServerStore } from "@/store/mcpServerStore";
import { useLogStore } from "@/store/logStore";

function formatUptime(seconds: number) {
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  if (h > 0) return `${h}h ${m}m`;
  if (m > 0) return `${m}m`;
  return `${seconds}s`;
}

export default function DashboardSummary() {
  const integration = useMcpIntegrationStore();
  const { servers } = useMcpServerStore();
  const { logs } = useLogStore();

  const activeServers = servers.filter((s) => s.enabled).length;
  const connectedServers = servers.filter((s) => s.status === "connected").length;
  const errorLogs = logs.filter((l) => l.result === "error").length;
  const healthyChecks = integration.healthChecks.filter((h) => h.level === "healthy").length;
  const totalChecks = integration.healthChecks.length || 3;

  const items = [
    {
      icon: Plug,
      label: "Integracion MCP",
      state: integration.status,
      extra:
        integration.status === "connected"
          ? "Conectado"
          : integration.status === "error"
          ? "Error"
          : "Sin configurar",
      badgeVariant:
        integration.status === "connected"
          ? ("success" as const)
          : integration.status === "error"
          ? ("destructive" as const)
          : ("secondary" as const),
    },
    {
      icon: Activity,
      label: "Health checks",
      state: `${healthyChecks}/${totalChecks} ok`,
      extra:
        healthyChecks === totalChecks
          ? "Todo saludable"
          : healthyChecks === 0
          ? "Sin datos"
          : "Degradado",
      badgeVariant:
        healthyChecks === totalChecks && totalChecks > 0
          ? ("success" as const)
          : healthyChecks > 0
          ? ("warning" as const)
          : ("secondary" as const),
    },
    {
      icon: Timer,
      label: "Uptime servidor MCP",
      state: integration.uptime > 0
        ? formatUptime(integration.uptime)
        : "Offline",
      extra: integration.latency > 0
        ? `Latencia ${integration.latency}ms`
        : "Sin datos",
      badgeVariant: integration.uptime > 0 ? ("success" as const) : ("secondary" as const),
    },
    {
      icon: Server,
      label: "Servidores MCP",
      state: `${connectedServers}/${activeServers} conectados`,
      extra: `${servers.length} registrados`,
      badgeVariant: connectedServers > 0 ? ("success" as const) : ("secondary" as const),
    },
    {
      icon: ScrollText,
      label: "Logs",
      state: `${logs.length} eventos`,
      extra: errorLogs > 0 ? `${errorLogs} errores` : "Sin errores",
      badgeVariant: errorLogs === 0 ? ("success" as const) : ("warning" as const),
    },
  ];

  return (
    <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
      {items.map((item) => (
        <Card key={item.label}>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              {item.label}
            </CardTitle>
            <item.icon className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-xl font-bold">{item.state}</div>
            <div className="flex items-center gap-2 mt-1">
              <Badge variant={item.badgeVariant} className="text-[10px]">
                {item.extra}
              </Badge>
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}
