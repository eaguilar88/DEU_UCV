# Análisis de brechas: Backend Go vs. mock Node de `diplomados/`

## 1. Introducción

`diplomados/` (`ecp_ucv`) es la app Next.js del portal de cursos/diplomados. Durante su desarrollo temprano incluyó un backend simulado propio: rutas de API de Next.js (`src/app/api/**`) que actúan como una capa fina de lógica de negocio delante de un almacén REST crudo servido con `json-server --watch db.json`. Ese mock se desarrolló en paralelo (y en algunos puntos, por delante) del backend real en Go, y por eso codifica un flujo de aprobación más completo del que existe hoy en `backend/internal/`.

Este documento compara ambos backends y documenta **en una sola dirección**: qué existe en el mock de Node que falta (total o parcialmente) en el backend Go. No se documentan capacidades que el backend Go tiene y el mock no — eso queda fuera del alcance.

**Alcance:** cursos, diplomados, usuarios, y los endpoints de administración relacionados con solicitudes de curso ("course requests"). Los endpoints exclusivos de grupos se excluyen; ver §6 para la sección de solapamiento curso+grupo.

**Método:** lectura directa del código de rutas/handlers en `diplomados/src/app/api/**` + esquemas de `db.json`, comparado contra `backend/internal/**` + el enrutamiento de `backend/cmd/deu/main.go`.

**Leyenda de prioridad:**
- 🔴 **Alta** — bloquea o distorsiona el flujo de aprobación/negocio central.
- 🟡 **Media** — funcionalidad real que falta, pero con workaround o impacto acotado.
- ⚪ **Baja** — diferencia de forma/nomenclatura o mejora menor, sin bloquear el flujo.

---

## 2. Cursos

### 2.1 Apertura de cohorte/período con reglas de negocio (🔴 Alta) — ✅ Resuelto

- **Node:** `POST /api/courses/[id]/open` (`diplomados/src/app/api/courses/[id]/open/route.ts`). Body: `{cohortName, startDate, endDate, capacity}`. Antes de crear la cohorte valida `canOpen`: el curso debe estar en `estado='cerrado'`, **o** estar `aprobado`/`aprobada` **y** tener un contrato legal asociado (`contrato_id`/`documento_legal_id`); si no, responde 400. Si procede: marca el curso `estado='abierto'` y crea un registro en `course-cycles` (`course_id, nombre_cohorte, fecha_inicio, fecha_fin, capacidad, estado:'activa', creado_en`).
- **Go (histórico):** `POST /courses/:course_id/periods` (`backend/internal/course_periods/endpoints.go:80`). Era un alta simple (`fecha_inicio, fecha_fin, fecha_inscripcion`), sin guarda de estado previo del curso, sin campo de capacidad (`entities.CoursePeriod` no tenía `Capacity`), y sin relación con el estado de un contrato legal (que tampoco existe en Go — ver §5.2).
- **Resuelto:** el mismo endpoint (`POST /courses/:course_id/periods`) ahora exige `capacidad` (> 0) y agrega una guarda de negocio: solo se permite crear un período si el curso existe/está aprobado (`courses.ErrCourseNotFound` si no) **y** no tiene ya un período activo sin cerrar (`ErrCoursePeriodAlreadyOpen` en caso contrario). Esto cubre ambas ramas de la regla `canOpen` de Node sin necesitar el concepto de contrato: un curso recién aprobado no tiene períodos activos (pasa), y un curso previamente cerrado tampoco (pasa); uno con un período ya abierto es rechazado con 409. **No** depende de §5.2 (contrato) — ver §8.

### 2.2 Solicitud de cierre de cohorte con evidencias (🟡 Media) — ✅ Resuelto (parcial)

