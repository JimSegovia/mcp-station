# Investigacion de XiaoZhi para este proyecto

## Objetivo de esta investigacion

Revisar el repositorio `78/xiaozhi-esp32` y sus enlaces mas importantes para anotar lo que puede servir al objetivo del proyecto:

- que el ESP32 con XiaoZhi se conecte a un servidor propio
- que ese servidor corra en mi computadora
- que la voz termine convirtiendose en acciones sobre mi PC
- que OpenCode o una capa similar pueda ser usada como puente para manipular la computadora

## Repositorio principal revisado

- `https://github.com/78/xiaozhi-esp32`

## Hallazgos mas importantes del repo principal

- XiaoZhi ya esta pensado como un chatbot basado en MCP.
- El firmware soporta dos formas de comunicacion con el backend:
  - `WebSocket`
  - `MQTT + UDP`
- El propio README menciona que del lado cloud MCP puede extender capacidades como:
  - control de hogar inteligente
  - operacion de escritorio de PC
  - busqueda de conocimiento
  - correo
- Eso significa que la idea de controlar una computadora desde XiaoZhi no va contra la arquitectura del proyecto; al contrario, encaja bastante bien.

## Lo mas util del Developer Documentation

### 1. MCP Protocol IoT Control Usage

Documento revisado:

- `https://github.com/78/xiaozhi-esp32/blob/main/docs/mcp-usage.md`

Lo importante:

- El backend actua como cliente MCP.
- El ESP32 actua como servidor MCP.
- Flujo base:
  1. el dispositivo se conecta al backend
  2. el backend envia `initialize`
  3. el backend pide `tools/list`
  4. el backend ejecuta acciones con `tools/call`
- El dispositivo expone herramientas con nombres y esquemas.
- Hay dos tipos de tools:
  - `AddTool`: visibles para el backend y utilizables por la IA
  - `AddUserOnlyTool`: ocultas por defecto, para acciones mas sensibles o manuales

Utilidad para este proyecto:

- Aqui esta la pieza conceptual clave: el ESP32 no tiene que controlar directamente tu PC.
- Lo correcto es que el ESP32 hable con tu backend.
- Luego tu backend decide cuando llamar herramientas de "control de PC".
- En tu caso, esas herramientas podrian ser algo como:
  - abrir programas
  - escribir texto
  - mover archivos
  - ejecutar comandos
  - llamar a OpenCode

### 2. MCP Protocol Interaction Flow

Documento revisado:

- `https://github.com/78/xiaozhi-esp32/blob/main/docs/mcp-protocol.md`

Lo importante:

- Los mensajes MCP viajan envueltos dentro de WebSocket o MQTT.
- El `payload` interno usa `JSON-RPC 2.0`.
- El dispositivo anuncia soporte MCP desde el mensaje `hello` con `features.mcp = true`.
- El backend puede:
  - inicializar sesion MCP
  - descubrir tools
  - llamar tools
- Tambien existen notificaciones iniciadas por el dispositivo.

Utilidad para este proyecto:

- Si hacemos un backend propio, no hace falta inventar otro protocolo.
- Ya existe una ruta limpia:
  - voz en ESP32
  - backend propio
  - tools o acciones del lado servidor
  - control de la PC

### 3. WebSocket Communication Protocol

Documento revisado:

- `https://github.com/78/xiaozhi-esp32/blob/main/docs/websocket.md`

Lo importante:

- Es la opcion mas simple para conectar el dispositivo con un backend propio.
- El ESP32 abre canal con `OpenAudioChannel()`.
- Envia un `hello` con:
  - `type`
  - `version`
  - `features`
  - `transport`
  - `audio_params`
- El servidor responde con otro `hello`.
- Luego viajan dos tipos de trafico:
  - audio binario Opus
  - mensajes JSON como `stt`, `tts`, `llm`, `mcp`, `system`, `alert`
- El dispositivo manda headers utiles:
  - `Authorization`
  - `Protocol-Version`
  - `Device-Id`
  - `Client-Id`

Utilidad para este proyecto:

- Para una primera version local, `WebSocket` parece la mejor opcion.
- Es mas amigable que `MQTT + UDP`.
- Es mas facil de depurar y de montar en una computadora personal.
- Si la meta inicial es "hacer que mi XiaoZhi hable con mi PC y esta ejecute cosas", WebSocket es el camino mas razonable.

### 4. MQTT + UDP Hybrid Protocol

Documento revisado:

- `https://github.com/78/xiaozhi-esp32/blob/main/docs/mqtt-udp.md`

Lo importante:

- MQTT se usa para control.
- UDP se usa para audio en tiempo real.
- El audio UDP va cifrado con AES-CTR.
- Tiene menor latencia, pero complica bastante mas la infraestructura.
- Requiere broker MQTT, servidor UDP, manejo de claves, puertos, firewall y red.

