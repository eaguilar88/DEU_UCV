# Recordatorios por correo (servicio `notifier`)

El servicio `notifier` envía correos de recordatorio para acciones pendientes y desactiva los grupos de extensión cuya renovación anual venció. Es un binario aparte del backend (`backend/cmd/notifier`), construido en la misma imagen Docker y ejecutado como su propio servicio en `docker-compose.yml` y `docker-compose.prod.yml`.

## Qué hace en cada ejecución

Cada ejecución (`notifications.Service.Run`, en `backend/internal/notifications/service.go`) realiza, en este orden:

1. **Desactiva los grupos vencidos.** Todo grupo activo cuyo `renewal_due_at` ya pasó queda con `is_active = false`. Aprobar su renovación lo reactiva (ver [Renovación anual de grupos](renovacion-grupos.md)).
2. **Recuerda a los coordinadores sus solicitudes pendientes.** Cuenta, por facultad, las solicitudes en estado `under_review` creadas hace `NOTIFIER_PENDING_REQUEST_AGE` o más:
   - solicitudes de cursos (`course_auth_requests`, facultad del curso);
   - solicitudes de grupos de extensión, tanto de registro como de renovación (`group_auth_requests`).

   Cada coordinador de la facultad (rol `faculty_admin`) recibe **un solo correo** con el resumen de ambas. Las solicitudes de la DEU también se envían a los usuarios `deu_admin`.
3. **Recuerda a los grupos su renovación.** Todo grupo aprobado cuyo `renewal_due_at` cae dentro de `NOTIFIER_RENEWAL_WINDOW` (o ya pasó) recibe un aviso en su correo de contacto (`deu.contacts`). Si ya venció, el correo lo indica. No se avisa a los grupos que ya tienen una renovación en revisión.

Si falla el envío a un destinatario, se registra en el log y se continúa con los demás. El error se reporta al final de la ejecución.

## Cuándo se ejecuta

- Una vez al arrancar el contenedor.
- Luego, todos los días a la hora `NOTIFIER_RUN_AT`, en la zona horaria `NOTIFIER_TIMEZONE`.

Ejecutarlo más de una vez es seguro: cada correo enviado queda registrado en `deu.notifications_log` (`kind`, `target`, `recipient`, `sent_at`), y el mismo recordatorio no se repite al mismo destinatario hasta que pase `NOTIFIER_RESEND_INTERVAL`. Por eso un reinicio del contenedor, o una ejecución manual, no duplica correos.

| `kind` | `target` | Destinatario |
|---|---|---|
| `coordinator_pending_requests` | Código de la facultad (p. ej. `Ciencias`) | Correo del coordinador o admin DEU |
| `group_renewal` | ID del grupo | Correo de contacto del grupo |

## Configuración (`.env`)

El `notifier` lee el mismo `.env` que el backend. Necesita la conexión a la base de datos (`POSTGRES_*`) y al servidor de correo (`EMAIL_*`), además de sus propias variables:

| Variable | Valor por defecto | Descripción |
|---|---|---|
| `NOTIFIER_RUN_AT` | `08:00` | Hora diaria de ejecución (`HH:MM`). |
| `NOTIFIER_TIMEZONE` | `America/Caracas` | Zona horaria de `NOTIFIER_RUN_AT`. La zona va embebida en el binario, así que funciona en la imagen alpine. |
| `NOTIFIER_PENDING_REQUEST_AGE` | `72h` | Antigüedad mínima de una solicitud en revisión para recordársela a los coordinadores. |
| `NOTIFIER_RENEWAL_WINDOW` | `720h` (30 días) | Desde cuánto antes de `renewal_due_at` se avisa a los grupos. **El backend también lee esta variable:** es la ventana en la que un grupo puede enviar su solicitud de renovación. |
| `NOTIFIER_RESEND_INTERVAL` | `168h` (7 días) | Tiempo mínimo antes de repetir el mismo recordatorio al mismo destinatario. |

Las duraciones usan el formato de Go (`time.ParseDuration`): `h`, `m`, `s`. No se admiten días (`d`): 30 días se escriben `720h`.

Ejemplo:

```
# -- Notifier (reminder emails) --
NOTIFIER_RUN_AT=08:00
NOTIFIER_TIMEZONE=America/Caracas
NOTIFIER_PENDING_REQUEST_AGE=72h
NOTIFIER_RENEWAL_WINDOW=720h
NOTIFIER_RESEND_INTERVAL=168h
```

## Ejecución

En Docker se levanta junto al resto de servicios (`make start-backend` o `make start-prod`). Para ver sus logs:

```
docker compose logs -f notifier
```

Para enviar los recordatorios pendientes una sola vez y terminar, use la opción `-once`. Es útil para probar:

```
docker compose run --rm notifier ./notifier -once
```

En desarrollo local, con solo la base de datos en Docker (`make start-db`), desde `backend/`:

```
go run ./cmd/notifier -once
```

## Plantillas de correo

Están en `backend/internal/email/templates/`:

| Plantilla | Uso |
|---|---|
| `coordinator_pending_requests_reminder.html` | Resumen de solicitudes pendientes para coordinadores y DEU. |
| `group_renewal_reminder.html` | Aviso de renovación (próxima o vencida) para el grupo. |

Para agregar un recordatorio nuevo: cree la plantilla, regístrela en `backend/internal/email/templates.go`, defina un nuevo `kind` en `backend/internal/entities/notification.go` y use `Service.remind`, que ya se encarga de no repetir el envío y de registrarlo.

## Datos relacionados

- `deu.extension_groups.renewal_due_at`: fecha en que vence el registro anual del grupo. Se fija al aprobar el grupo (un año después) y se extiende un año al aprobar cada renovación. Es `NULL` en los grupos que nunca fueron aprobados, y esos grupos no reciben recordatorios.
- `deu.notifications_log`: historial de recordatorios enviados. Para revisar los últimos:

  ```sql
  SELECT kind, target, recipient, sent_at
  FROM deu.notifications_log
  ORDER BY sent_at DESC
  LIMIT 20;
  ```