- **Node:** `POST /api/courses/[id]/closures` (`diplomados/src/app/api/courses/[id]/closures/route.ts`). Multipart con 3 archivos **obligatorios**: `archivo_participantes`, `archivo_vouchers`, `archivo_encuesta`, más `titulo_curso`, `nombre_cohorte`, `observaciones`. Al crear la solicitud de cierre (`course-cycle-closures`), además bloquea el curso poniendo `estado_gestion='solicitud-cierre'`.
- **Go (histórico):** `POST /course-cycle-close-requests` (`backend/internal/course_cycle_close_requests/endpoints.go:48`). Body JSON simple: `{course_cycle_id, observaciones}`. No exigía ni aceptaba adjuntos (participantes/vouchers/encuesta), y no había ningún bloqueo contra solicitudes duplicadas para el mismo período.
- **Resuelto:** el mismo endpoint ahora recibe `multipart/form-data` y exige los 3 archivos (`archivo_participantes`, `archivo_vouchers`, `archivo_encuesta`), que se suben a B2 y se persisten en la tabla `files` con `owner_type='course_cycle_close_request'`. Se agregó además una guarda de duplicados: si ya existe una solicitud de cierre `under_review` para el mismo `course_cycle_id`, la nueva es rechazada con `ErrCloseRequestAlreadyPending` (409).
- **Parcial — no resuelto:** a diferencia de Node (`estado_gestion='solicitud-cierre'` bloquea el curso completo mientras la revisión está pendiente), Go solo impide *solicitudes de cierre duplicadas* para el mismo período — no bloquea otras operaciones sobre el curso/período (p. ej. crear un nuevo período) mientras una solicitud está pendiente, porque eso requeriría un estado a nivel de curso que no existe hoy (ver §5.2). Se deja como posible trabajo futuro, no bloqueante para cerrar esta brecha en su alcance original.

### 2.3 Listado público filtrado por curso "listo" (🟡 Media)

- **Node:** `GET /api/courses/public` (`diplomados/src/app/api/courses/public/route.ts`). Filtra únicamente cursos con contrato legal firmado (`documento_legal_id`/`contrato_id`) y `estado ∈ {abierto, cerrado}` — pensado para el landing público.
- **Go:** `GET /courses` (`backend/internal/courses/endpoints.go:65`) es público pero no aplica ningún filtro de este tipo (y no hay campo de contrato para filtrar por — ver §5.2). Devuelve todos los cursos sin distinguir si están "listos" para mostrarse públicamente.
- **Impacto:** el listado público en Go puede exponer cursos aún no habilitados legalmente para dictarse.

### 2.4 Anuncios/publicaciones ligados a curso+cohorte, no solo a período (⚪ Baja)

- **Node:** `publications` (`GET/POST /api/publications`) se indexan por `course_id` **y** `cohort_id` de forma independiente, y en el detalle de curso (`GET /api/courses/[id]`) solo se muestran las del cohorte más reciente.
- **Go:** los anuncios (`Announcement`) cuelgan exclusivamente de `course_periods/:period_id/announcements` (`backend/internal/course_periods/endpoints.go:149-212`) — modelo equivalente en la práctica (un período ≈ un cohorte), solo difiere la forma de la clave. No requiere cambio funcional, se documenta como diferencia de modelado menor.

### 2.5 Endpoint de inscripción/participantes (🔴 Alta)

Ver el detalle completo en §5.1 — se incluye aquí como referencia cruzada porque también es, en esencia, una funcionalidad de "curso" (gestionar quién cursa un período).

---

## 3. Diplomados

**No existe un concepto de "diplomado" separado de "curso" en ninguno de los dos backends.**

- **Go:** `entities.RequestType_COURSE = "diplomado"` (`backend/internal/entities/entities.go:23`) — un `Course` **es** el diplomado; no hay módulo, tabla ni entidad `Diplomado` distinta.
- **Node:** no existe colección `diplomado` en `db.json`, ni ruta, ni servicio; todo se modela como `curso`/`course-requests`, con `tipo_curso ∈ {formulacion-curso-directa, formulacion-curso-indirecta}`.

Por lo tanto esta sección no aporta brechas nuevas de endpoints — las brechas reales de "diplomado" son las de cursos (§2) y solicitudes de curso (§5). Sí hay una diferencia de **forma de datos** a nivel de entidad/response que vale la pena señalar:

### 3.1 Campos de clasificación/evaluación ausentes en la entidad Course de Go (⚪ Baja — solapa con §5.3)

- **Node:** el registro de `course-requests`/`courses` incluye `clasificacion`, `calificacion`, `motivo_rechazo`, `archivo_evaluacion_url`, `contrato_id`/`documento_legal_id` directamente sobre el curso.
- **Go:** `Course`/`GetCourseResponse` (`backend/internal/courses/response.go`) no expone ninguno de estos campos; `motivo_rechazo`-equivalente (`Comments`) vive solo en `CourseRequest`, no en `Course`, y `clasificacion`/`calificacion`/contrato no existen en absoluto en el modelo (ver §5.3 y §5.2 para el detalle funcional).

