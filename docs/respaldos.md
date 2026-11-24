# Respaldos de la base de datos

El servicio `backup` (`backend/cmd/backup`) respalda a diario la base de datos PostgreSQL del sistema y guarda los respaldos fuera del VPS, en Backblaze B2. Usa la misma imagen que el backend con otro punto de entrada, igual que el `notifier`.

## Qué se respalda

La base de datos completa `${POSTGRES_DB}` del servicio `db` (postgres:16):

- El esquema `deu`, con todas las tablas del backend Go: usuarios, cursos, grupos, certificados, cola de trabajos, bitácora de notificaciones, etc.
- El esquema `public`, con las tablas de la landing en Rails, que comparte la misma base de datos.

No incluye la base de datos de Espacios Universitarios (`espacios-db`). Ese subsistema es independiente y tiene su propio script de respaldo (`espacios-universitarios/scripts/backup/backup.sh`).

Tampoco incluye los archivos subidos (imágenes, PDF, certificados), porque ya viven en B2.

## Qué hace en cada ejecución

1. Ejecuta `pg_dump --format=custom --no-owner --no-privileges` contra `db`. La contraseña viaja en `PGPASSWORD`, nunca como argumento visible. El volcado se escribe en un archivo temporal. Así, si `pg_dump` falla, nunca queda un respaldo parcial en el bucket.
2. Si el volcado está vacío o `pg_dump` terminó con error, la ejecución falla y no sube nada.
3. Sube el volcado a B2 con la clave `${BACKUP_PREFIX}${POSTGRES_DB}_AAAAMMDD_HHMMSS.dump`, con la hora en UTC. Por ejemplo: `backups/postgres/deu_20261124_070000.dump`.
4. Borra de B2 los volcados `.dump` bajo `BACKUP_PREFIX` con más de `BACKUP_RETENTION_DAYS` días. Nunca borra el volcado recién subido. Si el borrado falla, el respaldo nuevo se conserva y la limpieza se reintenta en la siguiente ejecución.

El formato *custom* de `pg_dump` ya viene comprimido y permite restaurar la base completa o tablas sueltas con `pg_restore`.

## Cuándo se ejecuta

- Una vez al arrancar el contenedor.
- Luego todos los días a las `BACKUP_RUN_AT`, en la zona horaria `BACKUP_TIMEZONE` (por defecto 03:00, hora de Caracas). Es la hora de menor uso del sistema y no coincide con el `notifier`, que corre a las 08:00.
- Si una ejecución falla, se registra en el log y se reintenta en la próxima ejecución programada.

## Configuración (`.env`)

| Variable | Por defecto | Descripción |
|---|---|---|
| `BACKUP_RUN_AT` | `03:00` | Hora diaria (HH:MM) |
| `BACKUP_TIMEZONE` | `America/Caracas` | Zona horaria de `BACKUP_RUN_AT` |
| `BACKUP_RETENTION_DAYS` | `30` | Días que se conserva cada respaldo; `0` conserva todos |
| `BACKUP_PREFIX` | `backups/postgres/` | Carpeta (prefijo) en el bucket |
| `BACKUP_TIMEOUT` | `1h` | Duración máxima de una ejecución |

También usa las variables `POSTGRES_*` y `BLACKBLAZE_B2_*` del backend.

En producción (`FLAVOR=prod`), las llaves de B2 se leen de los secretos de Docker `b2_key_id` y `b2_application_key`, igual que en la API.

## Ejecución

```bash
# Producción: el servicio arranca con el resto del stack
docker compose -f docker-compose.prod.yml up -d backup
docker compose -f docker-compose.prod.yml logs -f backup

# Respaldo inmediato (por ejemplo, antes de una migración delicada)
make backup-now-prod

# Local: el servicio está en el perfil "tools" y solo corre a pedido
make backup-now
```

El despliegue (`.github/workflows/deploy.yml`) construye y levanta `backup` junto con `backend` y `notifier`, porque comparten la imagen.

## Restauración

1. Descargar el respaldo desde B2: desde la consola web, o con cualquier cliente S3 (`aws s3 cp`, `rclone`) apuntando al endpoint de B2.
2. Copiarlo al contenedor de la base de datos y restaurarlo:

```bash
docker compose -f docker-compose.prod.yml cp deu_20261124_070000.dump db:/tmp/respaldo.dump

# Restauración completa sobre la base existente: borra y recrea cada objeto del respaldo
docker compose -f docker-compose.prod.yml exec db \
  pg_restore --clean --if-exists --no-owner -U "$POSTGRES_USER" -d "$POSTGRES_DB" /tmp/respaldo.dump

# Restaurar una sola tabla (por ejemplo, deu.groups)
docker compose -f docker-compose.prod.yml exec db \
  pg_restore --data-only --table=groups --schema=deu -U "$POSTGRES_USER" -d "$POSTGRES_DB" /tmp/respaldo.dump
```

3. Detener el `backend`, el `notifier` y la `landing` mientras se restaura la base completa, y levantarlos de nuevo al terminar.

Para probar un respaldo sin tocar producción, restaurarlo en una base nueva:

```bash
docker compose -f docker-compose.prod.yml exec db createdb -U "$POSTGRES_USER" deu_prueba
docker compose -f docker-compose.prod.yml exec db pg_restore --no-owner -U "$POSTGRES_USER" -d deu_prueba /tmp/respaldo.dump
```

## Código

- `backend/cmd/backup/main.go`: programación diaria, señales y configuración.
- `backend/internal/backups/service.go`: volcado a archivo temporal, subida a B2 y limpieza por retención.
- `backend/internal/backups/pgdump.go`: invocación de `pg_dump`.
- `backend/internal/storage/client.go`: `ListObjects`, usado para la limpieza.
- `backend/internal/schedule`: cálculo de la próxima ejecución, compartido con el `notifier`.
