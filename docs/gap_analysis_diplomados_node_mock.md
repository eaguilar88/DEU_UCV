# Análisis de brechas: Backend Go vs. mock Node de `diplomados/`

## 1. Introducción

`diplomados/` (`ecp_ucv`) es la app Next.js del portal de cursos/diplomados. Durante su desarrollo temprano incluyó un backend simulado propio: rutas de API de Next.js (`src/app/api/**`) que actúan como una capa fina de lógica de negocio delante de un almacén REST crudo servido con `json-server --watch db.json`. Ese mock se desarrolló en paralelo (y en algunos puntos, por delante) del backend real en Go, y por eso codifica un flujo de aprobación más completo del que existe hoy en `backend/internal/`.

Este documento compara ambos backends y documenta **en una sola dirección**: qué existe en el mock de Node que falta (total o parcialmente) en el backend Go. No se documentan capacidades que el backend Go tiene y el mock no — eso queda fuera del alcance.

**Alcance:** cursos, diplomados, usuarios, y los endpoints de administración relacionados con solicitudes de curso ("course requests"). Los endpoints exclusivos de grupos se excluyen; ver §6 para la sección de solapamiento curso+grupo.

**Método:** lectura directa del código de rutas/handlers en `diplomados/src/app/api/**` + esquemas de `db.json`, comparado contra `backend/internal/**` + el enrutamiento de `backend/cmd/deu/main.go`.

**Última re-evaluación:** 2026-09-28, sobre el estado del backend en `c3788d0` (incluye `0fc656c` — `estado_gestion`, `GET /courses/public`, `GET /course-requests` — y `c3788d0` — transiciones `abierto`/`cerrado`). Resultado: §2.2, §2.3, §5.4 y §5.5 pasan a resueltos; §5.1 se reclasifica; se agregan §2.6–§2.9 y §5.8 como brechas no documentadas antes.

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

### 2.2 Solicitud de cierre de cohorte con evidencias (🟡 Media) — ✅ Resuelto

- **Node:** `POST /api/courses/[id]/closures` (`diplomados/src/app/api/courses/[id]/closures/route.ts`). Multipart con 3 archivos **obligatorios**: `archivo_participantes`, `archivo_vouchers`, `archivo_encuesta`, más `titulo_curso`, `nombre_cohorte`, `observaciones`. Al crear la solicitud de cierre (`course-cycle-closures`), además bloquea el curso poniendo `estado_gestion='solicitud-cierre'`.
- **Go (histórico):** `POST /course-cycle-close-requests` (`backend/internal/course_cycle_close_requests/endpoints.go:48`). Body JSON simple: `{course_cycle_id, observaciones}`. No exigía ni aceptaba adjuntos (participantes/vouchers/encuesta), y no había ningún bloqueo contra solicitudes duplicadas para el mismo período.
- **Resuelto:** el mismo endpoint ahora recibe `multipart/form-data` y exige los 3 archivos (`archivo_participantes`, `archivo_vouchers`, `archivo_encuesta`), que se suben a B2 y se persisten en la tabla `files` con `owner_type='course_cycle_close_request'`. Se agregó además una guarda de duplicados: si ya existe una solicitud de cierre `under_review` para el mismo `course_cycle_id`, la nueva es rechazada con `ErrCloseRequestAlreadyPending` (409).
- **Bloqueo a nivel de curso — resuelto en `0fc656c`:** se agregó la columna `courses.estado_gestion` (`entities.CourseManagementStatus`, migraciones `000025`/`000026`). `PostgresRepository.CreateCourseCycleCloseRequest` ahora inserta la solicitud **y** pone `estado_gestion='solicitud-cierre'` en la misma transacción, y `course_periods.CreateCoursePeriod` rechaza abrir un período nuevo mientras el curso esté en ese estado (`courses.ErrCourseClosureRequestPending`, 409). Esto replica el bloqueo de Node.

### 2.3 Listado público filtrado por curso "listo" (🟡 Media) — ✅ Resuelto

