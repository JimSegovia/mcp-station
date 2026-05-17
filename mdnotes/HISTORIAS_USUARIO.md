# Historias de usuario

## Objetivo de este documento

Registrar las primeras epicas e historias de usuario para la primera etapa del proyecto, alineadas con:

- [OBJETIVO.md](D:\Proyects\Code\XiaonZhi\Server\OBJETIVO.md)
- [ARQUITECTURA.md](D:\Proyects\Code\XiaonZhi\Server\ARQUITECTURA.md)

La idea es que estas historias ayuden a construir una v1 real y util, no solo una demo tecnica.

## Alcance de esta primera parte

Esta primera parte se enfoca en validar la base del producto:

- aprovechar el agente oficial de XiaoZhi en vez de rehacer la capa de voz
- conectar `Custom Services` a un servidor MCP local
- ejecutar acciones reales sobre la computadora a traves de tools
- administrar modulos y MCPs desde una interfaz
- establecer una base segura y escalable

## Epica 1: Integracion con el agente oficial de XiaoZhi

### HU-001 Conectar XiaoZhi oficial a un MCP local

Como usuario que ya usa XiaoZhi en la consola oficial, quiero conectar `Custom Services` a un servidor MCP local, para extender mi agente sin rehacer la capa de voz.

Criterios de aceptacion:

- el sistema debe permitir registrar o pegar el `MCP endpoint` oficial
- el sistema debe levantar o preparar un servidor MCP local compatible
- el sistema debe mostrar si la conexion al endpoint esta activa
- el usuario debe poder saber si la integracion quedo funcionando o no

### HU-002 Ver el estado de la integracion MCP

Como usuario, quiero ver si mi integracion con `Custom Services` esta conectada, desconectada o con errores, para entender rapidamente el estado del sistema.

Criterios de aceptacion:

- debe existir una vista de estado de la integracion
- debe mostrarse si el endpoint esta activo
- deben mostrarse errores basicos de conexion

### HU-003 Registrar el MCP endpoint oficial del agente

Como usuario, quiero ver y registrar el `MCP endpoint` oficial de XiaoZhi, para poder conectarlo correctamente a mi software.

Criterios de aceptacion:

- el sistema debe mostrar donde se obtiene el endpoint
- el usuario debe poder registrar o pegar el endpoint
- el usuario debe poder copiar ese valor facilmente

## Epica 2: Servidor MCP local minimo

### HU-004 Levantar un servidor MCP local

Como usuario, quiero que mi app pueda levantar un servidor MCP local, para exponer tools a mi agente oficial de XiaoZhi.

Criterios de aceptacion:

- el sistema debe poder iniciar un runtime MCP local
- debe registrar si el runtime inicio correctamente
- el sistema debe mostrar el estado del runtime

### HU-005 Exponer al menos una tool util al endpoint oficial

Como usuario, quiero que mi servidor MCP local exponga al menos una tool real, para confirmar que el flujo minimo funciona.

Criterios de aceptacion:

- el sistema debe registrar al menos una tool disponible
- la tool debe ser visible en el contexto de la integracion
- la disponibilidad de la tool debe quedar registrada en logs

### HU-006 Ver el historial reciente de conexion MCP

Como usuario, quiero revisar las conexiones recientes del endpoint y del servidor MCP local, para entender que paso cuando una integracion falla o funciona.

Criterios de aceptacion:

- el sistema debe registrar conexiones recientes
- debe mostrarse hora de inicio y estado
- debe existir al menos una vista simple de consulta

## Epica 3: Ejecucion de acciones sobre la PC

### HU-007 Ejecutar una accion local simple

Como usuario, quiero que XiaoZhi pueda ejecutar una accion simple en mi PC, para validar que el sistema no solo conversa sino que actua.

Criterios de aceptacion:

- debe existir al menos una accion local real
- la accion debe poder activarse desde una orden del agente
- la accion debe registrarse en el historial

Ejemplos validos para v1:

- abrir una aplicacion
- crear un archivo
- leer una carpeta permitida

### HU-008 Ejecutar una tarea a traves de OpenCode

Como usuario, quiero que ciertas tareas sean delegadas a OpenCode, para usarlo como motor principal de ejecucion inteligente sobre mi PC.

