# Inventario multirregión

**Arquitectura de software · Go, WebAssembly y TypeScript**

¿Qué pedidos conviene dejar pendientes durante una partición para proteger el inventario y las confirmaciones? Este laboratorio aplica dos políticas a los mismos pedidos y luego simula la pérdida de una región de venta.

**[Abrir demo pública](https://marianoalvez.github.io/regional-inventory-tradeoffs/)** · **[Leer la decisión de arquitectura](docs/adr/001-politica-de-confirmacion.md)** · **[Ver las ejecuciones de CI](https://github.com/marianoAlvez/regional-inventory-tradeoffs/actions/workflows/quality.yml)**

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

El dominio desconoce JSON, la terminal y el navegador. El caso de uso compara ambas políticas sobre la misma entrada. El binario WebAssembly permite que la demo pública ejecute el mismo Go que el CLI y las pruebas, con hospedaje estático y sin un proceso servidor. TypeScript mantiene formularios y presentación fuera del dominio. La separación hexagonal se aplica en las fronteras que efectivamente cambian.

**Alcance del modelo.** Hay dos regiones de venta, un tercer dominio de réplica para quórum y tres fases: asignación inicial, demanda durante la partición y pérdida posterior de hasta una región. La entrada indica qué región tiene acceso a mayoría. El modelo parte de la garantía de commit por consenso y comprueba sus consecuencias sobre pedidos. No mide latencia, no implementa Raft, elecciones ni replicación real. Las cifras son ficticias.

## Reproducir

Requisitos locales: Go 1.24 o posterior y Node.js 24 con npm.

```powershell
Get-Content -Raw examples/regional-loss.json | go run ./cmd/inventory
go vet ./...
go test -race ./...
npm ci --prefix web
node scripts/build.mjs
node scripts/check-wasm.mjs
node scripts/serve.mjs
```

Abrí `http://127.0.0.1:4173/`. En Linux/macOS, el comando CLI equivalente es `go run ./cmd/inventory < examples/regional-loss.json`. Si la instalación local de Go en Windows carece del compilador C requerido por el detector de carreras, ejecutá `go test ./...`. Las pruebas verifican invariantes de inventario y demanda en 6.615 combinaciones acotadas. CI ejecuta el detector en Linux y compara la respuesta del motor WebAssembly con la del CLI antes de publicar.

## Documentación

- [Requisitos de producto](docs/producto/01-prd.md)
- [Requisitos técnicos](docs/producto/02-trd.md)
- [Diseño de interfaz](docs/producto/03-diseno-ui-ux.md)
- [Flujos](docs/producto/04-appflow.md)
- [Contrato y estados](docs/producto/05-esquema-backend.md)
- [Plan de implementación](docs/producto/06-plan-implementacion.md)

Código original del portafolio de Mariano Alvez, desarrollado con asistencia de IA para este laboratorio. Los datos y escenarios son ficticios. El repositorio no contiene código, datos ni artefactos de clientes.
