# Investigacion de MCP Router para este proyecto

## Repositorio revisado

- `https://github.com/mcp-router/mcp-router`

## Que es

`MCP Router` es una aplicacion enfocada en administrar servidores `MCP` desde una sola interfaz local.

Su propuesta principal gira alrededor de:

- registrar servidores MCP locales y remotos
- agruparlos en proyectos
- separar contextos por workspaces
- activar o desactivar servidores y herramientas
- exponer una capa de integracion para clientes externos
- guardar logs y configuracion localmente

## Por que si es relevante para nuestro proyecto

Aunque no resuelve el problema de `XiaoZhi` ni el flujo completo de voz, si toca una parte central de nuestro producto:

- el panel de administracion de `MCPs`
- la organizacion modular del agente
- la observabilidad de herramientas y ejecuciones
- la experiencia de conectar capacidades enchufables

En otras palabras:

- no es la referencia principal del runtime de voz
- si es una referencia fuerte para el `Module and MCP Manager`

## Ideas concretas que vale la pena tomar

### 1. Un panel unificado para administrar MCPs

La idea mas valiosa es tener una sola pantalla donde el usuario pueda:

- ver que servidores MCP existen
- saber si estan activos o inactivos
- revisar sus tools disponibles
- activarlos o desactivarlos sin tocar archivos manuales

Esto encaja casi perfecto con la vision de nuestro producto.

### 2. Separar contexto por proyectos o espacios de trabajo

La idea de `Projects` y `Workspaces` es buena porque evita que el usuario tenga un solo entorno gigante y confuso.

Para nuestro caso, eso podria traducirse despues en:

- perfiles de uso
- modos del agente
- conjuntos distintos de modulos segun el contexto

Ejemplos:

- modo trabajo
- modo desarrollo
- modo personal

### 3. Activacion y desactivacion por herramienta

No solo importa si un servidor MCP esta conectado.

Tambien importa poder decidir que herramientas concretas estan habilitadas.

Esto es muy valioso para seguridad y experiencia de usuario:

- reducir superficie de riesgo
- evitar tools innecesarias
- dar control fino sin borrar integraciones completas

### 4. Logs y observabilidad desde el panel

Otra idea fuerte es que el usuario pueda ver:

- que se ejecuto
- cuando se ejecuto
- con que resultado
- que fallo

Para nuestro producto esto es todavia mas importante, porque estamos hablando de voz y control de PC.

### 5. Modelo mental de "manager" y no de "hack suelto"

Lo mas rescatable del repo no es solo una feature puntual.

Es el enfoque de producto:

- tratar los MCPs como un sistema administrable
- no como configuraciones escondidas y dispersas

Eso va muy alineado con lo que queremos construir.

## Lo que no conviene copiar tal cual

### 1. No usarlo como arquitectura base completa

`MCP Router` no esta disenado alrededor de:

- un dispositivo `XiaoZhi`
- un runtime conversacional de voz
- un transporte `WebSocket` de dispositivo IoT
- una orquestacion centrada en `OpenCode`

Por eso no deberia convertirse en la base total del proyecto.

### 2. No adoptar de entrada un enfoque desktop-first con Electron

El repo esta orientado a `Electron`.

Para nosotros sigue teniendo mas sentido:

- `web-first`
- backend separado
- empaquetado de escritorio opcional despues

### 3. No mezclar demasiado pronto cliente, runtime, configuracion y exposicion externa

Nuestro sistema necesita una separacion mas clara entre:

- vinculacion de dispositivo
- runtime de conversacion
- orquestacion
- ejecucion
- administracion de modulos y MCPs

`MCP Router` sirve muy bien en una de esas capas, no en todas.

## Advertencias utiles que nos deja

Su documentacion de seguridad es especialmente valiosa como advertencia de diseno.

Las areas que debemos cuidar desde el inicio son:

- almacenamiento seguro de tokens y secretos
- validacion estricta de URLs remotas
- proteccion frente a `SSRF`
- permisos finos para tools y modulos
- logs sin exponer secretos
- cuidado con rutas locales y archivos
- validacion de inputs antes de ejecutar acciones
- evitar que un modulo tenga mas privilegios de los necesarios

No hace falta copiar su implementacion, pero si aprender de esos riesgos.

## Como integrarlo en nuestra arquitectura

La mejor forma de tomar inspiracion de `MCP Router` es usarlo como referencia para una sola parte del sistema:

### Capa mas influenciada por MCP Router

- `Module and MCP Manager`

### Responsabilidades que si deberia heredar esta capa

- registrar MCPs locales y remotos
- listar tools por servidor
- activar o desactivar servidores
- activar o desactivar tools concretas
- mostrar estado de salud y ultima conexion
- guardar configuracion local
- mostrar logs y ejecuciones relacionadas

### Responsabilidades que no deberia absorber esta capa

- handshake del dispositivo `XiaoZhi`
- transporte `WebSocket` del dispositivo
- transcripcion o runtime de voz
- decision final de orquestacion
- ejecucion total del sistema

## Decision recomendada

`MCP Router` debe ser tratado como:

- una referencia de producto fuerte para el panel de administracion de `MCPs`
- una referencia secundaria para observabilidad y permisos
- una referencia debil para la arquitectura total del sistema

## Conclusiones para nuestro proyecto

- si vale la pena estudiarlo
- no conviene usarlo como base completa
- si conviene tomar su modelo de administracion de MCPs
- si conviene aprender de sus advertencias de seguridad
- la inspiracion correcta es su capa de control, no su producto entero
