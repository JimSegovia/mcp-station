import { cn } from "@/lib/utils";
import type { HealthLevel } from "@/store/mcpIntegrationStore";

const dotColors: Record<HealthLevel, string> = {
  healthy: "bg-emerald-400 shadow-[0_0_6px_rgba(52,211,153,0.4)]",
  degraded: "bg-amber-400 shadow-[0_0_6px_rgba(251,191,36,0.4)]",
  unhealthy: "bg-red-400 shadow-[0_0_6px_rgba(248,113,113,0.5)]",
  pending: "bg-sky-400 shadow-[0_0_6px_rgba(56,189,248,0.4)]",
  info: "bg-slate-400 shadow-[0_0_6px_rgba(148,163,184,0.35)]",
  unknown: "bg-zinc-400 shadow-[0_0_6px_rgba(161,161,170,0.35)]",
};

interface HealthIndicatorProps {
  level: HealthLevel;
  label?: string;
  className?: string;
  size?: "sm" | "md";
}

export default function HealthIndicator({
  level,
  label,
  className,
  size = "sm",
}: HealthIndicatorProps) {
  const sizeClass = size === "sm" ? "h-2 w-2" : "h-3 w-3";

  return (
    <div className={cn("flex items-center gap-1.5", className)} title={label}>
      <span
        className={cn(
          "inline-block rounded-full",
          sizeClass,
          dotColors[level],
          level === "unhealthy" ? "animate-ping" : ""
        )}
      />
      {label && (
        <span className="text-xs text-muted-foreground">{label}</span>
      )}
    </div>
  );
}