- **Node:** `GET /api/courses/public` (`diplomados/src/app/api/courses/public/route.ts`). Filtra únicamente cursos con contrato legal firmado (`documento_legal_id`/`contrato_id`) y `estado ∈ {abierto, cerrado}` — pensado para el landing público.
- **Go (histórico):** `GET /courses` era público pero no aplicaba ningún filtro de este tipo.
- **Resuelto en `0fc656c`:** nuevo endpoint público `GET /courses/public` (`backend/internal/courses/endpoints.go` → `GetPublicCourses`; query `queries.GetPublicCourses`). Filtra `is_active=true` (aprobado) **y** `has_documentation=true` (amparado por contrato, §5.2) **y** que exista al menos un período no borrado (el curso se abrió alguna vez).
- **Diferencia residual menor (⚪):** Go usa "existe un período" como proxy de `estado ∈ {abierto, cerrado}` en lugar de leer `estado_gestion`, que ya existe. La consecuencia es que un curso en `solicitud-cierre` **sí** aparece en el listado público de Go, mientras que Node lo oculta (su filtro solo acepta `abierto`/`cerrado`). Si se quiere paridad exacta, el filtro podría pasar a `estado_gestion IN ('abierto','cerrado')`; no se considera bloqueante porque ocultar un curso mientras se revisa su cierre parece más un efecto colateral del mock que una regla de negocio.

### 2.4 Anuncios/publicaciones ligados a curso+cohorte, no solo a período (⚪ Baja)

- **Node:** `publications` (`GET/POST /api/publications`) se indexan por `course_id` **y** `cohort_id` de forma independiente, y en el detalle de curso (`GET /api/courses/[id]`) solo se muestran las del cohorte más reciente.
- **Go:** los anuncios (`Announcement`) cuelgan exclusivamente de `course_periods/:period_id/announcements` (`backend/internal/course_periods/endpoints.go:149-212`) — modelo equivalente en la práctica (un período ≈ un cohorte), solo difiere la forma de la clave. No requiere cambio funcional, se documenta como diferencia de modelado menor.

### 2.6 Facultad/coordinador de origen inmutable tras una redirección (🟡 Media) — ✅ Resuelto

- **Node:** al crear el curso (`POST /api/courses`, `diplomados/src/app/api/courses/route.ts`) guarda dos pares de campos: `facultad`/`coordinador_id` (dueño **actual**, que cambia con `remit`) y `facultad_origen`/`coordinador_origen` (el "anfitrión inmutable"). Los usa de forma distinta según el trámite:
  - la revisión de la solicitud va a la facultad **actual** (`admin/courses/route.ts`: `isMyTurn = isUnderReview && facultadCurso === miFacultad`);
  - la bandeja de "aprobados sin contrato" va a la facultad de **origen** (`isMyProviderMissingContract = … && facultadOrigen === miFacultad`);
  - la solicitud de cierre de cohorte se envía al coordinador de **origen** (`courses/[id]/closures/route.ts`: `coordinador_id: course.coordinador_origen || course.coordinador_id`).
- **Go:** hay una sola columna `courses.faculty`. `PostgresRepository.RedirectCourseRequest` (`backend/internal/postgres_repository/postgres_course_requests_repository.go:210`) la **sobrescribe** con `queries.UpdateCourseFaculty`, sin conservar la original. `GET /admin/course-cycle-close-requests` filtra por `c.faculty` (`queries.GetCourseCycleCloseRequests`), es decir, por la facultad actual.
- **Impacto:** después de redirigir un curso, sus solicitudes de cierre llegan a la facultad a la que se redirigió y no a la facultad anfitriona, al revés de lo que hace Node. Resolverlo requiere guardar la facultad de origen (columna nueva o tomarla del proveedor, ver §2.7) y usarla para dirigir los cierres.
- **Resuelto:** nueva columna `courses.origin_faculty` (migración `000032`; las filas existentes se rellenan con `faculty`, así que un curso que ya se había redirigido queda con la facultad actual como origen). Se fija al crear el curso con la misma facultad heredada del proveedor (§2.7) y ninguna operación la modifica: `RedirectCourseRequest` sigue cambiando solo `faculty`. No se reutilizó `providers.faculty` porque se puede editar (`UpdateProvider`). Uso, igual que en Node:
  - la bandeja de solicitudes de curso (`GET /admin/course-requests`) sigue filtrando por la facultad **actual** (`c.faculty`);
  - la bandeja de cierres (`GET /admin/course-cycle-close-requests`) ahora filtra por la facultad de **origen** (`c.origin_faculty`).

  Se expone como `facultad_origen` en `GetCourseResponse`. `coordinador_origen` no se porta: Go dirige las bandejas por facultad, no por ID de coordinador.

