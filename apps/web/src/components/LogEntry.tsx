import { cn } from "@/lib/utils";
import { Badge } from "./ui/badge";
import type { LogEntry, LogType, LogResult } from "@/store/logStore";

const typeBadge: Record<LogType, { label: string; variant: "success" | "destructive" | "warning" | "secondary" | "outline" }> = {
  tool_call: { label: "Tool", variant: "success" },
  connection: { label: "Conexion", variant: "outline" },
  module: { label: "Modulo", variant: "secondary" },
  error: { label: "Error", variant: "destructive" },
  security: { label: "Seguridad", variant: "warning" },
};

const resultBadge: Record<LogResult, { label: string; variant: "success" | "destructive" | "warning" }> = {
  success: { label: "OK", variant: "success" },
  error: { label: "Error", variant: "destructive" },
  blocked: { label: "Bloqueado", variant: "warning" },
};

function formatDate(iso: string) {
  return new Date(iso).toLocaleString("es-CL", {
    dateStyle: "short",
    timeStyle: "medium",
  });
}

export default function LogEntryRow({ log }: { log: LogEntry }) {
  return (
    <div
      className={cn(
        "flex items-start gap-3 py-3 px-4 rounded-md border border-transparent transition-colors hover:bg-muted/30",
        log.result === "error" && "border-l-2 border-l-destructive bg-destructive/5"
      )}
    >
      <div className="shrink-0 mt-0.5">
        <Badge variant={typeBadge[log.type].variant} className="text-[10px]">
          {typeBadge[log.type].label}
        </Badge>
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="text-xs font-medium text-muted-foreground">
            {log.source}
          </span>
          <Badge variant={resultBadge[log.result].variant} className="text-[10px]">
            {resultBadge[log.result].label}
          </Badge>
        </div>
        <p className="text-sm text-foreground mt-0.5">{log.message}</p>
      </div>
      <span className="text-xs text-muted-foreground shrink-0">
        {formatDate(log.timestamp)}
      </span>
    </div>
  );
}
