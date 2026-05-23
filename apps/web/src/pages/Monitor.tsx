import { useEffect } from "react";
import { Cpu, MemoryStick, Timer, Gauge } from "lucide-react";
import { useSystemStore } from "@/store/systemStore";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

function fmtBytes(b: number): string {
  if (b >= 1073741824) return (b / 1073741824).toFixed(1) + " GB";
  if (b >= 1048576) return (b / 1048576).toFixed(1) + " MB";
  if (b >= 1024) return (b / 1024).toFixed(1) + " KB";
  return b + " B";
}

function fmtUptime(s: number): string {
  const d = Math.floor(s / 86400);
  const h = Math.floor((s % 86400) / 3600);
  const m = Math.floor((s % 3600) / 60);
  if (d > 0) return `${d}d ${h}h ${m}m`;
  if (h > 0) return `${h}h ${m}m`;
  if (m > 0) return `${m}m`;
  return `${s}s`;
}

export default function Monitor() {
  const { stats, load } = useSystemStore();

  useEffect(() => {
    load();
    const interval = setInterval(load, 2000);
    return () => clearInterval(interval);
  }, [load]);

  if (!stats) {
    return (
      <div className="max-w-2xl space-y-6">
        <h1 className="text-2xl font-bold tracking-tight">Monitor</h1>
        <p className="text-sm text-muted-foreground">Cargando estadisticas del sistema...</p>
      </div>
    );
  }

  const { memory } = stats;
  const heapUsagePct = memory.heapSys > 0 ? ((memory.heapAlloc / memory.heapSys) * 100).toFixed(1) : "0";
  const sysUsagePct = memory.sys > 0 ? ((memory.alloc / memory.sys) * 100).toFixed(1) : "0";

  return (
    <div className="max-w-2xl space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Monitor</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Uso de recursos del servidor MCP Station
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">CPU</CardTitle>
            <Cpu className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-xl font-bold">{stats.numCPU} cores</div>
            <p className="text-xs text-muted-foreground mt-1">{stats.goroutines} goroutines activas</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Uptime</CardTitle>
            <Timer className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-xl font-bold">{fmtUptime(stats.uptime)}</div>
            <p className="text-xs text-muted-foreground mt-1">PID {stats.pid}</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Go Runtime</CardTitle>
            <Gauge className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-xl font-bold">{stats.goVersion}</div>
            <p className="text-xs text-muted-foreground mt-1">GC ciclos: {memory.numGC}</p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <MemoryStick className="h-4 w-4 text-muted-foreground" />
            Memoria
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-1">
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Heap en uso</span>
              <span className="font-mono text-xs">
                {fmtBytes(memory.heapAlloc)} / {fmtBytes(memory.heapSys)} ({heapUsagePct}%)
              </span>
            </div>
            <div className="h-2 bg-muted rounded-full overflow-hidden">
              <div
                className="h-full bg-blue-500 rounded-full transition-all duration-500"
                style={{ width: `${Math.min(Number(heapUsagePct), 100)}%` }}
              />
            </div>
          </div>

          <div className="space-y-1">
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Memoria total del SO</span>
              <span className="font-mono text-xs">
                {fmtBytes(memory.alloc)} / {fmtBytes(memory.sys)} ({sysUsagePct}%)
              </span>
            </div>
            <div className="h-2 bg-muted rounded-full overflow-hidden">
              <div
                className="h-full bg-emerald-500 rounded-full transition-all duration-500"
                style={{ width: `${Math.min(Number(sysUsagePct), 100)}%` }}
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4 pt-2">
            <div>
              <span className="text-xs text-muted-foreground">Total asignado</span>
              <p className="text-sm font-mono">{fmtBytes(memory.totalAlloc)}</p>
            </div>
            <div>
              <span className="text-xs text-muted-foreground">Del sistema</span>
              <p className="text-sm font-mono">{fmtBytes(memory.sys)}</p>
            </div>
            <div>
              <span className="text-xs text-muted-foreground">Goroutines</span>
              <p className="text-sm font-mono">{stats.goroutines}</p>
            </div>
            <div>
              <span className="text-xs text-muted-foreground">GC ciclos</span>
              <p className="text-sm font-mono">{memory.numGC}</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