### 2.7 Facultad del curso enviada por el cliente en vez de heredada del proveedor (🟡 Media) — ✅ Resuelto

- **Node:** `POST /api/courses` no acepta la facultad del cliente. La **hereda** del proveedor: busca el `provider` del usuario, toma su `coordinador_id` y de ahí la `facultad` del coordinador (`admin/courses/route.ts` hace lo mismo leyendo `provider.facultad` directamente).
- **Go:** `POST /courses` (`backend/internal/courses/decoders.go:32` → `toCourseEntity`) exige un campo de formulario `facultad` y lo usa tal cual. `courses.service.CreateCourse` ya carga el proveedor (`GetProviderByUserID`) y `entities.Provider.Faculty` existe, pero solo se usa para `Owner.ID`.
- **Impacto:** un proveedor puede enviar su solicitud a cualquier facultad eligiéndola en el formulario, y así saltarse la facultad a la que está adscrito. El arreglo es acotado: tomar `course.Faculty = provider.Faculty` en el servicio y dejar de exigir el campo.
- **Resuelto:** `POST /courses` ya no lee `facultad`; si el cliente la envía, se ignora. `courses.service.CreateCourse` asigna `Faculty` y `OriginFaculty` desde `provider.Faculty`; si el proveedor no tiene facultad, usa `DEU` (el valor por defecto de la columna). `PUT /courses/:id` también dejó de aceptar `facultad`: Node no tiene ninguna ruta que cambie la facultad de un curso, y en Go la actualización la estaba poniendo en `NULL` porque el decoder no la mapeaba. Ahora la facultad solo cambia al crear el curso o al redirigir la solicitud.

### 2.8 `GET /courses` sin filtros por dueño/estado ni visibilidad por rol (⚪ Baja) — ✅ Resuelto

- **Node:** `GET /api/courses` acepta `usuario_id`, `codigo_proveedor` y `estado`. Si quien consulta no es admin/coordinador ni dueño, solo devuelve cursos en `abierto`/`cerrado`.
- **Go:** `GET /courses` (`backend/internal/courses/endpoints.go:65`) es público, solo pagina, y devuelve todo curso con `is_active=true` (aprobado), tenga contrato o no, sin importar quién consulta.
- **Mitigación existente:** el landing ya puede usar `GET /courses/public` (§2.3), y el proveedor ya puede ver lo suyo con `GET /course-requests` (§5.5). Lo que falta es solo el filtro por `estado`/proveedor sobre el listado general y, si importa, dejar de exponer públicamente cursos aprobados sin contrato por `GET /courses`.
- **Resuelto:** `GET /courses` ahora tiene auth opcional (`OptionalJWTMiddleware`, igual que grupos y usuarios) y acepta los mismos filtros que Node:
  - `usuario_id`: cursos del proveedor cuyo usuario es ese;
  - `codigo_proveedor`: código del proveedor;
  - `estado`: `abierto`, `cerrado` o `solicitud-cierre`. Un valor distinto devuelve 400.

  La visibilidad sigue la regla de Node. Si quien consulta no es admin (`root`, `deu_admin` o `faculty_admin`, el equivalente del "coordinador" de Node) ni pide sus propios cursos (`usuario_id` = su ID), solo ve cursos con `estado_gestion` en `abierto`/`cerrado`. La regla vive en `courses.service.GetCourses`, no en el handler. Diferencias con Node:
  - se sigue devolviendo solo cursos aprobados (`is_active=true`); en Node la colección `courses` tampoco contiene solicitudes pendientes;
  - la paginación se aplica siempre, también con `usuario_id`/`codigo_proveedor` (Node la omite en ese caso).

