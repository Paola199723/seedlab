# SeedLab

Aplicación en Go para generar SQL, Excel y diagramas UML a partir de una base de datos PostgreSQL.

## ✨ Características de v1.0

### 1️⃣ Generador de PNG de la base de datos
- Crea diagramas con:
  - Nombres de tablas  
  - Columnas de cada tabla  
  - Relaciones entre tablas  
- Actualmente, las relaciones se muestran **verticalmente**.  

### 2️⃣ Creador de diagramas editables en draw.io
- Genera un archivo que se puede abrir y **modificar directamente en draw.io**.  
- Representa la estructura completa de la base de datos.  

### 3️⃣ Generador de Excel editable
- Cada hoja representa una tabla de la base de datos.  
- Cada columna de la hoja corresponde a una columna de la tabla.  
- Permite **llenar datos manualmente** que luego se transforman en SQL.  

### 4️⃣ Generador de SQL versionado
- Genera **sentencias INSERT listas para ejecutar** en la base de datos.  
- Incluye scripts de **rollback** para revertir cambios si es necesario.  

### 5️⃣ Soporte de variables de entorno (`.env`)
- Permite configurar:
  - Conexión a la base de datos (`DATABASE_URL`)  
  - Nivel de logs (`LOG_LEVEL`)  
- El usuario puede usar su propia base de datos local **sin exponer credenciales privadas**.  

---
## Arquitectura

Clean Architecture con capas:
- Domain: Entidades y lógica de negocio
- UseCase: Casos de uso
- Repository: Acceso a datos
- Adapter: Interfaces (CLI)
- Frameworks: Librerías externas

## Configuración

1. Instalar Go 1.19+
2. Clonar el repositorio
3. Configurar .env en configs/.env con DATABASE_URL
4. Ejecutar `go mod tidy`
5. Construir `go build ./cmd/seedlab`
6. Ejecutar `./seedlab` o ejecutar proyecto completo `go run cmd/seedlab/main.go`

## Funcionalidades

- Leer tablas y relaciones de BD PostgreSQL
- Ordenar tablas por dependencias
- Seleccionar tablas relacionadas automáticamente
- Generar Excel con parámetros de tablas
- Generar .draw (Draw.io) y PNG de diagramas UML

## Dependencias

- github.com/jackc/pgx/v5: PostgreSQL driver
- github.com/rivo/tview: Terminal UI
- github.com/xuri/excelize/v2: Excel generation
- github.com/awalterschulze/gographviz: Graph generation