---

## 4. Usuarios

Esta es la sección con **menor brecha relativa** — el módulo `auth` de Node también es mínimo (solo login), así que no hay que sobredimensionar las diferencias aquí.

### 4.1 Endpoints presentes en ambos lados, con diferencias de forma (⚪ Baja)

|          | Node                                                                    | Go                                                                                     |
| -------- | ----------------------------------------------------------------------- | -------------------------------------------------------------------------------------- |
| Registro | `POST /api/users` (multipart) — `diplomados/src/app/api/users/route.ts` | `POST /users` (multipart) — `backend/internal/users/endpoints.go:73`                   |
| Listado  | `GET /api/users?rol=`                                                   | `GET /users` (paginado, sin filtro por rol) — `backend/internal/users/endpoints.go:50` |
| Detalle  | `GET /api/users/[id]`                                                   | `GET /users/:id` — `backend/internal/users/endpoints.go:36`                            |
| Login    | `POST /api/auth/login`                                                  | `POST /auth/login` — `backend/internal/auth/endpoints.go:26`                           |

- **Nomenclatura de campos inconsistente dentro del propio Go:** `POST /users` usa `cedula, nombres` mientras `PUT /users/:id` usa `ci, primer_nombre` para conceptos equivalentes (`backend/internal/users/request.go` — comparar create vs. update). Es una inconsistencia interna, no causada por Node, pero se señala porque complica igualar el contrato con el cliente de `diplomados/` si algún día se apunta directo al backend Go.
- **Filtro por rol en listado de usuarios:** Node soporta `GET /api/users?rol=` (usado para poblar selects de "coordinadores", por ejemplo `getCoordinadores()` en `diplomados/src/servicios/users-service.js`). Go's `GET /users` no acepta un filtro por rol — impacto bajo, se puede filtrar client-side, pero es un endpoint usado activamente en el mock para poblar UI de administración.
- **Rol por defecto al registrar:** Node asigna `rol: 'visitante'` por defecto si no se especifica (`diplomados/src/app/api/users/route.ts`). Confirmar que el alta en Go asigna un rol por defecto equivalente y no deja el campo vacío.

### 4.2 Sin brecha de autenticación real

Ninguno de los dos lados tiene registro completo con verificación de email, logout, refresh token o recuperación de contraseña — Node lo simula con un JWT sin firmar y Go tiene DTOs de registro (`RegisterRequest`/`RegisterResponse` en `backend/internal/auth/request.go:8-11` y `response.go:14-16`) que **no están conectados a ningún handler** (código muerto). No se lista como brecha porque Node tampoco lo implementa — es una carencia compartida, fuera del alcance de este documento comparativo.

---

## 5. Admin — Solicitudes de curso ("course requests")

Esta es la sección central del documento: el flujo de aprobación de solicitudes de curso es donde el mock de Node modela más reglas de negocio que el backend Go aún no implementa.

### 5.1 Sin gestión de inscripción/participantes de un período (🔴 Alta)

- **Go:** `entities.CoursePeriod.Participants []User` existe como campo (`backend/internal/entities/course_period.go:6`), pero **no hay ningún endpoint HTTP** que lo lea o escriba — ni en `internal/courses` ni en `internal/course_periods` (confirmado por búsqueda directa en el código, sin resultados de un handler de inscripción).
- **Node:** tampoco tiene un endpoint explícito de inscripción de participantes individuales (el mock no llegó a modelar esa parte tampoco), pero sí registra `capacidad` por cohorte en `course-cycles` (ver §2.1) como paso previo a habilitar inscripciones.
- **Nota:** se marca como Alta porque es un campo ya modelado en la entidad Go pero completamente huérfano de endpoint — es la brecha con mayor riesgo de quedar "olvidada" al no aparecer en ningún inventario de rutas.

### 5.2 Sin paso de "estado legal"/contrato que finaliza la habilitación de un curso (🔴 Alta — la brecha más grande del documento) — ✅ Resuelto

