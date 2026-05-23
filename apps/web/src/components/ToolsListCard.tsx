import { useEffect, useState } from "react";
import { Wrench, ChevronDown, ChevronRight, Server, Bot, Plug } from "lucide-react";
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

interface ToolDef {
  name: string;
  description: string;
  origin: string;
  enabled: boolean;
}

interface ToolGroup {
  origin: string;
  tools: ToolDef[];
}

const originIcons: Record<string, React.ReactNode> = {
  "mcp-station": <Server className="h-3.5 w-3.5" />,
  opencode: <Bot className="h-3.5 w-3.5" />,
};

const originLabels: Record<string, string> = {
  "mcp-station": "MCP Station",
  opencode: "OpenCode",
};

async function toggleToolApi(name: string) {
  const res = await fetch(`/api/tools/${encodeURIComponent(name)}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
  });
  if (!res.ok) return null;
  return res.json();
}

export default function ToolsListCard() {
  const [groups, setGroups] = useState<ToolGroup[]>([]);
  const [collapsed, setCollapsed] = useState<Record<string, boolean>>({});
  const [loading, setLoading] = useState(true);

  const load = async () => {
    try {
      const res = await fetch("/api/tools");
      if (!res.ok) return;
      const data: ToolGroup[] = await res.json();
      setGroups(data);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const handleToggle = async (name: string) => {
    const result = await toggleToolApi(name);
    if (result) {
      setGroups((prev) =>
        prev.map((g) => ({
          ...g,
          tools: g.tools.map((t) =>
            t.name === name ? { ...t, enabled: result.enabled } : t
          ),
        }))
      );
    }
  };

  const totalTools = groups.reduce((n, g) => n + g.tools.length, 0);
  const enabledTools = groups.reduce(
    (n, g) => n + g.tools.filter((t) => t.enabled).length,
    0
  );

  if (loading) return null;
  if (totalTools === 0) return null;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <Wrench className="h-4 w-4 text-muted-foreground" />
          Tools disponibles
        </CardTitle>
        <CardDescription>
          {enabledTools} de {totalTools} tools habilitadas · {groups.length} origenes
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {groups.map((group) => {
          const isCollapsed = collapsed[group.origin] ?? false;
          const icon = originIcons[group.origin] ?? (
            <Plug className="h-3.5 w-3.5" />
          );
          const label =
            originLabels[group.origin] ?? group.origin;

          return (
            <div key={group.origin}>
              <button
                onClick={() =>
                  setCollapsed((c) => ({
                    ...c,
                    [group.origin]: !isCollapsed,
                  }))
                }
                className="flex items-center gap-2 w-full text-left hover:text-foreground transition-colors"
              >
                {isCollapsed ? (
                  <ChevronRight className="h-3.5 w-3.5 text-muted-foreground" />
                ) : (
                  <ChevronDown className="h-3.5 w-3.5 text-muted-foreground" />
                )}
                {icon}
                <span className="text-sm font-medium">{label}</span>
                <Badge variant="secondary" className="text-[10px]">
                  {group.tools.filter((t) => t.enabled).length}/
                  {group.tools.length}
                </Badge>
              </button>

              {!isCollapsed && (
                <div className="mt-2 ml-7 space-y-1">
                  {group.tools.map((tool, i) => (
                    <div key={tool.name}>
                      {i > 0 && <Separator className="my-1.5" />}
                      <div className="flex items-center justify-between gap-3 py-1">
                        <div className="min-w-0">
                          <div className="flex items-center gap-2">
                            <code className="text-xs font-mono text-foreground">
                              {tool.name}
                            </code>
                            <Badge
                              variant={
                                tool.enabled ? "success" : "secondary"
                              }
                              className="text-[10px]"
                            >
                              {tool.enabled ? "on" : "off"}
                            </Badge>
                          </div>
                          <p className="text-xs text-muted-foreground mt-0.5 truncate">
                            {tool.description}
                          </p>
                        </div>
                        <Switch
                          checked={tool.enabled}
                          onCheckedChange={() =>
                            handleToggle(tool.name)
                          }
                          className="shrink-0"
                        />
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          );
        })}
      </CardContent>
    </Card>
  );
}