Criterios de aceptacion:

- debe existir un modulo adaptador hacia OpenCode
- el orquestador debe poder decidir cuando invocarlo
- el resultado debe quedar registrado

### HU-009 Diferenciar acciones de lectura y escritura

Como usuario, quiero que el sistema distinga entre acciones seguras y acciones sensibles, para evitar ejecuciones peligrosas sin control.

Criterios de aceptacion:

- las acciones deben clasificarse al menos en lectura y escritura
- las acciones sensibles deben poder bloquearse o requerir confirmacion

## Epica 4: Administracion modular

### HU-010 Ver los modulos disponibles

Como usuario, quiero ver que modulos tiene disponibles mi sistema, para entender que capacidades tengo activas o instalables.

Criterios de aceptacion:

- debe existir una lista de modulos
- cada modulo debe mostrar nombre, estado y descripcion breve

### HU-011 Activar y desactivar modulos

Como usuario, quiero activar o desactivar modulos, para controlar las capacidades disponibles del agente sin tocar codigo.

Criterios de aceptacion:

- el usuario debe poder activar un modulo
- el usuario debe poder desactivar un modulo
- el sistema debe reflejar el cambio de estado

### HU-012 Ver si un modulo fallo

Como usuario, quiero saber cuando un modulo falla, para diagnosticar problemas sin adivinar que paso.

Criterios de aceptacion:

- los errores de modulo deben quedar registrados
- la interfaz debe mostrar al menos un estado de error o advertencia

## Epica 5: Gestion de MCPs

### HU-013 Registrar un servidor MCP

Como usuario, quiero registrar un servidor MCP en el sistema, para expandir capacidades del agente desde una interfaz administrable.

Criterios de aceptacion:

- debe poder darse de alta un MCP
- el MCP debe guardar nombre, tipo y configuracion minima
- el sistema debe registrar si esta activo o no

### HU-014 Ver los MCPs conectados

Como usuario, quiero ver que MCPs tengo conectados, para administrar las capacidades externas del agente.

Criterios de aceptacion:

- debe existir una lista de MCPs registrados
- debe mostrarse estado basico de conexion o disponibilidad

### HU-015 Activar o desactivar un MCP

Como usuario, quiero activar o desactivar un MCP, para controlar que capacidades externas puede usar el agente.

Criterios de aceptacion:

- debe poder cambiarse el estado de un MCP
- el sistema debe reflejar el cambio sin editar archivos manualmente

## Epica 6: Seguridad minima de la v1

### HU-016 Bloquear acciones peligrosas por defecto

Como usuario, quiero que las acciones potencialmente peligrosas no se ejecuten libremente, para proteger mi computadora.

Criterios de aceptacion:

- debe existir una politica por defecto restrictiva
- las acciones sensibles deben estar bloqueadas o requerir aprobacion

### HU-017 Ver el historial de ejecuciones

Como usuario, quiero revisar que acciones ejecuto el agente, para tener trazabilidad y control.

Criterios de aceptacion:

- el sistema debe registrar accion, hora y resultado
- el historial debe poder consultarse desde la interfaz

### HU-018 Ver errores y eventos importantes

Como usuario, quiero revisar errores relevantes del sistema, para poder diagnosticar problemas de conexion, modulos o ejecuciones.

Criterios de aceptacion:

- deben registrarse errores de conexion
- deben registrarse errores de ejecucion
- debe existir una vista minima de logs

## Epica 7: Interfaz de administracion v1

### HU-019 Tener una interfaz local de administracion

Como usuario, quiero una interfaz local para administrar mi agente, para no depender de editar archivos manualmente.

Criterios de aceptacion:

- debe existir una interfaz web local
- debe ser accesible desde la misma PC
- debe cubrir al menos configuracion basica, dispositivo, modulos y logs

### HU-020 Entender rapidamente el estado general del sistema

Como usuario, quiero ver un resumen del estado del sistema al entrar, para saber rapidamente si todo esta funcionando.

Criterios de aceptacion:

- la interfaz debe mostrar al menos:
  - estado del endpoint MCP
  - estado del servidor MCP local
  - modulos activos
  - errores recientes

## Backlog priorizado

### P0

