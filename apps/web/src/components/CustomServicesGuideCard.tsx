import { HelpCircle } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "./ui/card";

export default function CustomServicesGuideCard() {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <HelpCircle className="h-4 w-4 text-muted-foreground" />
          Como obtener el MCP endpoint oficial
        </CardTitle>
        <CardDescription>
          Pasos para obtener la URL de Custom Services desde la consola de XiaoZhi
        </CardDescription>
      </CardHeader>
      <CardContent>
        <ol className="space-y-3 text-sm">
          <li className="flex gap-3">
            <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-medium">
              1
            </span>
            <span className="text-muted-foreground">
              Entra a la{" "}
              <strong className="text-foreground">consola de XiaoZhi</strong>
              {" "}en xiaozhi.me y selecciona tu agente
            </span>
          </li>
          <li className="flex gap-3">
            <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-medium">
              2
            </span>
            <span className="text-muted-foreground">
              Ve a{" "}
              <strong className="text-foreground">Configure Role</strong>
              {" "}y abre la seccion de{" "}
              <strong className="text-foreground">MCP Settings</strong>
            </span>
          </li>
          <li className="flex gap-3">
            <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-medium">
              3
            </span>
            <span className="text-muted-foreground">
              Haz clic en{" "}
              <strong className="text-foreground">Get MCP Endpoint</strong>
              {" "}para generar la URL de Custom Services
            </span>
          </li>
          <li className="flex gap-3">
            <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-medium">
              4
            </span>
            <span className="text-muted-foreground">
              Copia la URL generada y pegala en el campo de abajo para
              conectar tu servidor MCP local
            </span>
          </li>
        </ol>
      </CardContent>
    </Card>
  );
}
