# Requisitos técnicos

El dominio Go aplica Strategy con dos funciones de asignación. Cada política decide qué pedidos confirma y dónde guarda los registros. Un evaluador calcula con las mismas reglas el stock, los pedidos y la recuperación. La aplicación convierte esos resultados en una comparación determinista. CLI y WebAssembly adaptan JSON al mismo caso de uso. La interfaz TypeScript presenta el resultado.

Se eligió WebAssembly para ejecutar Go desde una demo estática sin servidor. El costo de descargar el runtime y el binario se acepta a cambio de compartir la implementación verificada. No hay persistencia, dependencias externas del dominio ni servicios de clientes.

Go valida el [contrato](05-esquema-backend.md). Las pruebas de dominio comprueban conservación de demanda e inventario, no sobreventa y recuperación según la pérdida. CI construye la demo y compara la salida del CLI con WebAssembly. La [decisión de negocio](../adr/001-politica-de-confirmacion.md) fija las garantías que se modelan.
