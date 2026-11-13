# Requisitos de producto

Para ver qué hace el proyecto y probarlo en un minuto, empezá por el [README](../../README.md).

Un visitante evalúa si un comercio ficticio debe confirmar pedidos por quórum o con cupos por región. El resultado esperado es que pueda identificar la garantía que se pierde al cambiar la política.

| ID | Requisito | Aceptación |
| --- | --- | --- |
| P1 | Comparar con la misma entrada | Stock 120, cupos 60/60, demanda 90/10 y mayoría A producen 90 confirmados y 10 pendientes por quórum, frente a 70 confirmados y 30 rechazados por cupos. |
| P2 | Examinar una pérdida posterior | Si se pierde A, la demo muestra 0 registros irrecuperables en quórum y 60 con cupos. Pendientes y registros irrecuperables tienen significado distinto. |
| P3 | Explorar alternativas | Hay ejemplos reproducibles y controles para cambiar stock, asignación, demanda, mayoría y pérdida. Las entradas inválidas reciben una indicación. |
| P4 | Verificar sin cuenta | Repo y demo públicos, ejecución CLI y prueba del contrato entre CLI y navegador. |

Alcance inicial de un producto, dos regiones de venta y un evento de pérdida después de la demanda. Los visitantes no ingresan datos personales ni realizan pedidos. No se simulan latencia, dinero, reintentos individuales o recuperación de stock.