- **Node:** `POST /api/legal-status/[user_id]` (`diplomados/src/app/api/legal-status/[user_id]/route.ts`). Es el paso que efectivamente **desbloquea** un curso aprobado: adjunta documentos legales (`carta_intencion` + `carta_compromiso` para el contrato inicial, o una `adenda` con `cursos_amparados[]` para ampliar cursos ya cubiertos por un contrato existente) al `provider`, y son estos documentos (`contrato_id`/`documento_legal_id`) los que habilitan que un curso `aprobada` pueda abrir cohorte (§2.1) o aparecer en el listado público (§2.3).
- **Go (histórico):** no existía ningún concepto de contrato/documento legal asociado a un `Provider` o `Course` como tal — existía una versión parcial y autoservicio (`POST /providers/documents` → `UploadProviderDocuments`) que subía `carta_intencion`/`carta_compromiso` y marcaba `courses.has_documentation=true` mediante una heurística de ventana de fechas (cursos creados entre la subida anterior y la actual de `carta_compromiso`), sin selección explícita de cursos y sin concepto de adenda.
- **Resuelto:** se reemplazó por completo ese flujo con un modelo explícito de contrato/adenda: nuevas tablas `deu.provider_contracts` (registro de cada envío: tipo `inicial`/`adenda`) y `deu.provider_contract_courses` (qué cursos quedaron amparados por cada envío), expuesto en `POST /admin/providers/:id/legal-contracts` (`backend/internal/providers/service.go` → `SubmitProviderContract`). El servidor infiere si el envío es el contrato inicial o una adenda según si el proveedor ya tiene un contrato `inicial` registrado (no se confía en un flag enviado por el cliente); cada envío ampara automáticamente **todos** los cursos aprobados del proveedor que aún no tuvieran cobertura legal (nunca una lista de cursos elegida por el cliente) — replicando el comportamiento real de la UI de Node (`profile-coordinator-review.tsx`), que ya cubre todos los "cursos sin contrato" sin selección individual. Amparar un curso solo marca `has_documentation=true` (ahora expuesto en `GET /courses` como `tiene_documentacion_legal`); a diferencia de Node, **no** se fuerza ningún cambio de `estado`/`estado_gestion` del curso — ese force-flip del mock se identificó como un atajo sin respaldo de negocio real, no como una regla a portar. Una adenda que no ampararía ningún curso nuevo se rechaza (`ErrNoCoursesToCoverage`, 409); el contrato inicial sí puede presentarse sin cursos que amparar (proveedor nuevo sin cursos aún).
- **Fuera de alcance de esta resolución:** §2.3 (filtro del listado público por curso "listo") sigue pendiente — ahora existe el campo real (`has_documentation`) sobre el cual construir ese filtro, pero conectarlo queda como trabajo de seguimiento separado y acotado.

### 5.3 Aprobación sin evaluación adjunta (🟡 Media)

- **Node:** `POST /api/course-requests/[id]/approve` (dentro de `[action]/route.ts`) acepta multipart con `calificacion`, `clasificacion`, y `archivo_evaluacion` (PDF), que quedan guardados en el propio registro de la solicitud.
- **Go:** `POST /admin/course-requests/:id/approve` (`backend/internal/course_requests/endpoints.go:44`) solo acepta `{tipo_curso, observaciones}` — no hay campo de calificación numérica, clasificación, ni adjunto de documento de evaluación.
- **Impacto:** se pierde la trazabilidad de *por qué* se aprobó un curso y con qué nota/documento de respaldo lo evaluó el comité.

### 5.4 Cierre de cohorte sin cascada de estado (🔴 Alta)

