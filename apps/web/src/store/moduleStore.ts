import { create } from "zustand";

type ModuleStatus = "ok" | "error" | "inactive";

export interface Module {
  id: string;
  name: string;
  description: string;
  enabled: boolean;
  status: ModuleStatus;
  lastError: string | null;
}

interface ModuleStore {
  modules: Module[];

  toggleModule: (id: string) => void;
  setModuleError: (id: string, message: string) => void;
  clearModuleError: (id: string) => void;
}

const defaultModules: Module[] = [
  {
    id: "opencode",
    name: "OpenCode Adapter",
    description: "Motor central de ejecucion inteligente. Envia tareas complejas a OpenCode y recibe resultados.",
    enabled: true,
    status: "ok",
    lastError: null,
  },
  {
    id: "filesystem",
    name: "File System Tools",
    description: "Operaciones controladas sobre archivos y directorios. Lectura, escritura y navegacion segura.",
    enabled: true,
    status: "ok",
    lastError: null,
  },
  {
    id: "terminal",
    name: "Terminal Runner",
    description: "Ejecuta comandos controlados en la terminal del sistema con lista blanca de comandos permitidos.",
    enabled: false,
    status: "inactive",
    lastError: null,
  },
  {
    id: "desktop",
    name: "Desktop Automation",
    description: "Controla aplicaciones del escritorio. Abrir, cerrar, enfocar ventanas y simular atajos.",
    enabled: false,
    status: "inactive",
    lastError: null,
  },
  {
    id: "browser",
    name: "Browser Tools",
    description: "Navegacion web automatizada. Abrir URLs, extraer contenido y ejecutar acciones en el navegador.",
    enabled: false,
    status: "inactive",
    lastError: null,
  },
];

export const useModuleStore = create<ModuleStore>((set) => ({
  modules: defaultModules,

  toggleModule: (id: string) =>
    set((state) => ({
      modules: state.modules.map((m) =>
        m.id === id
          ? {
              ...m,
              enabled: !m.enabled,
              status: m.enabled ? ("inactive" as ModuleStatus) : ("ok" as ModuleStatus),
              lastError: null,
            }
          : m
      ),
    })),

  setModuleError: (id: string, message: string) =>
    set((state) => ({
      modules: state.modules.map((m) =>
        m.id === id
          ? {
              ...m,
              status: "error" as ModuleStatus,
              lastError: message,
            }
          : m
      ),
    })),

  clearModuleError: (id: string) =>
    set((state) => ({
      modules: state.modules.map((m) =>
        m.id === id
          ? {
              ...m,
              status: m.enabled ? ("ok" as ModuleStatus) : ("inactive" as ModuleStatus),
              lastError: null,
            }
          : m
      ),
    })),
}));
