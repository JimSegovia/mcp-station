import { Plug } from "lucide-react";

export default function EmptyIntegrationState() {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      <div className="rounded-full bg-muted p-4 mb-5">
        <Plug className="h-8 w-8 text-muted-foreground" />
      </div>
      <h3 className="text-lg font-semibold mb-2">
        Sin integracion configurada
      </h3>
      <p className="text-sm text-muted-foreground max-w-sm">
        Todavia no has registrado el MCP endpoint de Custom Services.
        Sigue los pasos de abajo para obtenerlo desde la consola oficial
        de XiaoZhi y conectarlo a tu servidor MCP local.
      </p>
    </div>
  );
}