- **Node:** `POST /api/course-cycle-closures/[id]/approve` (`diplomados/src/app/api/course-cycle-closures/[id]/[action]/route.ts`). Al aprobar, además de marcar la propia solicitud de cierre, **cascada**: pone el curso en `estado_gestion='cerrado'` y marca **todas** sus cohortes activas (`course-cycles?estado=activa`) como `cerrada`. Al rechazar, revierte el curso a `estado_gestion='abierto'`.
- **Go:** `POST /admin/course-cycle-close-requests/:id/approve` (`backend/internal/course_cycle_close_requests/endpoints.go:72`) **sí es transaccional**: `PostgresRepository.ApproveCourseCycleCloseRequest` (`backend/internal/postgres_repository/postgres_course_cycle_close_requests_repository.go:84-113`) marca la solicitud `approved` **y**, en la misma transacción, ejecuta `queries.SetCourseCycleClosed(cycleID)` — que pone `course_cycles.closed_at = NOW()` e `is_active = false` para el ciclo/período específico de la solicitud. `.../reject` (`endpoints.go:94`) solo cambia el `Status` de la solicitud, sin revertir nada (no había nada bloqueado que revertir).
- **Corrección respecto a una versión anterior de este documento:** este ítem originalmente afirmaba que Go "no tocaba" el período al aprobar — eso era incorrecto; sí lo hace, a nivel del `course_cycle` específico. La brecha real, más acotada, es: (a) Go no tiene un estado a nivel de **curso** (`estado_gestion`) que reflejar, porque ese campo no existe en la entidad `Course` (ver §5.2/§3.1); y (b) Go solo cierra el ciclo referenciado por la solicitud, no **todas** las cohortes activas del curso como hace Node — en la práctica esto no debería diferir si (como ahora, tras la resolución de §2.1) solo puede existir un período activo por curso a la vez, pero el comportamiento no es una cascada explícita a nivel de curso como en Node.
- **Impacto:** menor de lo que se pensaba — el período sí se cierra correctamente al aprobar. Queda pendiente solo si en el futuro se modela un estado explícito a nivel de curso (`estado_gestion`), momento en el cual este approve debería actualizarlo también.

### 5.5 Sin bandeja de "mis solicitudes" para el solicitante (⚪ Baja)

- **Node:** el cliente arma esta vista agregando client-side contra los mismos endpoints de administración (no hay un endpoint dedicado tampoco en Node), filtrando por `usuario_id`/`coordinador_id`.
- **Go:** `GET /admin/course-requests` (`backend/internal/course_requests/endpoints.go:140`) es exclusivamente admin/faculty-scoped (vía `facultyscope.Resolve`) — no existe una variante para que un proveedor/coordinador consulte el estado de sus propias solicitudes enviadas.
- **Impacto:** bajo, ya que tampoco es un endpoint real en el mock, pero se señala porque es una necesidad de producto evidente en el flujo (el solicitante necesita saber en qué estado quedó su curso) que ninguno de los dos backends resuelve hoy con un endpoint propio.

### 5.6 Redirección de solicitud — ya cubierto en ambos lados (sin brecha)

`POST /admin/course-requests/:id/redirect` (Go, `backend/internal/course_requests/endpoints.go:104`) y la acción `remit` de `POST /api/course-requests/[id]/[action]` (Node) son equivalentes: reasignan la solicitud a otra facultad/coordinador. Se documenta aquí solo para confirmar que **no** es una brecha — Go ya tiene esta funcionalidad.

### 5.7 Nota sobre nomenclatura de roles (no es una brecha funcional)

`entities.Role` (`backend/internal/entities/user.go`) no define constantes para `deu_admin`/`faculty_admin`/`root`, aunque estos strings se usan de forma consistente y correcta como literales en `backend/cmd/deu/main.go:146`, `backend/internal/providers/endpoints.go:326-327` y `backend/internal/facultyscope/`. Node tiene el mismo patrón (roles como strings sueltos, sin un enum central). No se lista como brecha porque ambos lados comparten esta característica — se documenta únicamente como observación de higiene de código para quien trabaje en agregar nuevos roles de administración.

---

## 6. Solapamiento Curso + Grupo

**No existe ningún endpoint ni entidad que relacione cursos y grupos en ninguno de los dos backends hoy.**

- **Go:** `internal/activities` (`entities.Activity`) y `internal/group_resource_requests` (`entities.GroupResourceRequest`) son exclusivamente de grupos — ninguna de sus entidades tiene un campo `CourseID`/`CoursePeriodID`, confirmado por revisión directa.
- **Node:** las colecciones `extension-groups`, `group-requests`, `group-members`, `group-resource-requests` y `activities` existen en `db.json` pero están **vacías y sin ninguna ruta ni servicio que las use** — son remanentes de un modelo que nunca se llegó a implementar en el mock. La única referencia visible a "grupo" en la UI de administración (pestaña "Grupo de Extensión" en `diplomados/src/app/admin/usuarios/page.tsx` y `solicitudes-table.tsx`) está respaldada por arrays hardcodeados en el frontend, no por datos reales.

No hay, por lo tanto, brechas que documentar en esta sección — se deja constancia de que la revisión se hizo y no encontró superficie compartida real en ninguno de los dos sistemas.

---

## 7. Dependencias

