# Arquitectura propuesta

## Perfil de usuario

### Usuario principal

Persona que ya tiene un dispositivo XiaoZhi o quiere usar uno como interfaz fisica de voz para su computadora, y busca una forma mas poderosa, local y administrable de convertir voz en acciones reales sobre su PC.

### Tipos de usuario que mas podrian aprovecharlo

- usuarios de XiaoZhi que quieren hacer mas que solo conversar
- makers que quieren un sistema local extensible
- developers que quieren conectar su agente a herramientas reales
- usuarios que valoran privacidad y control local
- personas que no quieren pagar varias APIs separadas
- usuarios que quieren una interfaz clara para activar o desactivar capacidades del agente

### Problemas que este producto resolveria

- demasiadas piezas sueltas para conectar voz, agente, herramientas y PC
- dependencia de servicios externos o multiples APIs
- falta de una interfaz clara para administrar MCPs y modulos
- dificultad para hacer crecer el sistema sin rehacerlo
- poca seguridad o poco control cuando se automatiza una computadora

## Propuesta de valor

Este producto convertiria a XiaoZhi en la puerta de entrada por voz a un agente local de PC, administrable desde una interfaz clara, modular y extensible.

La propuesta de valor principal seria:

- conectar un XiaoZhi a una PC de forma ordenada
- administrar el agente local desde una sola interfaz
- gestionar modulos y servidores MCP como piezas enchufables
- usar OpenCode como motor central de ejecucion
- escalar el sistema sin depender de una arquitectura improvisada
- mantener el control local y reducir dependencia de multiples APIs

### Diferenciador principal

No seria solo "un backend para XiaoZhi" ni solo "un agente local", sino una plataforma local de agente personal donde XiaoZhi actua como interfaz fisica de voz y el usuario administra todo desde una consola propia.

## MVP real

El MVP real no deberia intentar resolverlo todo.

La v1 deberia demostrar que el producto ya sirve de verdad, no solo que tecnicamente funciona.

### Nota importante sobre la estrategia de v1

Despues de revisar mejor la consola oficial de XiaoZhi, existe una alternativa muy conveniente para acelerar la primera version:

- dejar que `xiaozhi.me` siga resolviendo voz, conversacion, memoria y speaker recognition
- dejar que el dispositivo siga conectado al ecosistema oficial
- hacer que tu app se enfoque primero en montar y administrar un servidor MCP local para `Custom Services`

Eso significa que la primera version no necesariamente tiene que empezar como un backend completo compatible con XiaoZhi.

Tambien puede empezar como una consola local especializada en:

- desplegar el MCP server
- registrar tools y modulos
- conectar `OpenCode` y otras capacidades locales
- mostrar logs, salud y estado de la integracion

Esta decision no invalida la vision grande del producto.

Solo propone una `v1 acelerada` para validar valor real mas rapido.

### Lo minimo que la v1 debe resolver bien

1. Vincular un dispositivo XiaoZhi a una computadora local.
2. Exponer un endpoint WebSocket funcional y estable.
3. Permitir una conversacion basica con el runtime.
4. Ejecutar al menos algunas acciones reales sobre la PC.
5. Tener una interfaz local para administrar configuracion basica.
6. Mostrar el estado del dispositivo, los modulos y las ejecuciones.
7. Tener controles minimos de seguridad.

### Funciones que deberia tener la v1

- vinculacion de un XiaoZhi
- configuracion local de WebSocket y OTA
- panel basico de estado del dispositivo
- modulo de OpenCode conectado
- modulo de acciones locales basicas
- registro de modulos disponibles
- activacion y desactivacion de modulos
- registro de ejecuciones
- confirmacion o bloqueo para acciones sensibles
- logs de sesion y errores

### Variante recomendada para una `v1 acelerada`

Si elegimos aprovechar la infraestructura oficial de XiaoZhi en esta primera etapa, la `v1` cambiaria a algo mas realista:

1. Permitir que el usuario obtenga y entienda su `MCP endpoint` oficial de `Custom Services`.
2. Ayudarle a levantar un servidor MCP local compatible.
3. Registrar y administrar tools o modulos locales desde una interfaz propia.
4. Conectar `OpenCode` como capability principal dentro de ese servidor MCP.
5. Mostrar estado del servidor, tools cargadas, errores, logs y salud.
6. Facilitar pruebas y debugging del endpoint.

En esa variante, la `v1` ya seria util sin rehacer:

- la voz
- el runtime conversacional
- la memoria
- `Speaker Recognition`

### Casos de uso concretos que la v1 deberia soportar

