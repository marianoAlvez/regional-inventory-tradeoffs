# Requisitos técnicos

El dominio Go mantiene la política de inventario y los estados. La aplicación evalúa ambos diseños sobre una entrada y produce una comparación determinista. CLI y WebAssembly adaptan JSON al mismo caso de uso. La interfaz TypeScript no calcula los resultados de negocio.

Se eligió WebAssembly para ejecutar Go desde una demo estática sin servidor. El costo de descargar el runtime y el binario se acepta a cambio de compartir la implementación verificada. No hay persistencia, dependencias externas del dominio ni servicios de clientes.

Go valida el [contrato](05-esquema-backend.md). Las pruebas de dominio comprueban conservación de demanda e inventario, no sobreventa y recuperación según la pérdida. CI construye la demo y compara la salida del CLI con WebAssembly. La [decisión de negocio](../adr/001-politica-de-confirmacion.md) fija las garantías que se modelan.
