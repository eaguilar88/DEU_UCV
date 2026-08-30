# Base de datos PostgreSQL

Este directorio contiene el esquema de la base de datos `deu`, gestionado con [golang-migrate](https://github.com/golang-migrate/migrate) (ejecutado como servicio Docker, no como dependencia de Go). Para más información sobre la herramienta en sí, ver su repositorio: https://github.com/golang-migrate/migrate.

## Estructura del directorio

```
postgresql/
├── README.md            este archivo
├── migrations/           migraciones versionadas — se aplican con `make migrate-up`
└── seeds/                datos de prueba opcionales — NUNCA se aplican automáticamente
```

- **`migrations/`** contiene todo el DDL (esquema, tipos enum, tablas, índices) y los datos semilla esenciales (roles, usuarios administradores/root, sus asignaciones de rol). Estas migraciones se aplican en todos los entornos, incluida producción, y son necesarias para poder iniciar sesión en el sistema.
- **`seeds/`** contiene datos de prueba/demo (`mock_grupos_completo.sql`: 27 grupos de extensión ficticios). Este archivo **no** es parte de la cadena de migraciones y nunca se ejecuta automáticamente — se carga manualmente solo en desarrollo.

## Requisitos previos

- Archivo `.env` configurado en la raíz del proyecto (`cp .env.example .env`).
- El contenedor `db` corriendo (`make start-backend` en desarrollo, `make start-prod` en producción).

## Ejecutar migraciones

Todos los comandos se ejecutan desde la raíz del repositorio.

| Comando | Qué hace |
|---|---|
| `make migrate-up` | Aplica todas las migraciones pendientes contra la base de datos de `docker-compose.yml` |
| `make migrate-down` | Revierte la última migración aplicada |
| `make migrate-down-all` | Revierte todas las migraciones (elimina el esquema `deu` por completo) |
| `make migrate-version` | Muestra la versión de migración actualmente aplicada |
| `make migrate-force version=N` | Fuerza la versión de migración registrada a `N` sin ejecutar SQL (ver [Solución de problemas](#solución-de-problemas)) |
| `make migrate-create name=nombre_descriptivo` | Crea un nuevo par de archivos `NNNNNN_nombre_descriptivo.up.sql` / `.down.sql` en `migrations/` |
| `make migrate-up-prod` | Igual que `migrate-up` pero contra `docker-compose.prod.yml` |

Estos targets son un envoltorio sobre el servicio `migrate` definido en los `docker-compose*.yml`. El equivalente directo (útil para depurar) es:

```
docker compose -f docker-compose.yml run --rm migrate \
  -path=/migrations -database="postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@db:$POSTGRES_PORT/$POSTGRES_DB?sslmode=disable" <comando>
```

## Crear una nueva migración

1. Ejecute `make migrate-create name=add_algo_table` — esto crea `NNNNNN_add_algo_table.up.sql` y `.down.sql` vacíos en `migrations/`, con el siguiente número secuencial.
2. Escriba el DDL/DML en el archivo `.up.sql` y su reversión exacta en el `.down.sql`.
3. **Una migración por unidad lógica**: una tabla (junto con los tipos `enum` que introduce) o un conjunto de datos semilla por migración — no mezcle varias tablas no relacionadas en un mismo archivo, y no separe una tabla de los tipos que solo ella usa.
4. **Regla de orden para tipos `enum` compartidos**: golang-migrate siempre revierte las migraciones en orden numérico descendente estricto. Si su migración reutiliza un tipo `enum` creado por una migración anterior (por ejemplo `deu.faculty_enum`), NO lo vuelva a crear ni lo elimine en su propio `.down.sql` — solo la migración que originalmente creó el tipo debe hacer `DROP TYPE` en su down, y solo es seguro porque para cuando esa migración se revierte, todas las migraciones posteriores que usan el tipo ya revirtieron sus tablas.
5. Pruebe el ciclo completo antes de hacer commit: `make migrate-up && make migrate-down && make migrate-up`.

## Reiniciar la base de datos de desarrollo

Dos opciones (reemplazan al antiguo script `limpiar_bd.sql`, que quedó obsoleto):

- **Reaplicar solo el esquema** (mantiene el contenedor/volumen):
  ```
  make migrate-down-all
  make migrate-up
  ```
- **Reinicio completo** (elimina también el volumen de datos):
  ```
  make stop-backend
  make start-backend
  make migrate-up
  ```

## Datos semilla

- **Datos esenciales** (roles, usuarios root/admin, asignaciones de rol — migraciones `000020`-`000022`): se aplican automáticamente con `make migrate-up` en todo entorno. Contraseña compartida de los usuarios semilla: `nolodire`.
- **Datos de prueba opcionales** (`seeds/mock_grupos_completo.sql`): deben cargarse manualmente, y solo tienen sentido en desarrollo. Ejemplo:
  ```
  docker compose -f docker-compose.yml exec -T db \
    psql -U $POSTGRES_USER -d $POSTGRES_DB -f - < postgresql/seeds/mock_grupos_completo.sql
  ```
  **Importante**: este script asume que los 14 usuarios semilla ya ocupan los ids 1-14 (crea sus propios usuarios de prueba a partir del id 15). Esto solo es válido justo después de un `make start-backend` + `make migrate-up` sobre un volumen nuevo — si se revierten y reaplican las migraciones de semilla sin recrear el volumen, las secuencias de identidad no se reinician y el script fallará o generará ids inesperados.

## Solución de problemas

- **"Dirty database version"**: una migración falló a mitad de camino y golang-migrate marcó la base como "dirty" para evitar aplicar más cambios sobre un estado inconsistente. Revise manualmente qué quedó aplicado, corrija la base si es necesario, y luego ejecute `make migrate-force version=N` (con `N` la versión correcta) para limpiar la marca antes de continuar. `force` no ejecuta SQL, solo actualiza el registro de versión — úselo con cuidado, ya que evita la protección de orden descrita arriba para tipos `enum` compartidos.
- **Ver el estado actual**: `make migrate-version`, o consulte directamente la tabla `public.schema_migrations` con `psql` (vive en el schema `public`, separada del schema `deu`, para que un `DROP SCHEMA deu CASCADE` nunca borre el historial de migraciones).
- Más información: https://github.com/golang-migrate/migrate
