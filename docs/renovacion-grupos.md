# Renovación anual de grupos de extensión

El registro de un grupo de extensión dura un año. Para renovarlo, el usuario del grupo envía una **solicitud de renovación** con todos los datos del grupo. Esos datos **no se aplican de inmediato**: quedan como una propuesta y solo reemplazan los datos del grupo cuando todas sus solicitudes de aval son aprobadas.

## Endpoint

`PUT /groups/requests/:id` (con autenticación JWT)

- **Payload:** el mismo formulario multipart que la creación (`POST /groups/requests`): `nombre`, `descripcion`, `tipo`, `fundacion`, `es_multidisciplinario`, `facultad`, `objetivo`, `ubicacion`, `correo`, `telefono`, `miembros` (JSON), `logo`, `proyecto_grupo` y `documento_miembro_N` por cada miembro.
- **Semántica PUT:** se envía el grupo completo y, una vez aprobado, reemplaza todo lo que hay en la base de datos para ese grupo: datos, contactos, miembros, logo, proyecto y documentos.
- **Respuesta:** `202 {"id": "<id de la renovación>"}`.

| Código | Motivo |
|---|---|
| `400` | Payload inválido o faltan archivos obligatorios. |
| `403` | Quien la envía no es el usuario del grupo (`extension_groups.user_id`, la cuenta creada al aprobar el grupo). |
| `404` | El grupo no existe. |
| `409` | La renovación aún no está abierta, o el grupo ya tiene una renovación en revisión. |

### Cuándo se puede enviar

- **El grupo debe haber sido aprobado.** Los grupos aprobados tienen `renewal_due_at`; si está en `NULL`, el grupo nunca fue aprobado y no puede renovar.
- **Debe estar dentro de la ventana.** Se acepta desde `renewal_due_at - NOTIFIER_RENEWAL_WINDOW` en adelante, también con el grupo ya vencido (ver [Recordatorios por correo](recordatorios.md#configuración-env)).
- **Solo puede haber una renovación en revisión por grupo.** Lo garantiza el índice único parcial `group_renewals_one_pending_idx`.

## Flujo

1. **Envío.** El usuario del grupo envía la renovación. Se crean:
   - una fila en `deu.group_renewals`, con la propuesta en `payload` (JSONB, sin archivos) y `status = 'under_review'`;
   - las solicitudes de aval en `deu.group_auth_requests`, con `renewal_id` apuntando a esa renovación. Los aprobadores son los mismos que en la creación: solo la DEU si el grupo es multidisciplinario; la facultad del grupo y la DEU si no;
   - los archivos propuestos, en B2 y en `deu.files` (ver la sección siguiente).

   Los aprobadores reciben el correo "Nueva solicitud de renovación de grupo de extensión".
2. **Revisión.** Los admins ven la renovación en `/admin/group-requests` igual que cualquier otra solicitud. La respuesta incluye `es_renovacion: true` y, en el detalle (`GET /admin/group-requests/:id`), `renovacion` con la propuesta completa y enlaces firmados a sus archivos.
3. **Aprobación.** Mientras queden solicitudes pendientes de esa renovación, aprobar solo marca la solicitud. Al aprobar la última, en una sola transacción:
   - se reemplazan los datos de `extension_groups`, sus contactos (`deu.contacts`) y sus miembros (los anteriores se borran de forma lógica con `deleted_at`);
   - los archivos propuestos pasan al grupo y a sus miembros nuevos, y los anteriores se borran de forma lógica;
   - el grupo se reactiva (`is_active = true`) y `renewal_due_at` se extiende un año desde la fecha vigente, o desde hoy si ya había vencido;
   - la renovación queda `approved`.

   El grupo recibe el correo "Renovación de grupo aprobada".
4. **Rechazo.** Rechazar cualquiera de sus solicitudes rechaza **toda** la renovación (las solicitudes hermanas pendientes y la fila de `group_renewals`). El grupo conserva sus datos actuales y puede enviar una renovación nueva.

Para decidir si "todas están aprobadas" solo se cuentan las solicitudes de la misma ronda: las de la creación (`renewal_id IS NULL`) o las de una misma renovación. Las solicitudes de rondas anteriores, aprobadas o rechazadas, no influyen.

## Archivos de una renovación: owner `group_renewal`

> **Importante:** mientras la renovación está en revisión, sus archivos **no** pertenecen al grupo. No los busque con `owner_type = 'group'` ni con `owner_type = 'group_member'`.

Los archivos propuestos se guardan en `deu.files` con:

| Campo | Valor |
|---|---|
| `owner_type` | `group_renewal` |
| `owner_id` | ID de la renovación (`deu.group_renewals.id`), **no** el ID del grupo |
| `purpose` | `logo`, `proyecto_grupo` o `documento_miembro` |
| `metadata.member_index` | Solo en `documento_miembro`: posición del miembro en la lista `miembros` de la propuesta. Los miembros aún no existen en `group_members`, así que no hay `member_id`. |
| `file_key` (B2) | `files/groups/{group_id}/renewals/{renewal_id}/...` |

Por eso, mientras la renovación esté en revisión:

- `GET /groups/:id` sigue mostrando el logo, el proyecto y los documentos **vigentes** del grupo.
- Los archivos propuestos solo se ven en el detalle admin de la solicitud de renovación.

Para consultarlos:

```sql
SELECT id, purpose, file_key, metadata
FROM deu.files
WHERE owner_type = 'group_renewal' AND owner_id = <renewal_id> AND deleted_at IS NULL;
```

### Qué pasa con ellos al resolverse la renovación

- **Aprobada:** las filas se **reasignan**, no se copian:
  - logo y proyecto pasan a `owner_type = 'group'` y `owner_id = <group_id>`;
  - cada documento pasa a `owner_type = 'group_member'` y `owner_id = <id del miembro nuevo>`, y se agrega `member_id` a su metadata.

  El `file_key` en B2 no cambia, así que los archivos de un grupo renovado tienen rutas `.../renewals/{renewal_id}/...`. Los archivos anteriores del grupo y de sus miembros quedan con `deleted_at`. Sus objetos en B2 no se eliminan.
- **Rechazada:** los archivos quedan como `group_renewal` del ID de la renovación rechazada, como historial. No se asignan al grupo.
- **Falla al subir los archivos durante el envío:** la renovación se elimina (con sus solicitudes, por cascada), y el usuario puede volver a enviarla.

## Vencimiento

El servicio `notifier` desactiva (`is_active = false`) los grupos cuyo `renewal_due_at` pasó. El usuario del grupo puede seguir iniciando sesión y enviar su renovación. Al aprobarse, el grupo se reactiva. Mientras tanto sigue recibiendo el recordatorio semanal de renovación. Ver [Recordatorios por correo](recordatorios.md).

## Código relevante

| Parte | Ubicación |
|---|---|
| Endpoint y validación del envío | `backend/internal/groups/endpoints.go` (`RenewGroup`), `backend/internal/groups/service.go` (`RenewGroup`, `storeRenewalFiles`) |
| Aprobación, rechazo y detalle | `backend/internal/group_requests/service.go` (`applyRenewal`, `sameRound`, `attachRenewalFiles`) |
| Transacciones y mapeo del payload | `backend/internal/postgres_repository/postgres_group_renewals_repository.go`, `queries/group_renewal_queries.go` |
| Esquema | `postgresql/migrations/000036_*` (`renewal_due_at`), `000038_create_group_renewals_table` |
