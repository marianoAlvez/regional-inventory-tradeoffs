# Flujos de uso

| Caso | Recorrido y resultado |
| --- | --- |
| Comparación inicial | El visitante abre la demo, espera al motor y usa el ejemplo de demanda asimétrica. Lee confirmaciones, pendientes, rechazos y stock inaccesible. |
| Pérdida posterior | Selecciona pérdida de A. La comparación mantiene la demanda y la confirmación anterior a la pérdida. Distingue qué registros pueden recuperarse. |
| Sin acceso a mayoría | Selecciona un ejemplo sin mayoría. En el lado de quórum queda demanda pendiente. Los cupos confirman hasta el límite local. |
| Entrada inválida | Intenta asignar a A más unidades que el stock total. La interfaz o el motor informan el problema. El visitante corrige y obtiene un resultado para la nueva entrada. |
| Reproducción | Abre el repo, revisa el ADR y ejecuta el ejemplo JSON con Go. |

Las reglas de los estados están en el [contrato](05-esquema-backend.md).
