import { useLogStore } from "@/store/logStore";
import LogEntryRow from "@/components/LogEntry";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { Trash2 } from "lucide-react";

export default function Logs() {
  const { logs, clearLogs } = useLogStore();

  return (
    <div className="max-w-3xl space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Logs</h1>
          <p className="text-sm text-muted-foreground mt-1">
            {logs.length} eventos registrados
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={clearLogs} disabled={logs.length === 0}>
          <Trash2 className="h-4 w-4 mr-1" />
          Limpiar
        </Button>
      </div>

      {logs.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-center text-muted-foreground">
          <p className="text-sm">No hay logs registrados.</p>
          <p className="text-xs mt-1">
            Las ejecuciones, conexiones y errores apareceran aqui.
          </p>
        </div>
      ) : (
        <div className="space-y-1">
          {logs.map((log) => (
            <LogEntryRow key={log.id} log={log} />
          ))}
        </div>
      )}

      <div className="flex items-center gap-3">
        <Separator className="flex-1" />
        <span className="text-xs text-muted-foreground">demo</span>
        <Separator className="flex-1" />
      </div>
      <p className="text-xs text-muted-foreground text-center">
        Logs mock para validar la UI. El backend Go producira entradas reales
        de ejecucion, conexion y errores.
      </p>
    </div>
  );
}
