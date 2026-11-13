# Inventario multirregión

**Arquitectura de software · Go, WebAssembly y TypeScript**

Una tienda vende el mismo producto desde dos regiones. Cuando se corta la comunicación, tiene que decidir qué compradores reciben una confirmación y qué pedidos quedan pendientes. La demo compara dos formas de tomar esa decisión con el mismo stock y los mismos pedidos.

**[Abrir demo pública](https://marianoalvez.github.io/regional-inventory-tradeoffs/)** · **[Leer la decisión de arquitectura](docs/adr/001-politica-de-confirmacion.md)** · **[Ver las ejecuciones de CI](https://github.com/marianoAlvez/regional-inventory-tradeoffs/actions/workflows/quality.yml)**

## Probalo en un minuto

1. Abrí la demo y elegí **Pérdida de B**. Hay 120 unidades y llegan 40 pedidos a A y 60 a B. Durante el corte, solo A puede guardar confirmaciones en dos de las tres réplicas.
2. **Quórum** confirma 40 pedidos y deja 60 pendientes. Cuando B se pierde, las 40 confirmaciones siguen recuperables.
3. **Cupos regionales** confirma los 100 pedidos. Las 60 confirmaciones guardadas solo en B quedan irrecuperables cuando esa región se pierde.

En ese escenario, los cupos confirman 60 pedidos más durante el corte. El quórum protege todas las confirmaciones que llega a dar. El [ADR](docs/adr/001-politica-de-confirmacion.md) explica qué elegiría el negocio ficticio y qué tendría que cambiar para reconsiderarlo.

## El caso

Ventas quiere confirmar en ambas regiones aun con la red dividida. Operaciones exige evitar la sobreventa. Finanzas exige recuperar las confirmaciones tras perder una región. El producto tiene que definir qué promete cuando esas condiciones compiten.

| Diseño | Qué protege | Qué cede |
| --- | --- | --- |
| Confirmación por quórum | Un inventario global y confirmaciones recuperables tras perder un dominio, cuando vuelve a operar una mayoría. | La región sin acceso a mayoría solo registra intenciones pendientes, sin asignar stock. |
| Cupos regionales | Confirma localmente dentro de una asignación fija y mantiene el límite global de stock. | Puede rechazar con stock libre en la otra región. Una confirmación que solo se guarda en la región perdida queda irrecuperable en el modelo. |

Con 120 unidades, 60 asignadas a cada región, demanda de 90 en A y 10 en B, mayoría en A y pérdida posterior de A:

| Resultado | Quórum | Cupos |
| --- | ---: | ---: |
| Confirmados antes de perder A | 90 | 70 |
| Pendientes | 10 | 0 |
| Rechazados | 0 | 30 |
| Confirmados recuperables después | 90 | 10 |
| Registros de confirmación irrecuperables | 0 | 60 |
| Unidades libres inaccesibles frente a la demanda rechazada | 0 | 30 |

Los pendientes nunca se prometieron. Los registros irrecuperables sí correspondían a pedidos confirmados. Esta cifra describe los registros que el modelo puede reconstruir. El laboratorio no reasigna unidades después de una pérdida.

## Cómo está hecho

```text
Navegador y TypeScript → adaptador WebAssembly ┐
                                              ├→ caso de uso Go → políticas de dominio
Entrada JSON en terminal → adaptador CLI      ┘
```

**Strategy.** Quórum y cupos son funciones Go del mismo tipo. Cada una decide qué pedidos confirma y dónde guarda los registros. Un evaluador aplica a ambas el mismo cálculo de stock, pedidos y recuperación. Así, una diferencia en la comparación viene de la política y sus copias durables.

**Puertos y adaptadores.** El caso de uso `application.Simulate` es el punto de entrada común. La terminal y WebAssembly adaptan su entrada al mismo contrato. El dominio desconoce JSON, la terminal y el navegador. TypeScript presenta el resultado y el hospedaje estático ejecuta el mismo Go que las pruebas. [Arquitectura hexagonal de Alistair Cockburn](https://alistair.cockburn.us/hexagonal-architecture/)

**Derechos de venta preasignados.** La política de cupos usa una variante fija de la idea de escrow. Divide el stock antes de la partición para que cada región pueda vender localmente hasta su límite. La confirmación queda en esa región durante la ventana simulada. [Trabajo original sobre escrow de Patrick O’Neil](https://ics.uci.edu/~cs223/papers/p405-o_neil.pdf)

| Para revisar | Archivo |
| --- | --- |
| Qué confirma cada política y dónde queda una copia durable | [Estrategias en Go](internal/domain/policies.go) |
| Cómo se contabilizan los pedidos y qué se recupera al perder una región | [Evaluador común](internal/domain/inventory.go) |
| Cómo se convierte el resultado en una comparación y una secuencia de eventos | [Caso de uso](internal/application/simulate.go) |
| Cómo terminal y navegador invocan el mismo caso de uso | [CLI](cmd/inventory/main.go) y [adaptador WebAssembly](cmd/wasm/main.go) |
| Cómo la demo presenta el resultado | [Interfaz TypeScript](web/src/main.ts) |

El simulador usa datos ficticios y tres fases: asignación inicial, pedidos durante el corte y pérdida posterior de una región de venta. Presupone un commit durable por consenso con dos de tres réplicas y modela sus consecuencias sobre los pedidos. La réplica C guarda una copia de las confirmaciones. Elección de líder, replicación real, latencia y recuperación de pedidos pendientes quedan fuera del modelo.

## Cómo se probó

| Prueba | Qué verifica |
| --- | --- |
| [Dominio](internal/domain/inventory_test.go) | Dos escenarios con resultados calculados de antemano, simetría al intercambiar regiones y 6.615 combinaciones de stock, cupos, demanda y fallas. Comprueba el límite del stock y la conservación de pedidos entre estados. |
| [Validación JSON](internal/adapter/jsoninput/simulate_test.go) | Rechaza campos obligatorios ausentes, cantidades inválidas, regiones desconocidas y entradas demasiado grandes. |
| [Paridad CLI y WebAssembly](scripts/check-wasm.mjs) | Compara la respuesta completa de los dos adaptadores en tres escenarios y comprueba cifras concretas para la pérdida de A. |
| [Navegador Chromium](web/test/demo.test.mjs) | Ejecuta la demo con el binario Go, comprueba las cifras de los tres escenarios, cambia la región perdida y prueba cupos inválidos y fallas del motor. También comprueba la vista móvil y el uso de un preset con teclado. |

GitHub Actions ejecuta estas pruebas y publica la demo solo después de que pasan. [Ver historial de CI](https://github.com/marianoAlvez/regional-inventory-tradeoffs/actions/workflows/quality.yml)

## Reproducir

Requisitos locales: Go 1.24 o posterior y Node.js 24 con npm.

```powershell
Get-Content -Raw examples/regional-loss.json | go run ./cmd/inventory
go vet ./...
go test ./...
npm ci --prefix web
node scripts/build.mjs
node scripts/check-wasm.mjs
```

Para ejecutar las pruebas de navegador desde el mismo directorio:

```powershell
npm exec --prefix web -- playwright install chromium
node --test web/test/demo.test.mjs
```

Para abrir una copia local:

```powershell
node scripts/serve.mjs
```

Abrí `http://127.0.0.1:4173/`. En Linux/macOS, el comando CLI equivalente es `go run ./cmd/inventory < examples/regional-loss.json`. CI usa `go test -race -cover ./...` en Linux.

## Documentación

- [Requisitos de producto](docs/producto/01-prd.md)
- [Requisitos técnicos](docs/producto/02-trd.md)
- [Diseño de interfaz](docs/producto/03-diseno-ui-ux.md)
- [Flujos](docs/producto/04-appflow.md)
- [Contrato y estados](docs/producto/05-esquema-backend.md)
- [Plan de implementación](docs/producto/06-plan-implementacion.md)

Código original del portafolio de Mariano Alvez, desarrollado con asistencia de IA para este laboratorio. Los datos y escenarios son ficticios. El repositorio no contiene código, datos ni artefactos de clientes.