Historias sin las cuales la v1 no cumple su promesa base.

- `HU-001` Conectar XiaoZhi oficial a un MCP local
- `HU-002` Ver el estado de la integracion MCP
- `HU-003` Registrar el MCP endpoint oficial del agente
- `HU-004` Levantar un servidor MCP local
- `HU-005` Exponer al menos una tool util al endpoint oficial
- `HU-007` Ejecutar una accion local simple
- `HU-008` Ejecutar una tarea a traves de OpenCode
- `HU-016` Bloquear acciones peligrosas por defecto
- `HU-017` Ver el historial de ejecuciones
- `HU-019` Tener una interfaz local de administracion
- `HU-020` Entender rapidamente el estado general del sistema

### P1

Historias muy importantes para que el producto ya se sienta administrable y extensible.

- `HU-010` Ver los modulos disponibles
- `HU-011` Activar y desactivar modulos
- `HU-013` Registrar un servidor MCP
- `HU-014` Ver los MCPs conectados
- `HU-015` Activar o desactivar un MCP
- `HU-018` Ver errores y eventos importantes

### P2

Historias valiosas, pero no criticas para validar el primer corte.

- `HU-006` Ver el historial reciente de conexion MCP
- `HU-009` Diferenciar acciones de lectura y escritura
- `HU-012` Ver si un modulo fallo

## Descomposiciones separadas

- `HU-001` Conectar XiaoZhi oficial a un MCP local:
  [HU_001_VINCULACION_XIAOZHI.md](D:\Proyects\Code\XiaonZhi\Server\HU_001_VINCULACION_XIAOZHI.md)

## Referencias utiles

### Referencia 1: Documentacion oficial sobre MCP endpoint de XiaoZhi

Referencia compartida:

- `https://my.feishu.cn/wiki/HiPEwZ37XiitnwktX13cEM5KnSb`

Resumen de utilidad:

- esta documentacion es util como referencia oficial sobre como XiaoZhi expone y usa `MCP endpoints`
- ayuda a entender como conectar herramientas remotas al ecosistema XiaoZhi
- aporta buenas practicas para el diseno de tools MCP
- deja ver restricciones operativas reales del sistema

### Utilidad para historias de usuario

#### Utilidad alta

- `HU-001` Conectar XiaoZhi oficial a un MCP local
- `HU-013` Registrar un servidor MCP
- `HU-014` Ver los MCPs conectados
- `HU-015` Activar o desactivar un MCP
- `HU-010` Ver los modulos disponibles
- `HU-011` Activar y desactivar modulos
- `HU-008` Ejecutar una tarea a traves de OpenCode

#### Utilidad media

- historias relacionadas con la arquitectura general del `Module and MCP Manager`
- historias relacionadas con el diseno de tools remotas y conectores administrables

#### Utilidad baja o parcial

- futuras historias de backend propio compatible con XiaoZhi

Motivo:

- con el enfoque nuevo, esta referencia ya es bastante util para `HU-001`
- lo que deja de cubrir bien es la implementacion de un backend propio completo para el dispositivo

### Aprendizajes importantes a conservar

- los nombres de tools MCP deben ser claros y no ambiguos
- los nombres de parametros deben ser semanticos y faciles de entender para el modelo
- cada tool debe incluir una descripcion o docstring que indique cuando debe usarse
- las respuestas de tools deben ser compactas y estructuradas
- los retornos tienen limites de longitud
- la lista de tools tambien tiene limites
- cada MCP endpoint tiene limite de conexiones

### Implicaciones para el producto

- el sistema deberia tener una interfaz clara para administrar `MCP endpoints`
- el panel deberia mostrar informacion de estado y restricciones de los MCPs
- el diseno del `MCP Registry` debe considerar limites de conexion y de carga de tools
- el sistema deberia favorecer tools bien descritas y parametros bien nombrados
- estas reglas tambien pueden aplicarse al futuro `OpenCode Adapter` y a otros modulos ejecutables

## Nota de enfoque

Estas historias deben servir para construir una v1 util y clara.

Todavia no estamos bajando esto a tareas tecnicas detalladas ni a tickets de implementacion.

Primero estamos definiendo que necesita el producto para que el primer corte sea realmente bueno.
