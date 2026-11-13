# ADR 001 · Qué cuenta como una venta confirmada

## Contexto

Dos regiones venden el mismo producto. Hay un stock global de 120 unidades y un tercer dominio disponible como réplica para la topología por quórum. Se debe decidir qué recibe el comprador durante la separación de la red y qué registro puede recuperar el negocio después de perder una región de venta. Los valores son supuestos del laboratorio.

Ventas prefiere confirmar inmediatamente. Operaciones exige que las ventas confirmadas nunca excedan el stock. Finanzas exige que una confirmación sobreviva a la pérdida de un dominio. Pendiente significa que la intención se recibió sin que se le asigne stock o se prometa una venta.

## Alternativas y decisión

1. **Confirmar por quórum.** El modelo presupone que un líder con mayoría confirma solo después de un commit durable en dos de tres dominios. Puede usar todo el inventario disponible. Donde no llega la mayoría, deja la intención pendiente. Si se pierde un dominio, las confirmaciones anteriores se recuperan cuando se establece la mayoría de supervivientes.
2. **Confirmar con cupo local.** Se asignan derechos de venta estáticos que suman el stock. Cada región confirma dentro de su límite y registra la confirmación únicamente allí. Una región puede agotar su cupo mientras sobra stock en la otra. La pérdida del único registro local hace que esa confirmación resulte irrecuperable.

La decisión inicial para este negocio ficticio es confirmar por quórum porque Finanzas considera obligatoria la recuperación de confirmaciones. Producto acepta una confirmación diferida en el lado aislado. Esta prioridad es un supuesto de negocio y puede cambiarse en la evaluación.

## Consecuencias observables

| Escenario | Quórum | Cupos |
| --- | --- | --- |
| Mayoría en A, demanda 90/10, cupos 60/60 | 90 confirmaciones y 10 pendientes | 70 confirmaciones y 30 rechazos, aunque hay stock disponible en B |
| Sin mayoría, demanda 50/40, cupos 60/60 | 0 confirmaciones y 90 pendientes | 90 confirmaciones locales |
| Tras el primer caso se pierde A | 90 confirmaciones recuperables después de restablecer mayoría | 10 recuperables y 60 registros irrecuperables |

El costo de disponibilidad se mide con los pedidos pendientes y rechazados. El costo de durabilidad se mide con los registros confirmados irrecuperables. La demo también muestra stock inaccesible por cupo: demanda rechazada en una región que podría satisfacerse con unidades libres de la otra.

## Condición de revisión

Ventas y Finanzas revisan la política si el costo de diferir pedidos durante las particiones supera el límite acordado. Pueden separar una intención provisional de la confirmación, permitir un riesgo explícito por una clase de ventas o invertir en replicación independiente para las confirmaciones locales. El laboratorio mantiene cupos fijos y no evalúa esas opciones adicionales.

El laboratorio comprueba invariantes de un modelo agregado y determinista. Antes de adoptar una topología real hay que validar reintentos, replicación, comportamiento del cliente ante resultados desconocidos, recuperación y tiempo de operación bajo fallas en un entorno representativo.

[Gilbert y Lynch sobre consistencia y disponibilidad bajo partición](https://groups.csail.mit.edu/tds/papers/Gilbert/Brewer2.pdf) · [Ongaro y Ousterhout sobre Raft](https://raft.github.io/raft.pdf)
