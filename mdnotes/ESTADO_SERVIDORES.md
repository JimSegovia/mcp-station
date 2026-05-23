# Estado real de servidores, modulos y tools

> ✅ = funcional (hace algo real) | ❌ = mock/simulado | 🚫 = bloqueado por politica

---

## 1. Servidores MCP (tabla `servers`)

Registros en la DB. Se administran desde el panel **Servidores MCP**.

| Servidor | Estado | Real? | Detalle |
|---|---|---|---|
| **OpenCode Local** (STDIO) | `connected` | ❌ Mock | Registro demo. No hay proceso `opencode serve` real corriendo. El status es ficticio. |
| **File System Proxy** (WS) | `disconnected` | ❌ Mock | Apunta a `ws://localhost:8081/mcp` donde no hay nada. Es un placeholder para que el usuario registre uno real. |

### Como agregar uno real

1. Anda al panel **Servidores MCP**
2. Clic **Agregar**
3. Pone nombre, tipo (`stdio` o `websocket`), endpoint real
4. Clic guardar

El registro se persiste en SQLite. La conexion real al servidor externo (via `mcp/client.go` saliente) esta pendiente para fase 2.

---

## 2. Runtime MCP local (servidor inbound)

Corre dentro del proceso Go en `ws://localhost:8080/ws`. Expone 6 tools via JSON-RPC 2.0.

| Tool | ¿Real? | Si la llamas... |
|---|---|---|
| `list_directory` | ✅ Real | `os.ReadDir` del path pedido |
| `read_file` | ✅ Real | `os.ReadFile` del path pedido |
| `create_file` | ✅ Real | `os.WriteFile` en el path pedido (crea directorios si no existen) |
| `execute_command` | 🚫 Bloqueado | Responde "restringido por politica". No ejecuta nada. |
| `open_application` | ❌ Mock | Responde "abierto (simulado)". No abre apps reales. |
| `opencode_task` | ❌ Mock | Responde "tarea despachada (simulado)". No llama a OpenCode. |

### Como probarlas

```bash
wscat -c ws://localhost:8080/ws
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}
{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_directory","arguments":{"path":"."}}}
```

Estas mismas 6 tools son las que XiaoZhi ve cuando se conecta via `mcp/client.go`.

---

## 3. Modulos (tabla `modules`)

Registros en la DB. Se administran desde el panel **Modulos**.

| Modulo | Estado default | Real? | Detalle |
|---|---|---|---|
| **OpenCode Adapter** | `activo` | ❌ Solo registro | El switch en la UI prende/apaga el registro en DB. No controla ningun proceso real. |
| **File System Tools** | `activo` | ❌ Solo registro | Idem. Las tools reales (`list_directory`, `read_file`, `create_file`) funcionan sin importar el estado de este modulo. |
| **Terminal Runner** | `inactivo` | ❌ Solo registro | Idem. |
| **Desktop Automation** | `inactivo` | ❌ Solo registro | Idem. |
| **Browser Tools** | `inactivo` | ❌ Solo registro | Idem. |

### Pendiente

La conexion entre modulo y tool no esta cableada. Activar/desactivar un modulo no afecta si las tools funcionan o no. Es una pieza para HU-009/011.

---

## 4. Cliente MCP saliente (conexion a XiaoZhi)

Archivo: `internal/mcp/client.go`

| Funcionalidad | Estado | Detalle |
|---|---|---|
| Dial a `wss://api.xiaozhi.me/mcp/?token=...` | ✅ Real | Usa `gorilla/websocket` |
| Handshake MCP (`initialize`) | ✅ Real | Responde con server info y capabilities |
| `tools/list` | ✅ Real | Responde con las 6 tools del registry |
| `tools/call` | ✅ Real | Despacha a la tool y ejecuta (3 reales, 3 mock) |
| Keepalive ping/pong | ✅ Real | Ping cada 30s, read deadline 60s |
| Reconexion automatica | ✅ Real | Reintenta cada 5s si la conexion se cae |
| Sincronizacion con DB | ✅ Real | `connected` / `error` / `disconnected` se reflejan en `integrations` |

---

## 5. Dashboards y monitoreo

| Endpoint | Estado | Detalle |
|---|---|---|
| `GET /api/dashboard` | ✅ Real | Agregado de los 4 stores |
| `GET /api/system` | ✅ Real | Memoria, goroutines, uptime, GC |
| `GET /api/integration` | ✅ Real | Estado actual del endpoint XiaoZhi |
| `GET /api/logs` | ✅ Real | Filtrable por tipo, fuente, resultado |

---

## Resumen rapido

| Capa | Que es real | Que es mock |
|---|---|---|
| **Tools del runtime** | `list_directory`, `read_file`, `create_file` | `open_application`, `opencode_task` |
| **Bloqueado** | — | `execute_command` |
| **Conexion a XiaoZhi** | 100% real (dial, handshake, tools, keepalive) | — |
| **Servidores registrados** | — | Los 2 del seeder son demo |
| **Modulos** | — | Los 5 son solo registros DB |
| **UI** | 100% real (React + API) | — |
| **DB** | 100% real (SQLite, migraciones, seeder) | — |
