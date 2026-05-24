# Playwright MCP Server Integration

**Playwright MCP** es un servidor virtual integrado en el backend de Go (`apps/server`) que expone herramientas para automatizar e interactuar con navegadores web (abrir páginas, hacer clic, llenar formularios, tomar capturas de pantalla, etc.) utilizando Playwright.

---

## 1. Funcionamiento del Servidor
El servidor corre de forma local e independiente en un subproceso iniciado por el backend de Go (`apps/server/internal/playwright/manager.go`).

Una vez activo, sus herramientas (`playwright_navigate`, `playwright_click`, etc.) se registran en el **Registry de Herramientas** interno del MCP Station y quedan disponibles para ser consumidas por cualquier cliente conectado, incluyendo el agente de XiaoZhi.

---

## 2. Tipos de Puertos y URLs

El servidor Playwright MCP puede ejecutarse con dos modos de puertos:

### A. Puerto Dinámico (Por Defecto)
Cada vez que se active o inicie el servidor de Go, se buscará y reservará un puerto TCP libre aleatorio para evitar colisiones con otros puertos en uso:
- **Inicio estándar**:
  ```bash
  go run main.go
  ```
- El endpoint dinámico se mostrará en el frontend en el panel **Servidores MCP** > **Playwright MCP**, por ejemplo: `http://localhost:54321/mcp`.

---

### B. Puerto Fijo (Configuración Estática)
Si necesitas usar la herramienta con clientes MCP externos (como una sesión independiente de **OpenCode** que requiere definir la URL exacta del servidor en su configuración `mcp.json` de forma estática), puedes fijar el puerto al iniciar el backend:

- **Inicio con puerto fijo**:
  ```bash
  go run main.go -playwright-port 9090
  ```
- **Endpoint estático resultante**:
  `http://localhost:9090/mcp`

De esta forma, no tendrás que cambiar la URL configurada en OpenCode cada vez que reinicies el servidor de Go.

---

## 3. Uso en Prompts

### Opción 1: A través del Agente de XiaoZhi (Recomendado)
Cuando conectas el frontend a tu agente de XiaoZhi, las herramientas de Playwright ya están expuestas por el MCP Station. Puedes indicarle en el prompt:
> *"Usa las herramientas de `playwright` locales para abrir el navegador en modo headless, navega a `https://github.com` y dime qué encuentras en la página principal."*

### Opción 2: Para OpenCode directamente
Si estás usando OpenCode con la configuración del puerto estático (ej: `9090`):
> *"Conéctate al servidor MCP local de Playwright en `http://localhost:9090/mcp` y usa sus herramientas de navegación para abrir `https://google.com` y buscar 'MCP Station'."*
