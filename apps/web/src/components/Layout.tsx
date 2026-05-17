import { Outlet, NavLink } from "react-router-dom";
import { LayoutDashboard, Plug, Package, Server, ScrollText } from "lucide-react";

const navItems = [
  { to: "/", icon: LayoutDashboard, label: "Resumen" },
  { to: "/mcp", icon: Plug, label: "Integracion MCP" },
  { to: "/modulos", icon: Package, label: "Modulos" },
  { to: "/servidores", icon: Server, label: "Servidores MCP" },
  { to: "/logs", icon: ScrollText, label: "Logs" },
];

export default function Layout() {
  return (
    <div className="flex min-h-screen">
      <aside className="w-56 border-r border-border bg-card/50 flex flex-col">
        <div className="px-6 py-5 border-b border-border">
          <span className="text-sm font-bold tracking-widest text-muted-foreground uppercase">
            MCP Station
          </span>
        </div>
        <nav className="flex-1 px-3 py-4 space-y-1">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.to === "/"}
              className={({ isActive }) =>
                `flex items-center gap-3 rounded-md px-3 py-2 text-sm transition-colors ${
                  isActive
                    ? "bg-accent text-accent-foreground font-medium"
                    : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                }`
              }
            >
              <item.icon className="h-4 w-4" />
              {item.label}
            </NavLink>
          ))}
        </nav>
        <div className="px-6 py-4 border-t border-border">
          <p className="text-[10px] text-muted-foreground leading-relaxed">
            MCP Station v1 ·{" "}
            <span className="text-emerald-400/60">local-first</span>
          </p>
        </div>
      </aside>
      <main className="flex-1 p-8">
        <Outlet />
      </main>
    </div>
  );
}
