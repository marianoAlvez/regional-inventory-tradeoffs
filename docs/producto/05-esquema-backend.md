# Datos y contrato

| Entrada JSON | Significado |
| --- | --- |
| `stock`, `quotaA` | Unidades iniciales y asignación local a A. B recibe la diferencia. |
| `demandA`, `demandB` | Pedidos de una unidad en cada región de venta. |
| `majority` | `A`, `B` o `none`. Acceso a la mayoría durante la demanda. |
| `loss` | `A`, `B` o `none`. Pérdida de un dominio de venta después de decidir pedidos. |

Las cantidades son enteros entre 0 y 1.000.000 y el cupo de A no supera el stock. `results` contiene `quorum` y `quota` en ese orden. Cada resultado tiene `confirmed`, `pending`, `rejected`, `recoverable`, `lost`, `remaining`, `stranded`, `regions` y `timeline`. Las regiones exponen `region`, `confirmed`, `pending`, `rejected` y `remaining`. La línea de tiempo contiene `title` y `detail`.

La suma de confirmados, pendientes y rechazados coincide con la demanda. La suma de recuperables e irrecuperables coincide con los confirmados. `remaining` mide stock libre antes de la pérdida. `stranded` mide la parte de la demanda rechazada que podría cubrirse con cupo libre de la otra región. Con quórum, el stock regional reportado es el stock global visible para la región con mayoría.

La entrada y el resultado son transitorios. El adaptador de navegador devuelve JSON con `error` ante una entrada inválida. Ver el [ejemplo](../../examples/regional-loss.json).
