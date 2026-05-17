# Investigacion sobre la consola web oficial de XiaoZhi

## Objetivo de esta investigacion

Entender, hasta donde las fuentes publicas lo permiten, como funciona la consola web de `xiaozhi.me` para:

- agregar un dispositivo
- reconocer que un dispositivo pertenece a un usuario o agente
- detectar si el dispositivo esta conectado o no
- identificar que partes de ese flujo podriamos reutilizar en nuestro software

## Aclaracion importante

No tenemos el codigo fuente de la consola oficial de `xiaozhi.me`.

Por eso este documento mezcla:

- hallazgos confirmados por documentacion oficial o repos publicos
- inferencias razonables basadas en servidores open source compatibles

Cuando algo sea una inferencia, lo voy a marcar como tal.

## Lo que si esta confirmado

### 1. El dispositivo primero se conecta a internet y luego se agrega a la consola

La documentacion oficial indica que, despues de configurar `Wi-Fi` o `4G`, el usuario debe ir a la consola de `xiaozhi.me` para agregar el dispositivo. El flujo oficial menciona un codigo de verificacion de 6 digitos que el dispositivo anuncia por voz o por pantalla.

Fuente:

- [Configuracion Wi-Fi y agregar dispositivos](https://home.xiaozhi.me/xz-docs/docs/tutorial-basics/configure-wifi-and-add-devices/)

### 2. El alta del dispositivo usa un codigo de verificacion de 6 digitos

La documentacion oficial describe este flujo:

- el dispositivo ya tiene conectividad
- el usuario despierta el dispositivo
- el dispositivo anuncia un codigo de 6 digitos
- el usuario entra a la consola
- el usuario crea o elige un agente
- el usuario pega ese codigo en `Add Device`

Eso coincide muy bien con tus capturas.

Fuente:

- [Configuracion Wi-Fi y agregar dispositivos](https://home.xiaozhi.me/xz-docs/docs/tutorial-basics/configure-wifi-and-add-devices/)

### 3. El dispositivo tiene identificadores propios antes de ser agregado

La documentacion del protocolo WebSocket confirma que, al abrir el canal, el dispositivo envia headers como:

- `Authorization`
- `Protocol-Version`
- `Device-Id`
- `Client-Id`

Y tambien envia un mensaje `hello`.

Esto indica que el backend puede reconocer tecnicamente al dispositivo antes incluso de que quede vinculado a una cuenta de usuario.

Fuente:

- [WebSocket protocol del repo oficial](https://github.com/78/xiaozhi-esp32/blob/main/docs/websocket.md)

### 3.1 La consola web real usa rutas privadas autenticadas del panel

Al revisar los HTML exportados de la consola y los bundles cargados por esa pagina, aparecen rutas y comportamiento del frontend real:

- `https://xiaozhi.me/console/agents`
- `https://xiaozhi.me/console/agents/{agentId}/devices`
- `https://xiaozhi.me/console/agents/{agentId}/chats`
- `https://xiaozhi.me/console/agents/{agentId}/config`
- `https://xiaozhi.me/console/agents/{agentId}/speakers`

Tambien aparece un wrapper de requests que:

- toma `token` desde `localStorage`
- envia `Authorization: Bearer <token>`
- redirige a login si recibe `401`

Esto confirma que la consola usa una API privada autenticada del panel, no una interfaz publica pensada para terceros.

Fuentes:

- [Console.html](C:\Users\JimXL\Downloads\Console.html)
- [Agent Management - Xiaozhi AI Official.html](C:\Users\JimXL\Downloads\Agent Management - Xiaozhi AI Official.html)

### 3.2 La pantalla `Configure Role` ya integra `MCP Settings` y `Speaker Recognition`

En los HTML exportados y en la configuracion visible del agente se ve que la consola oficial ya ofrece, dentro del mismo flujo:

- configuracion de voz
- configuracion del modelo
- memoria
- `MCP Settings`
- `Speaker Recognition`

Eso importa mucho para este proyecto porque confirma que la experiencia base de conversacion ya existe en el ecosistema oficial.

Si tu objetivo inmediato es ahorrar tiempo, no parece necesario rehacer desde cero en `v1`:

- el flujo de hablarle al dispositivo
- la respuesta de voz
- la configuracion base del agente
- el reconocimiento de speaker

En cambio, tu app podria concentrarse primero en:

- montar un servidor MCP local
- conectarlo al `Custom Services` endpoint del agente
- administrar tools, modulos y estado desde una interfaz propia

Fuentes:

- [Console (1).html](C:\Users\JimXL\Downloads\Console (1).html)
- [Console (2).html](C:\Users\JimXL\Downloads\Console (2).html)

### 3.3 El flujo real de `Custom Services` usa un endpoint MCP oficial separado

El bundle `agent-config.pBndCyaq.js` deja ver el flujo tecnico del boton `Get MCP Endpoint`.

Hay dos pasos claros:

1. El panel pide un token temporal al backend principal del agente:
   - `POST /agents/{agentId}/generate-mcp-endpoint-token`
2. Con ese token, la UI construye la URL final:
   - `wss://api.xiaozhi.me/mcp/?token=...`

Despues, la consola consulta el estado del endpoint en un servicio MCP centralizado:

- `GET https://api.xiaozhi.me/mcp/endpoints/list?endpoint_ids=agent_{agentId}`

Tambien aparece una llamada para obtener los servicios oficiales activables del agente:

- `GET /agents/common-mcp-tool/list?agentId={agentId}`

Esto confirma varias cosas:

- el `MCP endpoint` no cuelga directamente del dispositivo
- el endpoint pertenece al agente dentro del ecosistema oficial
- XiaoZhi expone un WebSocket MCP oficial reutilizable para `Custom Services`
- la consola ya sabe mostrar `status` y `tools` de ese endpoint

Fuentes:

- [Console (1).html](C:\Users\JimXL\Downloads\Console (1).html)
- [agent-config.pBndCyaq.js](https://xiaozhi.me/assets/agent-config.pBndCyaq.js)

## Hallazgo clave sobre el flujo real de reconocimiento

### 4. El paso previo a la vinculacion parece ocurrir en el endpoint de OTA

En servidores compatibles open source aparece un patron muy importante:

- el dispositivo consulta un endpoint `OTA`
- si el dispositivo aun no esta vinculado, el backend devuelve un bloque `activation`
- ese bloque contiene:
  - `code`
  - `message`
  - `challenge`
- si el dispositivo ya esta vinculado, el backend devuelve el `websocket.url`

Esto es muy valioso porque nos muestra que el flujo de "agregar dispositivo" no nace en la interfaz web, sino en la comunicacion previa del dispositivo con el backend.

Fuentes:

- [Issue con ejemplo de respuesta OTA](https://github.com/xinnan-tech/xiaozhi-esp32-server/issues/2324)
- [DeviceAppService.java](https://github.com/joey-zhou/xiaozhi-esp32-server-java/blob/main/xiaozhi-server/src/main/java/com/xiaozhi/device/DeviceAppService.java)
- [DeviceServiceImpl.java](https://github.com/xinnan-tech/xiaozhi-esp32-server/blob/main/main/manager-api/src/main/java/xiaozhi/modules/device/service/impl/DeviceServiceImpl.java)

### 4.1 La UI real del panel confirma el alta por codigo de 6 digitos

En el bundle de `devices` se observa el flujo del modal de agregar dispositivo:

- carga la lista con `GET /agents/{agentId}/devices`
- abre modal con campo `verificationCode`
- valida por regex que sea exactamente de 6 digitos
- envia `POST /agents/{agentId}/devices`
- manda `{"verificationCode":"123456"}`

Tambien se ve una accion de desvincular:

- `POST /agents/{agentId}/devices/delete`
- body con `deviceId`

Esto confirma que la consola web no escanea el dispositivo directamente; solo consume el backend del panel para:

- listar dispositivos ya vinculados
- vincular un nuevo dispositivo usando el codigo
- desvincular un dispositivo existente

Fuentes:

- [Console.html](C:\Users\JimXL\Downloads\Console.html)
- [Agent Management - Xiaozhi AI Official.html](C:\Users\JimXL\Downloads\Agent Management - Xiaozhi AI Official.html)

## Flujo probable de la consola oficial

Lo siguiente es una reconstruccion razonable del flujo completo.

### Flujo reconstruido

1. El dispositivo arranca y se conecta a `Wi-Fi` o `4G`.
2. El dispositivo consulta al backend para revisar `OTA` y configuracion.
3. El backend identifica al dispositivo por `Device-Id` o `MAC`.
4. Si el dispositivo no esta vinculado, el backend genera un codigo temporal de activacion.
5. El dispositivo reproduce o muestra ese codigo al usuario.
6. El usuario entra a `xiaozhi.me` y lo introduce en `Add Device`.
7. La consola llama al backend para resolver ese codigo.
8. El backend convierte ese codigo en una vinculacion real entre:
   - dispositivo
   - usuario
   - agente o rol
9. En las siguientes consultas, el backend ya devuelve la direccion `WebSocket` y el token de sesion si aplica.
10. El dispositivo abre el `WebSocket`, hace `hello` y queda listo para conversar.

### Donde esta el punto de reconocimiento real

La consola web probablemente no "descubre" el dispositivo directamente desde el navegador.

La deteccion real parece venir del backend porque:

- el backend ya vio al dispositivo cuando este llamo a `OTA`
- el backend ya conoce su `device id` o `mac`
- el backend guarda un codigo temporal asociado a ese dispositivo
- la consola solo consulta o confirma ese estado

Esta parte es una inferencia fuerte, pero esta muy respaldada por los repos compatibles.

## Como detecta que el dispositivo esta encendido o conectado

### Lo mas probable

Lo mas probable es que la consola no detecte esto por magia ni por una conexion directa navegador-dispositivo.

Lo normal seria:

- el dispositivo se conecta al backend por `WebSocket` o `MQTT`
- el backend mantiene estado de sesion activa
- la interfaz web consulta ese estado al backend

### Evidencia util

En implementaciones compatibles aparecen varias pistas:

- actualizacion de `lastConnectedAt`
- consulta de estado online desde un servicio intermedio
- administracion de dispositivos con monitoreo en tiempo real

Fuentes:

- [DeviceServiceImpl.java](https://github.com/xinnan-tech/xiaozhi-esp32-server/blob/main/main/manager-api/src/main/java/xiaozhi/modules/device/service/impl/DeviceServiceImpl.java)
- [README del server Java](https://github.com/joey-zhou/xiaozhi-esp32-server-java/blob/main/README.md)

### Evidencia directa desde el frontend exportado

En el bundle de `devices` del panel real aparecen campos de dispositivo como:

- `online`
- `last_connected_at`
- `app_version`
- `auto_update`

Y el frontend renderiza comportamientos distintos segun `a.online`:

- si `online` es verdadero, muestra una etiqueta verde de dispositivo en linea
- si `online` es falso, no muestra ese estado positivo
- ciertas acciones de personalizacion solo se habilitan cuando el dispositivo esta en linea
- cuando no esta en linea, aparece el mensaje equivalente a `device not powered on`

Esto refuerza que la consola simplemente pinta un estado que ya viene resuelto desde backend.

Fuentes:

- [Console.html](C:\Users\JimXL\Downloads\Console.html)
- [Agent Management - Xiaozhi AI Official.html](C:\Users\JimXL\Downloads\Agent Management - Xiaozhi AI Official.html)

### Mi inferencia sobre online y offline

La consola probablemente arma el estado del dispositivo combinando cosas como:

- sesion WebSocket activa
- conexion MQTT activa si usa ese transporte
- `last_connected_at`
- ultima actividad o ultima conversacion

O sea:

- `online` no significa necesariamente "el navegador ve al dispositivo"
- significa "el backend lo considera actualmente conectado o recientemente activo"

## Lo mas util para nuestro software

### 1. Separar vinculacion de sesion conversacional

Esto me parece una gran idea para copiar.

No conviene que "hablar" sea lo mismo que "vincular".

Lo correcto seria tener dos flujos:

- flujo de activacion o pairing
- flujo de conversacion

### 2. Usar un codigo temporal de vinculacion

Esto tambien vale mucho la pena.

Ventajas:

- el usuario entiende bien el proceso
- no necesita editar IDs manualmente
- no hace falta exponer directamente `MAC`, `Device-Id` o tokens
- es un patron de producto claro y reutilizable

### 3. Que la UI consulte al backend, no al dispositivo

Esto es importante para tu arquitectura.

La pantalla `Dispositivos` deberia obtener el estado desde tu backend local, no hablar directamente con XiaoZhi.

Eso te da:

- una sola fuente de verdad
- mejor seguridad
- menos acoplamiento
- mas facilidad para soportar varios transportes despues

Esto ya no es solo una recomendacion teorica: los HTML del panel oficial apuntan justamente a ese mismo patron.

### 4. Guardar un registry de dispositivos con estados claros

Yo te recomendaria manejar estados como:

- `unpaired`
- `pairing_pending`
- `paired_offline`
- `connecting`
- `online`
- `error`

Eso te va a ayudar mucho mas que solo `connected/disconnected`.

### 5. Mantener un estado online efimero y uno persistente

Conviene separar:

- estado persistente:
  - nombre
  - device id
  - client id
  - fecha de alta
  - ultima conexion
  - ultima desconexion
- estado efimero:
- sesion activa
- socket activo
- ultimo heartbeat
- ultimo error

### 6. Existe una ruta de `v1` mucho mas rapida: usar la consola oficial como runtime conversacional

Con lo que ahora se ve en `agent-config`, hay una estrategia muy valida para empezar:

- dejar que `xiaozhi.me` siga resolviendo voz, modelo, memoria y speaker recognition
- dejar que el dispositivo siga operando contra el ecosistema oficial
- usar tu app solo para preparar, arrancar y administrar el servidor MCP local que se conectara al `Custom Services` endpoint

Eso cambiaria bastante el alcance inicial.

En vez de construir primero:

- backend conversacional completo
- compatibilidad total `WebSocket` con XiaoZhi
- gestion propia del estado del dispositivo

la `v1` podria enfocarse en:

- crear o registrar un servidor MCP local
- conectarlo al endpoint `wss://api.xiaozhi.me/mcp/?token=...`
- administrar tools y modulos locales
- facilitar despliegue, logs, salud y debugging

Esta opcion no reemplaza tu vision de largo plazo, pero si puede ser una muy buena `v1 acelerada`.

## Lo que no conviene copiar tal cual

### 1. No acoplar toda la activacion a un servicio OTA remoto

En la consola oficial y en servidores compatibles, `OTA` parece ser un punto muy importante del flujo.

Para tu software local, puedes inspirarte en eso, pero no necesariamente copiarlo igual.

Podrias simplificarlo asi:

- un endpoint de registro del dispositivo
- un endpoint de activacion
- un endpoint de estado

Y dejar `OTA` solo para firmware y configuracion de servidor si luego lo necesitas.

### 2. No depender de un backend web publico para reconocer online u offline

Tu caso es distinto.

Tu software puede vivir en la misma PC donde corre el agente.

Eso te permite una version mucho mas simple:

- backend local mantiene las sesiones
- UI local consulta ese backend
- no necesitas una consola cloud para saber si el dispositivo esta vivo

### 2.1 No tomar la API privada de xiaozhi.me como base de producto

Aunque tecnicamente la consola usa endpoints internos del panel, no conviene basar tu app en eso porque:

- es una API privada y no documentada para terceros
- depende de sesion web y `Bearer token` del panel oficial
- puede cambiar sin aviso
- solo vera correctamente dispositivos que siguen ligados al backend oficial
- si tu dispositivo apunta a tu backend propio, `xiaozhi.me` dejara de ser la fuente de verdad

### 2.2 Si usas `Custom Services`, el valor de tu app no deberia ser "clonar la consola"

Si eliges la estrategia acelerada, tu app no deberia competir con:

- la pantalla de agentes
- la voz
- el historial conversacional
- `Speaker Recognition`

La oportunidad real estaria en otra parte:

- despliegue local sencillo del MCP server
- administracion de modulos y tools
- presets de integracion
- logs utiles
- permisos
- health checks
- facilidad para conectar `OpenCode` y otros MCPs

O sea: tu app no seria "otra consola XiaoZhi", sino "la consola local para hacer potente el endpoint MCP oficial".

### 3. No hacer que el navegador "detecte hardware"

Lo correcto sigue siendo:

- dispositivo -> backend
- backend -> UI

No:

- dispositivo -> navegador

## Propuesta adaptada a tu proyecto

### Flujo recomendado para tu software

1. XiaoZhi se conecta a tu backend local o a un gateway controlado por ti.
2. Si el dispositivo no esta vinculado, el backend genera un codigo temporal.
3. La UI local muestra un formulario `Agregar dispositivo`.
4. El usuario introduce ese codigo.
5. El backend resuelve el codigo y vincula el dispositivo con el perfil o agente local.
6. A partir de ahi, la pantalla `Dispositivos` muestra:
   - nombre
   - device id
   - ultima conexion
   - estado actual
   - version
   - si tiene OTA habilitado o no
7. Cuando el dispositivo abre el `WebSocket`, el backend actualiza su estado a `online`.
8. Si el socket se cierra o expira un heartbeat, pasa a `offline` o `error`.

### Arquitectura recomendada para esto

- `device-pairing-service`
- `device-registry-service`
- `session-registry`
- `transport/websocket`
- `control-plane api`
- `ui/devices`

## Respuesta corta a tu duda principal

Si, creo que si puedes reutilizar esa idea de la consola oficial.

Lo mas reutilizable no es la UI exacta, sino el patron:

- el dispositivo se presenta primero al backend
- el backend genera una prueba temporal de identidad
- la consola confirma esa vinculacion
- el estado online se consulta desde backend

Eso si encaja muy bien con tu software.

## Respuesta a la nueva duda: si conviene no rehacer todo y usar `Custom Services`

### Respuesta corta

Si, me parece una decision muy razonable para `v1`.

### Por que si tiene sentido

Porque la consola oficial ya te resuelve varias piezas caras:

- conversacion dispositivo-agente
- voz
- configuracion del personaje
- modelos
- memoria
- speaker recognition

Entonces tu app puede concentrarse en lo que hoy te da mas valor diferencial:

- levantar el servidor MCP local
- conectarlo al endpoint oficial del agente
- administrar herramientas y modulos locales
- facilitar que XiaoZhi pueda usar capacidades nuevas en tu PC

### Como cambia la arquitectura de `v1`

En esta estrategia, tu app deja de ser primero un `backend completo compatible con XiaoZhi` y pasa a ser primero un:

- `MCP local manager`
- `tool host`
- `module manager`
- `deployment helper`

### Riesgos de esta estrategia

- dependes del ecosistema oficial para la conversacion
- dependes de un endpoint y flujo que no controlas por completo
- si mañana cambian la UI o el flujo interno, tu integracion puede requerir ajustes

Aun asi, para ahorrar tiempo y validar valor real, la decision me parece buena.

## Respuesta a la duda: usar el backend oficial para saber si el device esta activo

### Respuesta corta

No te recomendaria usar el backend oficial de `xiaozhi.me` como mecanismo principal para saber si el dispositivo esta activo en tu app.

### Cuando si podria servirte

Solo como referencia secundaria o temporal si:

- el dispositivo sigue apuntando al backend oficial
- el usuario esta autenticado en `xiaozhi.me`
- aceptas depender de una API privada no estable

En ese escenario, tu app podria intentar consultar el panel oficial de forma parecida a su frontend.

### Por que no conviene como base

Para tu producto, eso tiene varios problemas:

- el estado real del dispositivo deberia vivir en tu backend
- si despues cambias a tu propio `WebSocket`, el backend oficial ya no sabra nada fiable
- quedarias atado a cambios de la web oficial
- tu app dependeria de credenciales y sesion externas

### Recomendacion correcta

Tu app deberia tener su propio estado de verdad para:

- `online`
- `offline`
- `pairing_pending`
- `error`
- `last_connected_at`
- `last_disconnected_at`
- `last_session_id`

Y, si algun dia quieres, podrias agregar una integracion opcional de solo lectura con `xiaozhi.me`, pero nunca como base principal de arquitectura.

## Conclusiones

- la consola oficial parece basarse en un flujo de activacion por codigo
- ese codigo parece nacer del backend cuando el dispositivo hace el chequeo inicial
- el navegador probablemente no detecta el dispositivo directamente
- la deteccion de online u offline seguramente viene del backend y sus sesiones activas
- esta idea si vale mucho la pena adaptarla a tu producto
- la consola oficial ya integra `MCP Settings` y `Speaker Recognition`
- el boton `Get MCP Endpoint` genera una URL `wss://api.xiaozhi.me/mcp/?token=...`
- la `v1` puede ser mucho mas rapida si tu app se enfoca primero en administrar un MCP server local para `Custom Services`

## Fuentes revisadas

- [Configuracion Wi-Fi y agregar dispositivos](https://home.xiaozhi.me/xz-docs/docs/tutorial-basics/configure-wifi-and-add-devices/)
- [WebSocket protocol del repo oficial](https://github.com/78/xiaozhi-esp32/blob/main/docs/websocket.md)
- [MQTT + UDP protocol del repo oficial](https://github.com/78/xiaozhi-esp32/blob/main/docs/mqtt-udp.md)
- [Issue con ejemplo de respuesta OTA](https://github.com/xinnan-tech/xiaozhi-esp32-server/issues/2324)
- [DeviceAppService.java](https://github.com/joey-zhou/xiaozhi-esp32-server-java/blob/main/xiaozhi-server/src/main/java/com/xiaozhi/device/DeviceAppService.java)
- [DeviceServiceImpl.java](https://github.com/xinnan-tech/xiaozhi-esp32-server/blob/main/main/manager-api/src/main/java/xiaozhi/modules/device/service/impl/DeviceServiceImpl.java)
- [README del server Java](https://github.com/joey-zhou/xiaozhi-esp32-server-java/blob/main/README.md)
- [Console.html](C:\Users\JimXL\Downloads\Console.html)
- [Agent Management - Xiaozhi AI Official.html](C:\Users\JimXL\Downloads\Agent Management - Xiaozhi AI Official.html)
- [Console (1).html](C:\Users\JimXL\Downloads\Console (1).html)
- [Console (2).html](C:\Users\JimXL\Downloads\Console (2).html)
- [agent-config.pBndCyaq.js](https://xiaozhi.me/assets/agent-config.pBndCyaq.js)