Algunos ítems de este documento no se pueden resolver de forma aislada — dependen de que otro ítem exista primero. Esta sección deja constancia explícita de esas dependencias para evitar que se aborden en el orden equivocado.

- **§2.3 (listado público filtrado por curso "listo") depende de §5.2 (paso de "estado legal"/contrato).** La regla de Node para `GET /api/courses/public` filtra explícitamente por la presencia de un contrato legal (`contrato_id`/`documento_legal_id`). Go no tiene ningún concepto de contrato/documento legal asociado a un curso o proveedor (§5.2), así que no existe ningún campo real sobre el cual replicar ese filtro. Implementar §2.3 con un proxy inventado (p. ej. "aprobado y con al menos un período histórico") sería una interpretación de producto no confirmada, no una migración fiel de la regla de Node — por eso se decidió **no** implementar §2.3 todavía y esperar a que §5.2 exista primero.
- **§2.1 (apertura de período) NO depende de §5.2.** A diferencia de §2.3, la guarda de apertura de período implementada no necesita el concepto de contrato: basta con que el curso esté aprobado (`courses.is_active`, ya existente) y no tenga un período activo sin cerrar. Se señala explícitamente para que quede claro que §2.1 pudo resolverse de forma independiente y ya está hecho (ver §2.1).
- **§5.4 (cascada de cierre) queda parcialmente ligada a §5.2/§3.1.** Como se detalla en §5.4, Go ya cierra correctamente el período/ciclo específico al aprobar una solicitud de cierre; lo único que falta es reflejar ese cierre en un estado a nivel de **curso** (`estado_gestion`), que hoy no existe en la entidad `Course`. Modelar ese campo probablemente ocurra junto con — o como parte de — el trabajo de §5.2, ya que ambos requieren extender `entities.Course` con nuevos campos de estado/ciclo de vida.

---

## 8. Tabla resumen

| #   | Brecha                                                                                                                         | Sección          | Prioridad | Estado |
| --- | ------------------------------------------------------------------------------------------------------------------------------ | ---------------- | --------- | ------ |
| 1   | Sin paso de "estado legal"/contrato que habilita un curso aprobado                                                             | §5.2             | 🔴 Alta    | ✅ Resuelto |
| 2   | Sin cascada de cierre de cohorte a nivel de curso (el período sí se cierra; falta reflejarlo en un estado de curso inexistente) | §5.4             | 🟡 Media   | Pendiente (alcance reducido, ver §5.4) |
| 3   | Sin endpoint de inscripción/gestión de participantes de un período                                                             | §5.1 (ref. §2.5) | 🔴 Alta    | Pendiente |
| 4   | Apertura de cohorte sin reglas de estado ni campo de capacidad                                                                 | §2.1             | 🔴 Alta    | ✅ Resuelto |
| 5   | Aprobación de solicitud sin calificación/clasificación/documento de evaluación                                                 | §5.3             | 🟡 Media   | Pendiente |
| 6   | Solicitud de cierre sin adjuntos obligatorios (participantes/vouchers/encuesta) ni bloqueo del período mientras está pendiente | §2.2             | 🟡 Media   | ✅ Resuelto (adjuntos + guarda de duplicados; bloqueo a nivel de curso queda pendiente, ver §2.2) |
| 7   | Listado público de cursos sin filtro por curso "listo" (contrato + estado)                                                     | §2.3             | 🟡 Media   | Bloqueado — depende de #1 (ver §7) |
| 8   | Sin bandeja de "mis solicitudes" para el solicitante                                                                           | §5.5             | ⚪ Baja    | Pendiente |
| 9   | Sin filtro por rol en listado de usuarios                                                                                      | §4.1             | ⚪ Baja    | Pendiente |
| 10  | Inconsistencia de nombres de campo entre alta y edición de usuario (`cedula`/`nombres` vs. `ci`/`primer_nombre`)               | §4.1             | ⚪ Baja    | Pendiente |
| 11  | Campos de clasificación/evaluación/contrato ausentes en la entidad `Course`                                                    | §3.1             | ⚪ Baja    | Pendiente |
| 12  | Modelado de anuncios por período vs. curso+cohorte independientes                                                              | §2.4             | ⚪ Baja    | Pendiente |

**Sin brechas:** redirección/remit de solicitudes (§5.6), autenticación básica (§4.2), solapamiento curso+grupo (§6).
