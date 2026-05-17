import { Wrench } from "lucide-react";
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
import { useMcpIntegrationStore } from "@/store/mcpIntegrationStore";

export default function ToolsListCard() {
  const { tools, toggleTool } = useMcpIntegrationStore();

  if (tools.length === 0) return null;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <Wrench className="h-4 w-4 text-muted-foreground" />
          Tools expuestas
        </CardTitle>
        <CardDescription>
          {tools.filter((t) => t.enabled).length} de {tools.length} tools
          habilitadas para el endpoint
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-2">
        {tools.map((tool, i) => (
          <div key={tool.name}>
            {i > 0 && <Separator className="my-2" />}
            <div className="flex items-center justify-between gap-3 py-1">
              <div className="min-w-0">
                <div className="flex items-center gap-2">
                  <code className="text-sm font-mono text-foreground">
                    {tool.name}
                  </code>
                  <Badge variant={tool.enabled ? "success" : "secondary"} className="text-[10px]">
                    {tool.enabled ? "activo" : "inactivo"}
                  </Badge>
                </div>
                <p className="text-xs text-muted-foreground mt-0.5 truncate">
                  {tool.description}
                </p>
              </div>
              <Switch
                checked={tool.enabled}
                onCheckedChange={() => toggleTool(tool.name)}
              />
            </div>
          </div>
        ))}
      </CardContent>
    </Card>
  );
}
