import { AlertCircle } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "./ui/card";
import { useMcpIntegrationStore } from "@/store/mcpIntegrationStore";

function formatDate(iso: string) {
  return new Date(iso).toLocaleString("es-CL", {
    dateStyle: "medium",
    timeStyle: "medium",
  });
}

export default function RecentErrorCard() {
  const { lastError } = useMcpIntegrationStore();

  if (!lastError) return null;

  return (
    <Card className="border-destructive/30">
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2 text-destructive">
          <AlertCircle className="h-4 w-4" />
          Error reciente
        </CardTitle>
        <CardDescription>
          {formatDate(lastError.timestamp)}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <p className="text-sm">{lastError.message}</p>
      </CardContent>
    </Card>
  );
}
