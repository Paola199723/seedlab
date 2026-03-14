# SeedLab

Aplicación en Go para generar SQL, Excel y diagramas UML a partir de una base de datos PostgreSQL.

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
6. Ejecutar `./seedlab`

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