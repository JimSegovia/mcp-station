# Objetivo

El objetivo inicial es que mi ESP32 con XiaoZhi, al que le puedo hablar y que puede conectarse a internet, se pueda conectar a mi computadora y, a traves de OpenCode, pueda manipular mi computadora.

## Producto final deseado

Como producto final, quiero construir un software propio al que pueda vincular mi XiaoZhi y que funcione como la interfaz principal para configurar, administrar y escalar mi agente local en la PC.

Este software deberia permitirme:

- vincular uno o varios dispositivos XiaoZhi
- administrar por modulos las capacidades del sistema
- conectar, activar, desactivar y configurar servidores MCP
- gestionar herramientas locales y remotas
- definir como mi agente puede actuar sobre la computadora
- usar OpenCode como una de las piezas centrales de ejecucion
- crecer con una arquitectura ordenada, extensible y mantenible

## Vision del sistema

La idea ya no es solo conectar el ESP32 a la computadora, sino crear una plataforma local de agente personal, con una buena arquitectura desde el inicio, para que despues pueda evolucionar sin rehacer todo.

En otras palabras, quiero que XiaoZhi sea la puerta de entrada por voz, y que mi software sea el centro de control del agente, los modulos, los MCP y las acciones sobre mi PC.