### 2.9 Cohorte sin nombre (⚪ Baja) — ✅ Resuelto

- **Node:** cada `course-cycles` tiene `nombre_cohorte` (viene de `cohortName` en `POST /api/courses/[id]/open`), y la solicitud de cierre lo repite en `payload.nombre_cohorte`.
- **Go:** `entities.CoursePeriod` (`backend/internal/entities/course_period.go`) no tiene campo de nombre; un período se identifica solo por ID y fechas.
- **Impacto:** cosmético/UX: la UI de `diplomados/` muestra el nombre de la cohorte en el detalle del curso y en la bandeja de cierres.
- **Resuelto:** nueva columna `course_cycles.name` (migración `000033`; los períodos existentes quedan con nombre vacío). `POST` y `PUT /courses/:course_id/periods` exigen `nombre_cohorte`, como el formulario de Node (`gestion-cohorte-form.tsx`). Un nombre vacío o de solo espacios se rechaza con 400 (`ErrCohortNameRequired`, validado en el servicio igual que `capacidad`). Se expone como `nombre_cohorte` en los períodos y en `ultimo_periodo` del detalle del curso. La solicitud de cierre no repite el nombre (Node lo copia en `payload.nombre_cohorte`): en Go la solicitud ya referencia el período por `course_cycle_id`, y de ahí se obtiene el nombre.

---

## 3. Diplomados

**No existe un concepto de "diplomado" separado de "curso" en ninguno de los dos backends.**

- **Go:** `entities.RequestType_COURSE = "diplomado"` (`backend/internal/entities/entities.go:23`) — un `Course` **es** el diplomado; no hay módulo, tabla ni entidad `Diplomado` distinta.
- **Node:** no existe colección `diplomado` en `db.json`, ni ruta, ni servicio; todo se modela como `curso`/`course-requests`, con `tipo_curso ∈ {formulacion-curso-directa, formulacion-curso-indirecta}`.

Por lo tanto esta sección no aporta brechas nuevas de endpoints — las brechas reales de "diplomado" son las de cursos (§2) y solicitudes de curso (§5). Sí hay una diferencia de **forma de datos** a nivel de entidad/response que vale la pena señalar:

### 3.1 Campos de clasificación/evaluación ausentes en la entidad Course de Go (⚪ Baja — solapa con §5.3)

- **Node:** el registro de `course-requests`/`courses` incluye `clasificacion`, `calificacion`, `motivo_rechazo`, `archivo_evaluacion_url`, `contrato_id`/`documento_legal_id` directamente sobre el curso.
- **Go:** `Course`/`GetCourseResponse` (`backend/internal/courses/response.go`) no expone ninguno de estos campos; `motivo_rechazo`-equivalente (`Comments`) vive solo en `CourseRequest`, no en `Course`, y `clasificacion`/`calificacion`/contrato no existen en absoluto en el modelo (ver §5.3 y §5.2 para el detalle funcional).
- **Actualización (2026-09-28):** la parte de contrato y estado ya está cubierta. `GetCourseResponse` expone `tiene_documentacion_legal` (§5.2) y `estado_gestion` (`abierto`/`cerrado`/`solicitud-cierre`, §2.2/§5.4). Faltan solo `clasificacion`, `calificacion` y `archivo_evaluacion_url`, que dependen de §5.3.

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

### 5.1 Sin gestión de inscripción/participantes de un período (⚪ Fuera de alcance — no es brecha frente a Node; antes 🔴 Alta)