Utilidad para este proyecto:

- No parece la mejor opcion para empezar.
- Solo valdria la pena si luego quieres mas rendimiento, mas concurrencia o una arquitectura mas cercana a produccion.
- Para un prototipo personal de control del PC, WebSocket gana claramente por simplicidad.

## Related Open Source Projects revisados

El README principal recomienda estos backends para correr el servidor en computadoras personales:

- `xinnan-tech/xiaozhi-esp32-server`
- `joey-zhou/xiaozhi-esp32-server-java`
- `AnimeAIChat/xiaozhi-server-go`
- `hackers365/xiaozhi-esp32-server-golang`

## Analisis de los servidores

### 1. xinnan-tech/xiaozhi-esp32-server

Repo revisado:

- `https://github.com/xinnan-tech/xiaozhi-esp32-server`

Documentos revisados:

- `https://github.com/xinnan-tech/xiaozhi-esp32-server/blob/main/README.md`
- `https://github.com/xinnan-tech/xiaozhi-esp32-server/blob/main/docs/Deployment.md`
- `https://github.com/xinnan-tech/xiaozhi-esp32-server/blob/main/docs/Deployment_all.md`

Lo mas util:

- Tiene modo simplificado y modo completo.
- El modo simplificado corre solo server y evita base de datos compleja.
- El despliegue puede hacerse con Docker o desde codigo fuente.
- La documentacion aterriza claramente las dos URLs clave:
  - WebSocket
  - OTA
- En despliegue local recalcan que hay que usar la IP LAN de la computadora.
- Mencionan explicitamente que desde el backend se pueden bajar comandos MCP al ESP32.
- Tambien tienen guias para:
  - Home Assistant
  - vision
  - voiceprint
  - endpoint MCP

Lo mas importante para tu proyecto:

- Es probablemente la referencia mas practica para levantar un backend personal rapido.
- Te da un camino ya probado para conectar el ESP32 a una PC usando WebSocket local.
- Si buscas "tener algo funcionando pronto", este repo parece muy buen candidato.

Riesgos o costo:

- La documentacion y el stack se sienten mas grandes de lo que necesitas para una primera version.
- Tiene muchas piezas extras que no son necesarias si solo quieres controlar tu propia computadora.

### 2. joey-zhou/xiaozhi-esp32-server-java

Repo revisado:

- `https://github.com/joey-zhou/xiaozhi-esp32-server-java`

Documento revisado:

- `https://github.com/joey-zhou/xiaozhi-esp32-server-java/blob/main/README.md`

Lo mas util:

- Es una opcion mas empresarial.
- Usa arquitectura de dos procesos:
  - `xiaozhi-server`
  - `xiaozhi-dialogue`
- Usa Spring Boot, Vue, MySQL y Redis.
- Tiene backend visual, gestion de usuarios, OTA, WebSocket y MQTT.
- Habla de escalado horizontal, monitoreo y pruebas de concurrencia.

Lo mas importante para tu proyecto:

- Sirve si mas adelante quieres una plataforma mas robusta y administrable.
- Para un proyecto personal de "voz -> servidor local -> controlar mi PC", parece demasiado pesado para arrancar.

Conclusion sobre esta opcion:

- Potente, pero no la mas conveniente como primer paso.

### 3. AnimeAIChat/xiaozhi-server-go

Repo revisado:

- `https://github.com/AnimeAIChat/xiaozhi-server-go`

Documento revisado:

- `https://github.com/AnimeAIChat/xiaozhi-server-go/blob/main/README.md`

Lo mas util:

- Tiene binarios listos para Windows.
- Indica configuracion directa de:
  - `web.websocket`
  - OTA
- Dice explicitamente que para pruebas en LAN debes usar la IP local de tu computadora.
- Soporta despliegue en una sola maquina.
- Usa sqlite en la version comunitaria.
- Soporta Docker.
- Tiene foco en `websocket` y en clientes ESP32, Python y Android.

Lo mas importante para tu proyecto:

- Esta opcion es especialmente interesante si quieres correr algo en Windows sin tanto dolor.
- Que tenga binario Windows precompilado es una ventaja muy fuerte para tu caso.
- Puede ser una base realista para prototipar rapido.

Observacion:

- El README habla tambien de version comercial y varias funciones extra. Conviene separar lo que es open source puro de lo que depende de capacidades no abiertas.

### 4. hackers365/xiaozhi-esp32-server-golang

Repo revisado:

- `https://github.com/hackers365/xiaozhi-esp32-server-golang`

Documento revisado:

- `https://github.com/hackers365/xiaozhi-esp32-server-golang/blob/main/README.md`

Lo mas util:

- Es un backend Go muy grande y bastante moderno.
- Soporta WebSocket y MQTT + UDP.
- Tiene consola web completa.
- Tiene docs separadas para:
  - WebSocket y OTA
  - MQTT + UDP
  - MCP
  - llamadas remotas MCP
  - OpenClaw
  - knowledge base
- Tambien ofrece despliegue rapido con paquetes listos.

Lo mas importante para tu proyecto:

- Es una base poderosa si quieres crecer mas adelante.
- Puede servir mucho como referencia de arquitectura para la capa "backend que interpreta y acciona herramientas".

Conclusiones sobre esta opcion:

- Muy interesante tecnicamente.
- Pero para arrancar puede ser mas de lo necesario.

## Conclusion tecnica para este proyecto

### Lo que ya se confirma

- XiaoZhi si puede conectarse a un backend propio.
- Ese backend puede correr en una computadora personal.
- El flujo oficial si contempla MCP del lado backend.
- La idea de "voz en el ESP32" y "acciones en la PC" es totalmente compatible con la arquitectura del ecosistema.

### La arquitectura que mejor encaja con tu idea

1. El ESP32 XiaoZhi se conecta por `WebSocket` a un backend propio corriendo en tu PC.
2. Ese backend procesa audio, STT, LLM y TTS.
3. Cuando detecta una intencion tipo "abre VS Code", "ejecuta esto", "usa OpenCode", no manda eso al ESP32.
4. En lugar de eso, ejecuta una herramienta del lado servidor para controlar la computadora.
5. Esa herramienta del lado servidor puede:
   - ejecutar comandos locales
   - abrir programas
   - automatizar teclado o mouse
   - hablar con OpenCode
   - exponer una capa de seguridad o aprobacion

### Lo mas recomendable para una primera fase

- usar `WebSocket`, no `MQTT + UDP`
- correr el backend en tu misma PC
- empezar con una sola accion simple del lado servidor, por ejemplo:
  - abrir una app
  - crear un archivo
  - ejecutar un comando controlado
- dejar OpenCode como segunda capa, no como primer paso

## Recomendacion inicial de backends para probar

### Opcion mas pragmatica

- `AnimeAIChat/xiaozhi-server-go`

Motivo:

- tiene foco en despliegue simple
- tiene binarios para Windows
- soporta WebSocket
- parece amigable para pruebas locales

### Opcion mas documentada para entender el flujo completo

- `xinnan-tech/xiaozhi-esp32-server`

Motivo:

- la documentacion de despliegue en computadora personal es bastante clara
- aterriza bien WebSocket, OTA e IP LAN

### Opcion mas robusta para crecer

- `hackers365/xiaozhi-esp32-server-golang`

Motivo:

- arquitectura mas amplia
- muchas extensiones y documentacion modular

## Implicaciones directas para el control de la PC

- No parece necesario modificar mucho del firmware al inicio.
- Lo critico esta del lado del backend.
- La pieza que realmente tendras que diseñar es una capa de "tools del servidor" para la computadora.
- En otras palabras:
  - XiaoZhi no necesita "controlar Windows" por si mismo
  - el backend debe traducir voz/intencion a acciones locales seguras

## Riesgos y advertencias utiles

- Controlar una PC por voz desde un backend local implica riesgo serio si se dejan acciones abiertas sin validacion.
- Si en algun punto expones el backend a internet, el riesgo sube mucho.
- Conviene pensar desde el inicio en:
  - lista blanca de comandos
  - confirmaciones para acciones peligrosas
  - logs
  - separacion entre acciones de lectura y acciones destructivas

## Mi conclusion mas concreta

- La idea es viable.
- El camino mas limpio es:
  - XiaoZhi -> WebSocket -> backend local en tu PC -> tools del servidor -> acciones en tu computadora
- Para empezar, la mejor estrategia parece ser montar un backend personal simple y luego agregar una capa propia para integrar OpenCode o control local del sistema.

## Fuentes revisadas

- `https://github.com/78/xiaozhi-esp32`
- `https://github.com/78/xiaozhi-esp32/blob/main/docs/mcp-usage.md`
- `https://github.com/78/xiaozhi-esp32/blob/main/docs/mcp-protocol.md`
- `https://github.com/78/xiaozhi-esp32/blob/main/docs/websocket.md`
- `https://github.com/78/xiaozhi-esp32/blob/main/docs/mqtt-udp.md`
- `https://github.com/xinnan-tech/xiaozhi-esp32-server`
- `https://github.com/xinnan-tech/xiaozhi-esp32-server/blob/main/docs/Deployment.md`
- `https://github.com/xinnan-tech/xiaozhi-esp32-server/blob/main/docs/Deployment_all.md`
- `https://github.com/joey-zhou/xiaozhi-esp32-server-java`
- `https://github.com/AnimeAIChat/xiaozhi-server-go`
- `https://github.com/hackers365/xiaozhi-esp32-server-golang`