- "abre VS Code"
- "crea un archivo de notas"
- "ejecuta una tarea en OpenCode"
- "muestrame que modulos tengo conectados"
- "desactiva temporalmente cierto modulo"

### Que no deberia intentar resolver la v1

- multiples protocolos complejos desde el inicio
- soporte completo para muchos dispositivos al mismo tiempo
- automatizacion total del escritorio sin limites
- marketplace complejo de plugins en el primer corte
- demasiados proveedores externos
- demasiada configuracion avanzada antes de validar la experiencia base

## Advertencias importantes

### Advertencia de producto

La clave del exito no sera solo que la voz funcione ni que OpenCode responda bien.

La clave real sera la experiencia de administracion.

Si vincular el dispositivo, activar modulos, conectar MCPs y definir permisos es sencillo, el producto tendra mucho valor.

Si esa experiencia es confusa, el sistema puede sentirse potente pero dificil de usar.

### Advertencia de alcance

Existe riesgo de querer construir demasiado desde la primera version.

Conviene priorizar:

- un solo dispositivo
- un solo transporte principal
- pocos modulos bien hechos
- una interfaz simple pero clara

### Advertencia tecnica

Controlar una PC con voz siempre tiene riesgo.

Si no se diseña bien la capa de permisos, confirmaciones y logs, el sistema puede volverse fragil o peligroso.

### Advertencia arquitectonica

No conviene convertir OpenCode en toda la plataforma.

OpenCode debe ser una pieza central de ejecucion, pero no debe absorber:

- transporte
- UI
- configuracion global
- politicas de seguridad

### Advertencia estrategica

No conviene copiar un solo repo completo.

Lo mas sano es combinar referencias y construir una base propia pensada especificamente para este producto.

## Objetivo de esta arquitectura

Definir una arquitectura base para construir el proyecto bien desde el inicio:

- local-first
- modular
- escalable
- facil de operar en una PC personal
- preparada para crecer a multiples modulos, MCPs y dispositivos

Esta propuesta toma inspiracion de lo mejor visto en los repos analizados en [INVESTIGACION_XIAOZHI.md](D:\Proyects\Code\XiaonZhi\Server\INVESTIGACION_XIAOZHI.md):

- de `xinnan-tech/xiaozhi-esp32-server`: claridad de despliegue local, WebSocket y OTA
- de `joey-zhou/xiaozhi-esp32-server-java`: separacion entre plano de administracion y plano de dialogo
- de `AnimeAIChat/xiaozhi-server-go`: enfoque pragmatico para correr bien en una sola PC
- de `hackers365/xiaozhi-esp32-server-golang`: modularidad, consola de gestion y capa MCP mas madura

## Principios de diseno

- `local-first`: todo debe poder correr en tu propia computadora
- `WebSocket primero`: la primera version debe usar WebSocket, no MQTT + UDP
- `modulos desacoplados`: cada capacidad importante debe vivir como modulo separado
- `MCP como capa nativa`: los conectores y tools deben pensarse desde el inicio como piezas administrables
- `OpenCode como modulo central de ejecucion`: no como todo el sistema, sino como el motor principal para tareas inteligentes sobre la PC
- `seguridad por defecto`: nada de acciones peligrosas sin control
- `escalado progresivo`: empezar simple, pero sin bloquear crecimiento futuro

## Arquitectura recomendada

```text
XiaoZhi Device
    ->
Transport Gateway
    ->
Session and Conversation Runtime
    ->
Intent and Orchestration Layer
    ->
Module and MCP Manager
    ->
Execution Layer
    ->
OpenCode / Local Tools / PC Actions

Admin Web UI
    ->
Control Plane API
    ->
Config, Modules, Devices, MCP Registry, Policies
```

## Capas del sistema

### 1. Transport Gateway

Responsabilidad:

- aceptar conexiones desde XiaoZhi
- manejar handshake `hello`
- exponer endpoint WebSocket
- recibir audio y mensajes JSON
- devolver TTS, estados y respuestas

Decision recomendada:

- usar `WebSocket` como transporte principal en v1
- dejar `MQTT + UDP` como fase futura

Motivo:

- mas simple de depurar
- mas facil de correr en Windows
- mejor para una primera version local

### 2. Session and Conversation Runtime

Responsabilidad:

- mantener sesiones activas por dispositivo
- administrar estado de la conversacion
- coordinar eventos como:
  - escucha
  - STT
  - respuesta
  - TTS
  - tool calls
- aislar cada sesion para que el sistema pueda crecer a varios dispositivos

Esta capa deberia conocer:

- `device_id`
- `client_id`
- `session_id`
- modo de conversacion
- contexto corto de la sesion

### 3. Intent and Orchestration Layer

Responsabilidad:

- interpretar la intencion del usuario
- decidir si una orden:
  - solo responde en texto o voz
  - llama a OpenCode
  - ejecuta una tool local
  - consulta un MCP externo
- aplicar reglas y politicas antes de ejecutar

Esta es la capa cerebro del sistema.

Aqui no conviene meter logica de UI ni logica de drivers.

Su trabajo es decidir.

### 4. Module and MCP Manager

Responsabilidad:

- registrar modulos internos del sistema
- registrar MCP servers conectados
- activar o desactivar modulos
- administrar configuracion por modulo
- ofrecer una vista unificada para que la interfaz web los gestione

Esta capa es clave para el producto final que quieres.

Debe permitir administrar cosas como:

- modulo de OpenCode
- modulo de archivos locales
- modulo de automatizacion del escritorio
- modulo de terminal
- modulo de navegador
- modulo de calendario o correo si luego agregas mas
- MCPs externos o locales

### 5. Execution Layer

Responsabilidad:

- ejecutar acciones reales en la computadora
- traducir decisiones del orquestador a operaciones concretas

Esta capa debe estar separada del orquestador porque ejecutar no es lo mismo que decidir.

Submodulos recomendados:

- `opencode-adapter`
- `local-command-runner`
- `filesystem-tools`
- `desktop-automation`
- `browser-automation`
- `mcp-client-adapter`

### 6. Control Plane API

Responsabilidad:

- servir la interfaz de administracion
- guardar configuraciones
- gestionar dispositivos, modulos, endpoints y permisos
- exponer operaciones administrativas

Esta capa es distinta del runtime de conversacion.

Inspiracion principal:

- separacion estilo `joey-zhou/xiaozhi-esp32-server-java`

No hace falta arrancar con dos procesos separados obligatoriamente, pero si con dos dominios logicos claros:

- `runtime`
- `control-plane`

## Interfaz principal del producto

El producto final deberia tener una interfaz web local desde donde puedas:

- vincular tu XiaoZhi
- ver el estado de conexion
- configurar la URL WebSocket y OTA
- administrar modulos
- administrar MCPs
- definir herramientas permitidas
- ver logs y trazas de ejecucion
- probar tools manualmente
- definir reglas de seguridad
- configurar comportamiento del agente

## Modulos base recomendados

### Modulo 1: Device Manager

Funciones:

- registrar dispositivos
- mostrar estado
- almacenar identificadores
- vincular configuraciones por dispositivo

### Modulo 2: Conversation Runtime

Funciones:

- sesiones
- contexto corto
- canal de mensajes
- control de respuesta

### Modulo 3: OpenCode Adapter

Funciones:

- enviar tareas a OpenCode
- recibir resultados
- convertir tareas del agente en operaciones concretas

Este modulo debe ser central para tu idea.

Pero no debe mezclarse con transporte ni UI.

### Modulo 4: MCP Registry

Funciones:

- alta, baja y edicion de MCPs
- estado de conexion
- categorias
- permisos
- mapeo a herramientas del agente

### Modulo 5: Local Action Tools

Funciones:

- abrir programas
- ejecutar comandos controlados
- crear o modificar archivos
- inspeccionar informacion local

### Modulo 6: Policy and Permissions

Funciones:

- lista blanca de acciones
- confirmacion para acciones peligrosas
- restricciones por modulo
- restricciones por dispositivo

### Modulo 7: Logs and Observability

Funciones:

- historial de sesiones
- historial de tools llamadas
- errores
- auditoria de acciones sobre la PC

## Separacion recomendada de procesos

### Opcion recomendada para v1

Un solo backend, pero dividido internamente en dominios claros:

- `transport`
- `runtime`
- `orchestrator`
- `modules`
- `admin-api`
- `storage`

Ventaja:

- mas simple para arrancar
- menos complejidad operativa
- suficiente para una sola PC

### Opcion recomendada para v2 o v3

Separar en procesos:

- `xiaonzhi-runtime`
- `xiaonzhi-admin`
- `xiaonzhi-worker`

Donde:

- `runtime` atiende WebSocket y sesiones
- `admin` sirve la interfaz de configuracion
- `worker` ejecuta acciones pesadas o sensibles

Ventaja:

- mejor aislamiento
- mejor escalado
- mas seguridad

## Persistencia recomendada

### Fase inicial

- `SQLite` para configuracion local

Guardar:

- dispositivos
- modulos
- MCP registry
- settings
- logs
- policies

Motivo:

- simple
- local
- sin dependencias pesadas

### Fase futura

- abstraccion de repositorios para poder migrar a `PostgreSQL` sin reescribir logica

## Comunicacion interna recomendada

Dentro del backend, usar eventos internos simples para desacoplar:

- `device.connected`
- `session.started`
- `speech.transcribed`
- `intent.detected`
- `tool.requested`
- `tool.executed`
- `tts.requested`
- `action.blocked`

Esto ayuda mucho a escalar sin acoplar todo.

## Integracion con OpenCode

OpenCode no deberia reemplazar toda la arquitectura.

Deberia entrar como un modulo central dentro de la capa de ejecucion.

### Rol correcto de OpenCode

- resolver tareas inteligentes
- operar sobre codigo, terminal y archivos
- ayudar a ejecutar acciones complejas

### Rol incorrecto de OpenCode

- manejar WebSocket del dispositivo
- servir la interfaz web completa
- guardar toda la configuracion del sistema
- actuar como unica capa de seguridad

## Seguridad recomendada

Esto hay que hacerlo bien desde el principio.

Medidas minimas:

- lista blanca de acciones
- separacion entre lectura y escritura
- confirmacion para acciones peligrosas
- logs de auditoria
- aislamiento de ejecucion
- timeouts
- modulos desactivables

Ejemplos de acciones peligrosas:

- borrar archivos
- ejecutar scripts arbitrarios
- cerrar procesos
- modificar configuraciones del sistema

## Referencias de producto y administracion

### MCP Router como referencia parcial

El repo `mcp-router/mcp-router` no deberia ser la base completa del producto, pero si es una referencia fuerte para disenar la capa de administracion de `MCPs`.

Lo mas util que nos aporta es:

- panel unificado para gestionar servidores MCP
- activacion y desactivacion por servidor y por tool
- organizacion por proyectos o contextos
- logs y observabilidad desde la interfaz
- enfoque local-first para configuracion y control

Donde debe influir mas:

- `Module and MCP Manager`
- panel de administracion
- permisos, estado y visibilidad de tools

Donde no debe marcar la arquitectura principal:

- vinculacion de `XiaoZhi`
- transporte `WebSocket`
- runtime conversacional
- orquestacion central

Referencia ampliada:

- ver [INVESTIGACION_MCP_ROUTER.md](D:\Proyects\Code\XiaonZhi\Server\INVESTIGACION_MCP_ROUTER.md)

## Stack sugerido

Sin casarnos todavia con implementacion final, la opcion que mejor encaja para este producto parece:

### Backend

- `Go`

Motivos:

- buen rendimiento para runtime y WebSocket
- binarios faciles de distribuir
- buena opcion para Windows
- equilibrio entre simplicidad y capacidad de crecimiento
- buena compatibilidad con una arquitectura modular tipo los repos Go analizados

### Frontend

- `React` + una UI simple de administracion

Motivos:

- facil construir panel modular
- buena experiencia para gestionar modulos y MCPs

### Base de datos inicial

- `SQLite`

### Transporte

- `WebSocket`

## Ruta de construccion recomendada

### Fase 1

- WebSocket local funcionando
- XiaoZhi conectado a la PC
- una interfaz minima
- un modulo local simple
- una integracion basica con OpenCode

### Fase 2

- registry de MCPs
- panel modular real
- policies y permisos
- logs y auditoria

### Fase 3

- workers separados
- mas aislamiento
- mas herramientas
- opcion de multiples dispositivos

## Decision arquitectonica principal

La mejor arquitectura para este proyecto, pensando en hacerlo bien desde el inicio, no es copiar un repo entero.

Lo correcto es combinar lo mejor de varios enfoques:

- simplicidad local de los repos de despliegue rapido
- separacion de responsabilidades del repo Java
- modularidad y consola del repo Go mas grande
- protocolo oficial de XiaoZhi como base no negociable

## Conclusiones

- XiaoZhi debe ser el cliente de voz.
- Tu software debe ser el centro de control.
- OpenCode debe ser un modulo central de ejecucion, no toda la plataforma.
- MCP debe ser una capacidad administrable por interfaz.
- La primera version debe priorizar buena arquitectura y simplicidad operativa.

## Arquitectura resumida en una frase

Construir un sistema local, modular y administrable donde XiaoZhi sea la entrada por voz, un backend WebSocket maneje las sesiones, un orquestador decida acciones, un gestor de modulos administre MCPs y OpenCode actue como motor principal para operar la PC.