- **Go:** `entities.CoursePeriod.Participants []User` existe como campo (`backend/internal/entities/course_period.go:6`), pero **no hay ningún endpoint HTTP** que lo lea o escriba — ni en `internal/courses` ni en `internal/course_periods` (confirmado por búsqueda directa en el código, sin resultados de un handler de inscripción).
- **Node:** tampoco tiene un endpoint explícito de inscripción de participantes individuales (el mock no llegó a modelar esa parte tampoco), pero sí registra `capacidad` por cohorte en `course-cycles` (ver §2.1) como paso previo a habilitar inscripciones.
- **Nota:** se marca como Alta porque es un campo ya modelado en la entidad Go pero completamente huérfano de endpoint — es la brecha con mayor riesgo de quedar "olvidada" al no aparecer en ningún inventario de rutas.
- **Reclasificación (2026-09-28):** este documento lista solo lo que Node tiene y Go no (§1). Node tampoco inscribe participantes, y la única pieza que sí tiene (`capacidad` por cohorte) ya existe en Go desde §2.1. Por eso deja de contarse como brecha de migración. Sigue siendo trabajo pendiente de producto (y `CoursePeriod.Participants` sigue sin endpoint, confirmado de nuevo), pero debe planificarse aparte, no como parte de la migración del mock.

### 5.2 Sin paso de "estado legal"/contrato que finaliza la habilitación de un curso (🔴 Alta — la brecha más grande del documento) — ✅ Resuelto

- **Node:** `POST /api/legal-status/[user_id]` (`diplomados/src/app/api/legal-status/[user_id]/route.ts`). Es el paso que efectivamente **desbloquea** un curso aprobado: adjunta documentos legales (`carta_intencion` + `carta_compromiso` para el contrato inicial, o una `adenda` con `cursos_amparados[]` para ampliar cursos ya cubiertos por un contrato existente) al `provider`, y son estos documentos (`contrato_id`/`documento_legal_id`) los que habilitan que un curso `aprobada` pueda abrir cohorte (§2.1) o aparecer en el listado público (§2.3).
- **Go (histórico):** no existía ningún concepto de contrato/documento legal asociado a un `Provider` o `Course` como tal — existía una versión parcial y autoservicio (`POST /providers/documents` → `UploadProviderDocuments`) que subía `carta_intencion`/`carta_compromiso` y marcaba `courses.has_documentation=true` mediante una heurística de ventana de fechas (cursos creados entre la subida anterior y la actual de `carta_compromiso`), sin selección explícita de cursos y sin concepto de adenda.
- **Resuelto:** se reemplazó por completo ese flujo con un modelo explícito de contrato/adenda: nuevas tablas `deu.provider_contracts` (registro de cada envío: tipo `inicial`/`adenda`) y `deu.provider_contract_courses` (qué cursos quedaron amparados por cada envío), expuesto en `POST /admin/providers/:id/legal-contracts` (`backend/internal/providers/service.go` → `SubmitProviderContract`). El servidor infiere si el envío es el contrato inicial o una adenda según si el proveedor ya tiene un contrato `inicial` registrado (no se confía en un flag enviado por el cliente); cada envío ampara automáticamente **todos** los cursos aprobados del proveedor que aún no tuvieran cobertura legal (nunca una lista de cursos elegida por el cliente) — replicando el comportamiento real de la UI de Node (`profile-coordinator-review.tsx`), que ya cubre todos los "cursos sin contrato" sin selección individual. Amparar un curso solo marca `has_documentation=true` (ahora expuesto en `GET /courses` como `tiene_documentacion_legal`); a diferencia de Node, **no** se fuerza ningún cambio de `estado`/`estado_gestion` del curso — ese force-flip del mock se identificó como un atajo sin respaldo de negocio real, no como una regla a portar. Una adenda que no ampararía ningún curso nuevo se rechaza (`ErrNoCoursesToCoverage`, 409); el contrato inicial sí puede presentarse sin cursos que amparar (proveedor nuevo sin cursos aún).
- **Fuera de alcance de esta resolución:** §2.3 (filtro del listado público por curso "listo") sigue pendiente — ahora existe el campo real (`has_documentation`) sobre el cual construir ese filtro, pero conectarlo queda como trabajo de seguimiento separado y acotado.

### 5.3 Aprobación sin evaluación adjunta (🟡 Media)

- **Node:** `POST /api/course-requests/[id]/approve` (dentro de `[action]/route.ts`) acepta multipart con `calificacion`, `clasificacion`, y `archivo_evaluacion` (PDF), que quedan guardados en el propio registro de la solicitud.
- **Go:** `POST /admin/course-requests/:id/approve` (`backend/internal/course_requests/endpoints.go:44`) solo acepta `{tipo_curso, observaciones}` — no hay campo de calificación numérica, clasificación, ni adjunto de documento de evaluación.
- **Impacto:** se pierde la trazabilidad de *por qué* se aprobó un curso y con qué nota/documento de respaldo lo evaluó el comité.

