import { Package } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "./ui/card";
import { Badge } from "./ui/badge";
import { Switch } from "./ui/switch";
import { Separator } from "./ui/separator";
import { Button } from "./ui/button";
import type { Module } from "@/store/moduleStore";

const statusBadge: Record<string, { label: string; variant: "success" | "destructive" | "secondary" }> = {
  ok: { label: "Activo", variant: "success" },
  error: { label: "Error", variant: "destructive" },
  inactive: { label: "Inactivo", variant: "secondary" },
};

interface ModuleCardProps {
  module: Module;
  onToggle: () => void;
  onClearError: () => void;
  onSimulateError: () => void;
}

export default function ModuleCard({
  module: mod,
  onToggle,
  onClearError,
  onSimulateError,
}: ModuleCardProps) {
  const badge = statusBadge[mod.status];

  return (
    <Card className={mod.status === "error" ? "border-destructive/30" : ""}>
      <CardHeader>
        <div className="flex items-start justify-between gap-3">
          <div className="flex items-center gap-2 min-w-0">
            <Package className="h-4 w-4 text-muted-foreground shrink-0" />
            <div>
              <CardTitle className="text-base">{mod.name}</CardTitle>
              <CardDescription className="mt-0.5">
                {mod.description}
              </CardDescription>
            </div>
          </div>
          <Switch checked={mod.enabled} onCheckedChange={onToggle} />
        </div>
      </CardHeader>
      <CardContent>
        <div className="flex items-center justify-between">
          <Badge variant={badge.variant}>{badge.label}</Badge>
          {mod.status === "error" && (
            <Button
              variant="outline"
              size="sm"
              onClick={onClearError}
            >
              Limpiar error
            </Button>
          )}
          {mod.enabled && mod.status === "ok" && (
            <Button
              variant="destructive"
              size="sm"
              onClick={onSimulateError}
            >
              Simular error
            </Button>
          )}
        </div>
        {mod.lastError && (
          <>
            <Separator className="my-3" />
            <p className="text-xs text-destructive">{mod.lastError}</p>
          </>
        )}
      </CardContent>
    </Card>
  );
}
