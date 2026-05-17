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
import type { IntegrationStatus } from "@/store/mcpIntegrationStore";

const statusConfig: Record<
  IntegrationStatus,
  { label: string; variant: "success" | "destructive" | "secondary" | "warning" }
> = {
  empty: { label: "Sin configurar", variant: "secondary" },
  connecting: { label: "Conectando...", variant: "warning" },
  connected: { label: "Conectado", variant: "success" },
  error: { label: "Error", variant: "destructive" },
};

function formatDate(iso: string | null) {
  if (!iso) return "Nunca";
  return new Date(iso).toLocaleString("es-CL", {
    dateStyle: "medium",
    timeStyle: "medium",
  });
}

export default function ConnectionStatusCard() {
  const { status, endpoint, lastConnected, latency } =
    useMcpIntegrationStore();

  if (status === "empty") return null;

  const cfg = statusConfig[status];

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="text-base">Estado de integracion</CardTitle>
          <Badge variant={cfg.variant}>{cfg.label}</Badge>
        </div>
        <CardDescription>
          Ultima conexion: {formatDate(lastConnected)}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Separator className="mb-4" />
        <div className="space-y-2 text-sm">
          <div>
            <span className="text-muted-foreground">Endpoint: </span>
            <code className="font-mono text-xs text-foreground break-all">
              {endpoint}
            </code>
          </div>
          <div className="flex items-center gap-4 text-xs text-muted-foreground pt-1">
            <span>Latencia: {latency}ms</span>
            <HealthIndicator
              level={status === "connected" ? "healthy" : "unhealthy"}
              label={status === "connected" ? "Endpoint vivo" : "Sin respuesta"}
              size="sm"
            />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