### 5.4 Cierre de cohorte sin cascada de estado (🔴 Alta) — ✅ Resuelto

- **Node:** `POST /api/course-cycle-closures/[id]/approve` (`diplomados/src/app/api/course-cycle-closures/[id]/[action]/route.ts`). Al aprobar, además de marcar la propia solicitud de cierre, **cascada**: pone el curso en `estado_gestion='cerrado'` y marca **todas** sus cohortes activas (`course-cycles?estado=activa`) como `cerrada`. Al rechazar, revierte el curso a `estado_gestion='abierto'`.
- **Go:** `POST /admin/course-cycle-close-requests/:id/approve` (`backend/internal/course_cycle_close_requests/endpoints.go:72`) **sí es transaccional**: `PostgresRepository.ApproveCourseCycleCloseRequest` (`backend/internal/postgres_repository/postgres_course_cycle_close_requests_repository.go:84-113`) marca la solicitud `approved` **y**, en la misma transacción, ejecuta `queries.SetCourseCycleClosed(cycleID)` — que pone `course_cycles.closed_at = NOW()` e `is_active = false` para el ciclo/período específico de la solicitud. `.../reject` (`endpoints.go:94`) solo cambia el `Status` de la solicitud, sin revertir nada (no había nada bloqueado que revertir).
- **Corrección respecto a una versión anterior de este documento:** este ítem originalmente afirmaba que Go "no tocaba" el período al aprobar — eso era incorrecto; sí lo hace, a nivel del `course_cycle` específico. La brecha real, más acotada, es: (a) Go no tiene un estado a nivel de **curso** (`estado_gestion`) que reflejar, porque ese campo no existe en la entidad `Course` (ver §5.2/§3.1); y (b) Go solo cierra el ciclo referenciado por la solicitud, no **todas** las cohortes activas del curso como hace Node — en la práctica esto no debería diferir si (como ahora, tras la resolución de §2.1) solo puede existir un período activo por curso a la vez, pero el comportamiento no es una cascada explícita a nivel de curso como en Node.
- **Impacto:** menor de lo que se pensaba — el período sí se cierra correctamente al aprobar. Queda pendiente solo si en el futuro se modela un estado explícito a nivel de curso (`estado_gestion`), momento en el cual este approve debería actualizarlo también.
- **Resuelto en `0fc656c` + `c3788d0`:** ya existe `courses.estado_gestion`, y todas las transiciones de Node están replicadas de forma transaccional:
  - abrir período (`CreateCoursePeriod`) → `estado_gestion='abierto'`;
  - enviar solicitud de cierre → `'solicitud-cierre'` (§2.2);
  - aprobar cierre → `'cerrado'`; además, `queries.SetCourseCycleClosed` ahora cierra **todos** los ciclos activos del curso (`course_id = (SELECT course_id …) AND is_active`), no solo el de la solicitud;
  - rechazar cierre → vuelve a `'abierto'`.

  Esto cubre los dos puntos pendientes: (a) el estado a nivel de curso y (b) la cascada explícita sobre todas las cohortes activas.

### 5.5 Sin bandeja de "mis solicitudes" para el solicitante (⚪ Baja) — ✅ Resuelto

- **Node:** el cliente arma esta vista agregando client-side contra los mismos endpoints de administración (no hay un endpoint dedicado tampoco en Node), filtrando por `usuario_id`/`coordinador_id`.
- **Go:** `GET /admin/course-requests` (`backend/internal/course_requests/endpoints.go:140`) es exclusivamente admin/faculty-scoped (vía `facultyscope.Resolve`) — no existe una variante para que un proveedor/coordinador consulte el estado de sus propias solicitudes enviadas.
- **Impacto:** bajo, ya que tampoco es un endpoint real en el mock, pero se señala porque es una necesidad de producto evidente en el flujo (el solicitante necesita saber en qué estado quedó su curso) que ninguno de los dos backends resuelve hoy con un endpoint propio.
- **Resuelto en `0fc656c`:** nuevo endpoint protegido `GET /course-requests` (`backend/internal/course_requests/endpoints.go` → `GetMyCourseRequests`). Resuelve el proveedor del usuario autenticado (`GetProviderByUserID`) y devuelve, paginadas, las solicitudes de sus cursos (`queries.GetCourseRequestsByProvider`). Usa el mismo formato de respuesta que el listado admin. Si el usuario no es proveedor, devuelve una lista vacía.

