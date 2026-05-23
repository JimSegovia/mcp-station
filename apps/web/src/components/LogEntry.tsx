import { cn } from "@/lib/utils";
import { Badge } from "./ui/badge";
import type { LogEntry } from "@/store/logStore";

const typeBadge: Record<string, { label: string; variant: "success" | "destructive" | "warning" | "secondary" | "outline" }> = {
  tool_call: { label: "Tool", variant: "success" },
  connection: { label: "Conexion", variant: "outline" },
  module: { label: "Modulo", variant: "secondary" },
  error: { label: "Error", variant: "destructive" },
  security: { label: "Seguridad", variant: "warning" },
};

const fallbackType = { label: "Info", variant: "outline" as const };

const resultBadge: Record<string, { label: string; variant: "success" | "destructive" | "warning" | "outline" }> = {
  success: { label: "OK", variant: "success" },
  error: { label: "Error", variant: "destructive" },
  blocked: { label: "Bloqueado", variant: "warning" },
  info: { label: "Info", variant: "outline" },
  warning: { label: "Warn", variant: "warning" },
};

const fallbackResult = { label: "?", variant: "outline" as const };

function formatDate(iso: string) {
  try {
    return new Date(iso).toLocaleString("es-CL", {
      dateStyle: "short",
      timeStyle: "medium",
    });
  } catch {
    return iso || "-";
  }
}

export default function LogEntryRow({ log }: { log: LogEntry }) {
  const type = (typeBadge[log.type] ?? fallbackType);
  const result = (resultBadge[log.result] ?? fallbackResult);

  return (
    <div
      className={cn(
        "flex items-start gap-3 py-3 px-4 rounded-md border border-transparent transition-colors hover:bg-muted/30",
        log.result === "error" && "border-l-2 border-l-destructive bg-destructive/5"
      )}
    >
      <div className="shrink-0 mt-0.5">
        <Badge variant={type.variant} className="text-[10px]">
          {type.label}
        </Badge>
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="text-xs font-medium text-muted-foreground">
            {log.source ?? "-"}
          </span>
          <Badge variant={result.variant} className="text-[10px]">
            {result.label}
          </Badge>
        </div>
        <p className="text-sm text-foreground mt-0.5">{log.message ?? ""}</p>
      </div>
      <span className="text-xs text-muted-foreground shrink-0">
        {formatDate(log.timestamp)}
      </span>
    </div>
  );
}
