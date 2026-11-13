# Plan de implementación

| Incremento | Depende de | Comprobación |
| --- | --- | --- |
| Modelo de pedidos e inventario en Go | Definiciones del ADR y contrato | P1 y P2. Comparar los valores del ejemplo y probar invariantes con variaciones de stock y pérdida. |
| CLI y adaptador WebAssembly | Caso de uso común | P3 y P4. Verificar entradas inválidas y paridad de resultados de los dos adaptadores. |
| Demo interactiva | Adaptador de navegador y diseño | P1 a P3. Ejecutar ejemplos, editar demanda y comprobar un error recuperable. |
| Publicación reproducible | Pruebas y compilación aprobadas | P4. CI publica la demo y una sesión anónima abre la página y recalcula los resultados. |

Los requisitos y sus identificadores están en el [PRD](01-prd.md). El laboratorio termina con un repositorio público y una demo verificable. La implementación de consenso real es un proyecto de alcance distinto.