### 5.6 Redirección de solicitud — ya cubierto en ambos lados (sin brecha)

`POST /admin/course-requests/:id/redirect` (Go, `backend/internal/course_requests/endpoints.go:104`) y la acción `remit` de `POST /api/course-requests/[id]/[action]` (Node) son equivalentes: reasignan la solicitud a otra facultad/coordinador. Se documenta aquí solo para confirmar que **no** es una brecha — Go ya tiene esta funcionalidad.

### 5.7 Nota sobre nomenclatura de roles (no es una brecha funcional)

`entities.Role` (`backend/internal/entities/user.go`) no define constantes para `deu_admin`/`faculty_admin`/`root`, aunque estos strings se usan de forma consistente y correcta como literales en `backend/cmd/deu/main.go:146`, `backend/internal/providers/endpoints.go:326-327` y `backend/internal/facultyscope/`. Node tiene el mismo patrón (roles como strings sueltos, sin un enum central). No se lista como brecha porque ambos lados comparten esta característica — se documenta únicamente como observación de higiene de código para quien trabaje en agregar nuevos roles de administración.

### 5.8 Bandejas de admin sin filtro por estado ni bandeja de "aprobados sin contrato" (🟡 Media)

- **Node:**
  - `GET /api/admin/courses` (`diplomados/src/app/api/admin/courses/route.ts`) no es un listado plano. Arma una bandeja de trabajo con dos grupos: solicitudes `under_review` y cursos `aprobado` **sin contrato** (`!contrato_id && !documento_legal_id`). Excluye rechazados y redirigidos. El coordinador ve el primer grupo según la facultad actual y el segundo según la facultad de origen (ver §2.6).
  - `GET /api/admin/closures` filtra por `estado`, y por defecto muestra solo `under_review`.
- **Go:**
  - `GET /admin/course-requests` (`backend/internal/course_requests/endpoints.go:147`) y `GET /admin/course-cycle-close-requests` (`backend/internal/course_cycle_close_requests/endpoints.go:124`) aceptan solo `faculty` y paginación, y devuelven solicitudes en **todos** los estados.
  - No hay forma de listar "cursos aprobados que aún no tienen contrato", que es justamente la lista que el admin necesita para saber a qué proveedores enviarles `POST /admin/providers/:id/legal-contracts` (§5.2). Por dentro esa consulta ya existe (`provider_contracts_queries.go` filtra `has_documentation=false` para amparar cursos), pero no está expuesta como listado. `GET /admin/providers` tampoco tiene un filtro equivalente.
- **Impacto:** el frontend tendría que traer todo y filtrar del lado del cliente, y para la bandeja de contratos pendientes directamente no hay de dónde sacar la información. Arreglo sugerido: agregar `?status=` a ambos listados admin y un filtro del tipo `has_documentation=false` (sobre cursos aprobados o sobre proveedores).

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

**Estado al 2026-09-28:** las tres dependencias de arriba ya están saldadas. §5.2 se resolvió y con eso §2.3 se pudo implementar sobre `has_documentation`. `estado_gestion` se modeló (`0fc656c`/`c3788d0`) y con eso se cerraron §5.4 y el bloqueo pendiente de §2.2. Dependencias vigentes entre las brechas abiertas:

- **§5.8 (bandeja de "aprobados sin contrato") dependía de §2.6 (facultad de origen)** para la parte de coordinadores: Node asigna esa bandeja por facultad de origen. §2.6 ya está resuelta, así que §5.8 puede filtrar por `courses.origin_faculty`.
- **§2.6 y §2.7 se resolvieron juntas.** Al final sí se agregó la columna `courses.origin_faculty`, porque `providers.faculty` se puede editar y no sirve como origen inmutable.
- **§3.1 depende de §5.3:** los campos restantes de `Course` (`clasificacion`, `calificacion`, evaluación) solo existen si se capturan al aprobar.

