import { Cpu, Timer, Activity } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "./ui/card";
import { Badge } from "./ui/badge";
import { Separator } from "./ui/separator";
import HealthIndicator from "./HealthIndicator";
import { useMcpIntegrationStore } from "@/store/mcpIntegrationStore";

function formatUptime(seconds: number) {
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = seconds % 60;
  if (h > 0) return `${h}h ${m}m ${s}s`;
  if (m > 0) return `${m}m ${s}s`;
  return `${s}s`;
}

export default function McpServerRuntimeCard() {
  const {
    status,
    serverPort,
    protocolVersion,
    uptime,
    latency,
    healthChecks,
  } = useMcpIntegrationStore();

  if (status !== "connected" && status !== "error") return null;

  const healthyCount = healthChecks.filter((h) => h.level === "healthy").length;
  const degradedCount = healthChecks.filter((h) => h.level === "degraded").length;
  const unhealthyCount = healthChecks.filter((h) => h.level === "unhealthy").length;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <Cpu className="h-4 w-4 text-muted-foreground" />
          Servidor MCP local
        </CardTitle>
        <CardDescription>
          Runtime local · puerto {serverPort} · {protocolVersion}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-3 gap-3 mb-4">
          <div className="rounded-md bg-muted/50 px-3 py-2 text-center">
            <div className="flex items-center justify-center gap-1 text-xs text-muted-foreground mb-0.5">
              <Timer className="h-3 w-3" />
              Uptime
            </div>
            <p className="text-sm font-mono font-medium tabular-nums">
              {formatUptime(uptime)}
            </p>
          </div>
          <div className="rounded-md bg-muted/50 px-3 py-2 text-center">
            <div className="flex items-center justify-center gap-1 text-xs text-muted-foreground mb-0.5">
              <Activity className="h-3 w-3" />
              Latencia
            </div>
            <p className="text-sm font-mono font-medium tabular-nums">
              {latency}ms
            </p>
          </div>
          <div className="rounded-md bg-muted/50 px-3 py-2 text-center">
            <div className="text-xs text-muted-foreground mb-0.5">Memoria</div>
            <p className="text-sm font-mono font-medium tabular-nums">34 MB</p>
          </div>
        </div>

        <Separator className="mb-3" />

        <div className="flex items-center justify-between mb-3">
          <span className="text-xs text-muted-foreground">
            Health checks
          </span>
          <div className="flex items-center gap-2">
            <Badge variant="success" className="text-[10px]">{healthyCount} ok</Badge>
            {degradedCount > 0 && (
              <Badge variant="warning" className="text-[10px]">{degradedCount} warn</Badge>
            )}
            {unhealthyCount > 0 && (
              <Badge variant="destructive" className="text-[10px]">{unhealthyCount} fail</Badge>
            )}
          </div>
        </div>

        <div className="space-y-2">
          {healthChecks.map((hc) => (
            <div
              key={hc.label}
              className="flex items-center justify-between rounded-md bg-muted/30 px-3 py-2"
            >
              <div className="flex items-center gap-2 min-w-0">
                <HealthIndicator level={hc.level} />
                <span className="text-xs text-foreground truncate">{hc.label}</span>
              </div>
              <span className="text-[10px] text-muted-foreground text-right ml-2 shrink-0">
                {hc.detail}
              </span>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
