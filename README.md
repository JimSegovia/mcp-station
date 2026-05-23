# MCP Station

Administrador local de servidores MCP con delegacion a OpenCode. Conecta tu agente XiaoZhi a herramientas reales del sistema operativo potenciadas por OpenCode.

## Que hace

```
XiaoZhi (voz)
    │
    ▼
┌─────────────────────┐
│     MCP Station      │  ← esta app
│  ┌───────────────┐   │
│  │ OpenCode Serve │   │  ← opencode como cerebro
│  └───────────────┘   │
│  ┌───────────────┐   │
│  │  Tool Registry │   │  ← tools reales (filesystem, comandos, apps)
│  └───────────────┘   │
└─────────────────────┘
```

- **Delega tareas a OpenCode** — cuando XiaoZhi pide algo complejo, MCP Station se lo pasa a OpenCode (`opencode serve`) que tiene acceso a modelos potentes y sus propias herramientas
- **Expone tools reales** — `read_file`, `create_file`, `list_directory`, `open_application`, `execute_command` (con whitelist)
- **Conecta con XiaoZhi** via `Custom Services` MCP endpoint (WebSocket `wss://api.xiaozhi.me/mcp/?token=...`)
- **Interfaz web** para administrar modulos, servidores MCP, tools y monitorear conexiones

## Requisitos

- **Go 1.24+** — backend
- **Node.js 20+** — frontend
- **OpenCode CLI** instalado (`curl -fsSL https://opencode.ai/install | bash`)

## Inicializacion

```bash
# 1. Instalar dependencias
cd apps/web && npm install && cd ../..
cd apps/server && go mod tidy && cd ../..

# 2. Iniciar backend (puerto 8080)
cd apps/server && go run .

# 3. Iniciar frontend (puerto 5173, proxy al :8080)
cd apps/web && npm run dev
```

El backend arranca automaticamente `opencode serve` en el puerto 4096. Si ya tienes uno corriendo, lo detecta y se conecta a el.

Abre `http://localhost:5173` para ver la interfaz.

## Arquitectura

```
apps/
├── server/          # Backend Go
│   ├── main.go      # Entry point, arranca OpenCode serve
│   └── internal/
│       ├── api/     # HTTP API (REST + WebSocket MCP en /ws)
│       ├── mcp/     # Runtime (inbound WS) + Client (outbound a XiaoZhi)
│       ├── opencode/# HTTP client contra opencode serve, manager de proceso
│       ├── tool/    # Registry de tools con implementaciones reales
│       ├── storage/ # SQLite (integraciones, modulos, servidores, logs)
│       └── model/   # Tipos compartidos
└── web/             # Frontend React + TypeScript + Vite + Tailwind
    └── src/
        ├── pages/       # Integracion MCP, Modulos, Servidores, Monitor, Logs
        ├── components/  # Cards, ToolTestPanel, UI components
        └── store/       # Zustand stores
```

## Paneles

| Ruta | Funcion |
|---|---|
| `/` | Dashboard — estado general, modulos activos, conexion |
| `/mcp` | **Integracion MCP** — conectar/desconectar endpoint XiaoZhi, probar tools |
| `/modulos` | Activar/desactivar modulos (OpenCode Adapter, File System, etc.) |
| `/servidores` | Registrar servidores MCP externos adicionales |
| `/opencode` | Sesiones OpenCode — tracking, limpieza, estado |
| `/logs` | Bitacora de eventos, tool calls, errores |
| `/monitor` | Uso de recursos del servidor |

## Flujo con XiaoZhi

1. Ve a la consola de XiaoZhi → `Configure Role` → `MCP Settings` → `Custom Services`
2. Copia el endpoint `wss://api.xiaozhi.me/mcp/?token=...`
3. Pegalo en MCP Station (`/mcp`) y haz clic en **Conectar**
4. MCP Station se conecta al relay de XiaoZhi y expone sus tools
5. Cuando le hablas a XiaoZhi, el agente puede usar `opencode_ask`, `read_file`, `create_file`, etc.

## Tools disponibles

| Tool | Funcion |
|---|---|
| `read_file` | Lee archivos del sistema |
| `create_file` | Crea archivos con contenido |
| `list_directory` | Lista directorios |
| `open_application` | Abre aplicaciones (xdg-open / open / start) |
| `execute_command` | Ejecuta comandos con whitelist |
| `opencode_ask` | Envia prompt a OpenCode y espera respuesta |
| `opencode_run` | Despacha tarea async a OpenCode |

## Configuracion de OpenCode

El backend usa `opencode serve` internamente. Para configurar modelos, API keys y MCP servers de OpenCode, usa tu `opencode.jsonc` habitual:

```jsonc
{
  "mcp": {
    "mcp-station": {
      "type": "remote",
      "url": "http://localhost:8081/mcp"
    }
  }
}
```

Opcional: setea `OPENCODE_SERVER_PASSWORD` si quieres proteger el acceso:

```bash
OPENCODE_SERVER_PASSWORD=tu-password go run .
```

## Variables de entorno

| Variable | Default | Descripcion |
|---|---|---|
| `OPENCODE_SERVER_USERNAME` | `opencode` | Usuario para basic auth |
| `OPENCODE_SERVER_PASSWORD` | *(ninguna)* | Password para basic auth |

## Stack

- **Backend**: Go, SQLite, gorilla/websocket, JSON-RPC 2.0
- **Frontend**: React, TypeScript, Tailwind CSS, shadcn/ui, Zustand, Vite
- **Protocolo**: MCP sobre WebSocket (JSON-RPC 2.0)