---

## 8. Tabla resumen

| #   | Brecha                                                                                                                         | Sección          | Prioridad | Estado |
| --- | ------------------------------------------------------------------------------------------------------------------------------ | ---------------- | --------- | ------ |
| 1   | Sin paso de "estado legal"/contrato que habilita un curso aprobado                                                             | §5.2             | 🔴 Alta    | ✅ Resuelto |
| 2   | Sin cascada de cierre de cohorte a nivel de curso (el período sí se cierra; falta reflejarlo en un estado de curso inexistente) | §5.4             | 🟡 Media   | ✅ Resuelto (`estado_gestion` + cierre de todas las cohortes activas) |
| 3   | Sin endpoint de inscripción/gestión de participantes de un período                                                             | §5.1 (ref. §2.5) | ⚪ —       | Fuera de alcance — Node tampoco lo tiene (ver §5.1) |
| 4   | Apertura de cohorte sin reglas de estado ni campo de capacidad                                                                 | §2.1             | 🔴 Alta    | ✅ Resuelto |
| 5   | Aprobación de solicitud sin calificación/clasificación/documento de evaluación                                                 | §5.3             | 🟡 Media   | Pendiente |
| 6   | Solicitud de cierre sin adjuntos obligatorios (participantes/vouchers/encuesta) ni bloqueo del período mientras está pendiente | §2.2             | 🟡 Media   | ✅ Resuelto (adjuntos + guarda de duplicados + bloqueo `solicitud-cierre`) |
| 7   | Listado público de cursos sin filtro por curso "listo" (contrato + estado)                                                     | §2.3             | 🟡 Media   | ✅ Resuelto (`GET /courses/public`; diferencia menor con `solicitud-cierre`, ver §2.3) |
| 8   | Sin bandeja de "mis solicitudes" para el solicitante                                                                           | §5.5             | ⚪ Baja    | ✅ Resuelto (`GET /course-requests`) |
| 9   | Sin filtro por rol en listado de usuarios                                                                                      | §4.1             | ⚪ Baja    | Pendiente |
| 10  | Inconsistencia de nombres de campo entre alta y edición de usuario (`cedula`/`nombres` vs. `ci`/`primer_nombre`)               | §4.1             | ⚪ Baja    | Pendiente |
| 11  | Campos de clasificación/evaluación/contrato ausentes en la entidad `Course`                                                    | §3.1             | ⚪ Baja    | Parcial — contrato y estado ya expuestos; clasificación/evaluación dependen de #5 |
| 12  | Modelado de anuncios por período vs. curso+cohorte independientes                                                              | §2.4             | ⚪ Baja    | Pendiente |
| 13  | Redirección sobrescribe la facultad del curso; no hay facultad/coordinador de origen para dirigir cierres                      | §2.6             | 🟡 Media   | ✅ Resuelto (`courses.origin_faculty`; cierres por facultad de origen) |
| 14  | Facultad del curso la elige el cliente en vez de heredarse del proveedor                                                       | §2.7             | 🟡 Media   | ✅ Resuelto (facultad heredada del proveedor) |
| 15  | Bandejas admin sin filtro por estado ni listado de "aprobados sin contrato"                                                    | §5.8             | 🟡 Media   | Pendiente |
| 16  | `GET /courses` sin filtros por dueño/estado ni visibilidad por rol                                                             | §2.8             | ⚪ Baja    | ✅ Resuelto (filtros `usuario_id`/`codigo_proveedor`/`estado` + visibilidad por rol) |
| 17  | Cohorte/período sin nombre (`nombre_cohorte`)                                                                                  | §2.9             | ⚪ Baja    | ✅ Resuelto (`nombre_cohorte`) |

**Sin brechas:** redirección/remit de solicitudes (§5.6), autenticación básica (§4.2), solapamiento curso+grupo (§6).
