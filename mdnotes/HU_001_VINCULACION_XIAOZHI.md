# HU-001 Conectar XiaoZhi oficial a un MCP local

> ✅ = completado | Sin marca = pendiente

## Historia seleccionada

Empezamos por `HU-001 Conectar XiaoZhi oficial a un MCP local`.

Motivo:

- es la forma mas rapida de lograr valor real
- evita rehacer en `v1` voz, conversacion y `Speaker Recognition`
- aprovecha la consola oficial de XiaoZhi
- nos deja enfocarnos en la parte diferencial del producto:
  - montar el servidor MCP local
  - conectar tools locales
  - administrar modulos

## Objetivo de la historia

Permitir que un usuario que ya usa XiaoZhi en la consola oficial pueda conectar su agente a un servidor MCP local administrado por nuestra app, usando `Custom Services`.

## Enfoque de trabajo

Vamos a empezar por la interfaz.

Motivo:

- el usuario necesita entender con claridad que ya no esta vinculando el dispositivo a un backend propio
- el flujo real ahora depende de:
  - obtener el `MCP endpoint` oficial del agente
  - levantar un servidor MCP local
  - conectarlo correctamente
- si la interfaz explica bien ese flujo, luego backend, estado y logs se construyen con una meta mucho mas clara

## Stack recomendado para iniciar esta historia

La recomendacion para `HU-001` sigue siendo empezar con un stack `web-first`, aunque despues el producto tambien pueda tener una version desktop.

### Decision de stack para arrancar

- frontend: `React` + `TypeScript` + `Vite` ✅
- estilos: `Tailwind CSS` ✅
- componentes base: `shadcn/ui` ✅
- routing: `React Router` ✅
- estado de UI: `Zustand` ✅
- backend local: `Go`
- transporte principal de integracion: `WebSocket`
- API local: `HTTP + JSON`
- almacenamiento inicial: `SQLite`

### Como entra Electron

`Electron` si puede ser util despues, pero no deberia bloquear el inicio de `HU-001`.

La idea recomendada es:

1. construir primero la interfaz como aplicacion web local
2. conectar esa UI a un backend local en `Go`
3. validar la conexion del servidor MCP local con el endpoint oficial de XiaoZhi
4. cuando la base funcione, envolver la misma UI con `Electron` para tener version desktop

### Motivo de esta decision

Esto permite:

- avanzar mas rapido en la integracion real
- no mezclar desde el inicio UI, backend y empaquetado desktop
- reutilizar casi toda la interfaz despues
- mantener abierta la opcion de una futura plataforma web

### Lo que otra IA puede asumir desde ya

- la pantalla principal de esta historia sera una vista `React`
- el backend local expondra una API para consultar:
  - configuracion del endpoint
  - estado del servidor MCP
  - tools cargadas
  - errores recientes
- el backend local abrira la conexion saliente o facilitara la conexion al endpoint MCP oficial
- `Electron` no es requisito para el primer corte funcional

### Estructura sugerida de arranque

- `apps/web` ✅
- `apps/server`
- `packages/shared` si luego hace falta compartir tipos

### Criterio para pasar a version desktop

Solo tiene sentido agregar `Electron` cuando ya exista:

- lectura clara del `MCP endpoint` oficial
- servidor MCP local util
- estado visible de conexion
- tools registradas
- flujo basico de errores y reconexion

## Interface-first

### Resultado de interfaz que buscamos

El usuario debe entrar al sistema y encontrar una vista clara donde pueda:

- entender que esta integrando su agente oficial de XiaoZhi
- pegar o registrar el `MCP endpoint` oficial
- levantar el servidor MCP local
- ver si el endpoint esta conectado o no
- ver que tools estan disponibles
- entender cuando hubo error de conexion

### Flujo ideal del usuario

1. El usuario abre la consola oficial de XiaoZhi.
2. Entra a `Configure Role`.
3. Abre `MCP Settings`.
4. Obtiene el `MCP endpoint` de `Custom Services`.
5. Abre nuestra app local.
6. Registra o pega ese endpoint.
7. Nuestra app levanta o prepara el servidor MCP local.
8. El sistema muestra si la conexion fue exitosa y que tools quedaron expuestas.

### Pantalla minima sugerida

Nombre tentativo:

- `Integracion MCP`

Bloques recomendados:

- bloque de explicacion del flujo
- bloque del `MCP endpoint` oficial
- bloque de estado de conexion
- bloque del servidor MCP local
- bloque de tools disponibles
- bloque de error reciente

### Estado vacio

La pantalla debe tener un estado vacio claro cuando aun no se configuro el endpoint.

Debe mostrar:

- mensaje de que todavia no hay endpoint configurado
- breve explicacion de como obtenerlo desde la consola oficial
- campo para registrar o pegar el endpoint

### Estado conectado

Cuando la integracion ya este funcionando, la pantalla debe mostrar:

- estado `conectado`
- endpoint registrado
- tools disponibles
- hora de ultima conexion valida

### Estado con error

Si la conexion falla, la pantalla debe mostrar:

- estado `error`
- mensaje corto y entendible
- ultimo intento de conexion si existe

## Tareas de interfaz

### Tareas de producto

- definir que significa exactamente "integracion conectada" en v1
- definir que datos minimos del endpoint vamos a guardar
- definir que estados mostraremos en la interfaz
- definir que feedback debe ver el usuario cuando la conexion funciona o falla

### Tareas de UX

- definir el flujo paso a paso para obtener el endpoint oficial
- definir los estados vacio, conectando, conectado, desconectado y error
- definir los textos de ayuda de la pantalla
- decidir que informacion mostrar sin abrumar al usuario

### Tareas de UI

- crear una vista minima de integracion MCP ✅
- mostrar al menos: ✅
  - endpoint registrado ✅
  - estado actual ✅
  - ultima conexion ✅
  - tools disponibles ✅
  - error reciente si existe ✅
- permitir copiar el endpoint facilmente ✅
- mostrar una accion clara para iniciar o reiniciar el servidor MCP local ✅

### Componentes sugeridos

- `McpEndpointCard` ✅
- `ConnectionStatusCard` ✅
- `CustomServicesGuideCard` ✅
- `ToolsListCard` ✅
- `RecentErrorCard` ✅
- `EmptyIntegrationState` ✅

### Prioridad interna de interfaz

#### UI-P0

- estado vacio de la pantalla ✅
- bloque para registrar el endpoint oficial ✅
- tarjeta de estado de conexion ✅
- tarjeta del servidor MCP local ✅

#### UI-P1

- lista de tools expuestas ✅
- boton de copiar endpoint ✅
- mensaje de error reciente ✅
- texto de ayuda de configuracion ✅

#### UI-P2

- mejoras visuales ✅
- detalles adicionales de diagnostico ✅
- indicadores mas ricos de salud ✅

## Tareas funcionales

- registrar un `MCP endpoint` oficial asociado al agente
- levantar un servidor MCP local compatible
- conectar el servidor MCP local al endpoint oficial
- exponer al menos una tool util
- detectar si la conexion esta activa o no
- reflejar el estado en la interfaz

## Tareas de backend

- crear el modulo `mcp-endpoint-registry`
- crear el modulo `mcp-server-runtime`
- implementar configuracion del endpoint oficial
- implementar conexion y reconexion al endpoint
- implementar registro de tools locales
- crear un servicio de estado para:
  - `connected`
  - `disconnected`
  - `error`
  - `last_connected_at`

## Tareas de almacenamiento

- definir la tabla o estructura local para endpoints MCP
- guardar:
  - identificador local
  - endpoint
  - nombre visible
  - ultimo estado
  - ultima conexion
  - ultimo error
  - metadata basica disponible

## Tareas de observabilidad

- registrar en logs cuando se intenta conectar el endpoint
- registrar cuando la conexion fue exitosa
- registrar cuando falla la conexion o el registro de tools
- registrar que tools fueron expuestas

## Tareas de validacion

- probar conexion con un endpoint oficial real de XiaoZhi
- verificar que el estado cambie correctamente en la interfaz
- verificar que al menos una tool quede visible y usable
- verificar comportamiento de reconexion y error

## Riesgos de esta historia

- el flujo oficial del endpoint puede cambiar con el tiempo
- la integracion depende de servicios que no controlamos por completo
- si el servidor MCP local no refleja bien estado y logs, luego costara diagnosticar problemas

## Resultado esperado al cerrar esta historia

El usuario ya puede tomar su agente oficial de XiaoZhi, obtener su `Custom Services MCP endpoint`, conectarlo a nuestra app local y ver desde la interfaz que el servidor MCP local esta funcionando y exponiendo tools.
