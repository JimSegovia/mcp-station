import DashboardSummary from "@/components/DashboardSummary";
import { Separator } from "@/components/ui/separator";

export default function Dashboard() {
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
        Los datos de este panel reflejan el estado actual de los stores
        locales. Cuando el backend Go este integrado, estos valores se
        sincronizaran en tiempo real via API.
      </p>
    </div>
  );
}
