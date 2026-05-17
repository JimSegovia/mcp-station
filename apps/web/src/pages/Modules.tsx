import { useModuleStore } from "@/store/moduleStore";
import ModuleCard from "@/components/ModuleCard";
import { Separator } from "@/components/ui/separator";

export default function Modules() {
  const { modules, toggleModule, setModuleError, clearModuleError } =
    useModuleStore();

  const activeCount = modules.filter((m) => m.enabled).length;
  const errorCount = modules.filter((m) => m.status === "error").length;

  return (
    <div className="max-w-2xl space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Modulos</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Activa o desactiva los modulos del sistema.{" "}
          {activeCount} de {modules.length} activos
          {errorCount > 0 && ` · ${errorCount} con error`}
        </p>
      </div>

      <div className="space-y-3">
        {modules.map((mod) => (
          <ModuleCard
            key={mod.id}
            module={mod}
            onToggle={() => toggleModule(mod.id)}
            onClearError={() => clearModuleError(mod.id)}
            onSimulateError={() =>
              setModuleError(
                mod.id,
                `Error en ${mod.name}: respuesta inesperada del runtime local`
              )
            }
          />
        ))}
      </div>

      <div className="flex items-center gap-3">
        <Separator className="flex-1" />
        <span className="text-xs text-muted-foreground">demo</span>
        <Separator className="flex-1" />
      </div>
      <p className="text-xs text-muted-foreground text-center">
        Modulos predefinidos para validar la UI de administracion. El backend
        Go gestionara el estado real de cada modulo.
      </p>
    </div>
  );
}
