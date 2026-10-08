# Análisis de brechas: Backend Go vs. mock Node de `diplomados/`

## 1. Introducción

`diplomados/` (`ecp_ucv`) es la app Next.js del portal de cursos/diplomados. Durante su desarrollo temprano incluyó un backend simulado propio: rutas de API de Next.js (`src/app/api/**`) que actúan como una capa fina de lógica de negocio delante de un almacén REST crudo servido con `json-server --watch db.json`. Ese mock se desarrolló en paralelo (y en algunos puntos, por delante) del backend real en Go, y por eso codifica un flujo de aprobación más completo del que existe hoy en `backend/internal/`.

Este documento compara ambos backends y documenta **en una sola dirección**: qué existe en el mock de Node que falta (total o parcialmente) en el backend Go. No se documentan capacidades que el backend Go tiene y el mock no — eso queda fuera del alcance.

**Alcance:** cursos, diplomados, usuarios, proveedores (solicitud de código de proveedor) y los endpoints de administración relacionados con solicitudes de curso ("course requests"). Los endpoints exclusivos de grupos se excluyen; ver §8 para la sección de solapamiento curso+grupo.

**Método:** lectura directa del código de rutas/handlers en `diplomados/src/app/api/**`, de los servicios del cliente (`diplomados/src/servicios/**`) y de los componentes que deciden por el contenido de las respuestas, más los esquemas de `db.json`, comparado contra `backend/internal/**` + el enrutamiento de `backend/cmd/deu/main.go`.

**Re-evaluaciones:**
- 2026-09-28, sobre el backend en `c3788d0` (incluye `0fc656c` — `estado_gestion`, `GET /courses/public`, `GET /course-requests` — y `c3788d0` — transiciones `abierto`/`cerrado`). Resultado: §2.2, §2.3, §5.4 y §5.5 pasan a resueltos; §5.1 se reclasifica; se agregan §2.6–§2.9 y §5.8.
- **2026-10-06**, sobre el backend en `15483aa` y `diplomados` en `main@312447a` (la comparación anterior fue contra `28b644f`). Los commits nuevos de `diplomados` tocan sobre todo proveedores (ISLR opcional, `tipo_persona='juridica'`, aprobación vía `/api/provider-requests`, listado con `status=all`), el historial completo de la bandeja admin (`status=all` / `estado=all`) y la descarga de la plantilla de participantes. Resultado: se agregan §2.10–§2.16, §4.3, §5.9, §5.10, la sección de proveedores (§6) y la de diferencias transversales (§7); §2.3, §2.4, §2.8, §4.1 y §5.8 reciben estrategia; §5.6 deja de considerarse sin brecha.
- **2026-10-06 (segunda pasada)**, contra la rama `feature/backend-go` de `diplomados` (`a692ac1`), rebaseada sobre `main@312447a`. La pasada anterior se hizo contra `main`, que todavía habla con el mock. Esta rama, en cambio, elimina las rutas de Node y reemplaza `src/app/api` por un proxy hacia Go (`src/app/api/[...path]/route.ts`). La traducción de respuestas está en `src/servicios/adapters.js` (`courseToLegacy`, `providerToLegacy`, `periodToCohort`, `requestStatusToLegacy`), y la del JWT en `src/utils/session.ts`. Toda brecha que el cliente ya traduce se marca como resuelta, con el mismo criterio que se aplicó en §2.10: el cliente se adapta al contrato de Go. La pasada también encontró defectos de la rama, la mayoría efecto del rebase (§10).

**Criterio para las estrategias:** la lógica de negocio de Node manda. Go debe comportarse lo más parecido posible: reglas, códigos de estado, rutas, nombres de campos y forma de las respuestas. Las excepciones se justifican caso por caso y solo por costo alto (p. ej. cambiar la autenticación, quitar claims del JWT o romper `grupos/`) o por seguridad. Cuando un nombre de Node choca con uno que Go ya expone, se prefiere **renombrar** el campo en go Go si es posible, antes de agregar. La rama `feature/backend-go` de `diplomados`, que adapta el cliente a Go tiene este trabajo en progreso.

**Leyenda de prioridad:**
- 🔴 **Alta** — bloquea o distorsiona el flujo de aprobación/negocio central.
- 🟡 **Media** — funcionalidad real que falta, pero con workaround o impacto acotado.
- ⚪ **Baja** — diferencia de forma/nomenclatura o mejora menor, sin bloquear el flujo.

---

## 2. Cursos

### 2.1 Apertura de cohorte/período con reglas de negocio (🔴 Alta) — ✅ Resuelto

- **Node:** `POST /api/courses/[id]/open` (`diplomados/src/app/api/courses/[id]/open/route.ts`). Body: `{cohortName, startDate, endDate, capacity}`. Antes de crear la cohorte valida `canOpen`: el curso debe estar en `estado='cerrado'`, **o** estar `aprobado`/`aprobada` **y** tener un contrato legal asociado (`contrato_id`/`documento_legal_id`); si no, responde 400. Si procede: marca el curso `estado='abierto'` y crea un registro en `course-cycles` (`course_id, nombre_cohorte, fecha_inicio, fecha_fin, capacidad, estado:'activa', creado_en`).
- **Go (histórico):** `POST /courses/:course_id/periods` (`backend/internal/course_periods/endpoints.go:80`). Era un alta simple (`fecha_inicio, fecha_fin, fecha_inscripcion`), sin guarda de estado previo del curso, sin campo de capacidad (`entities.CoursePeriod` no tenía `Capacity`), y sin relación con el estado de un contrato legal (que tampoco existe en Go — ver §5.2).
- **Resuelto:** el mismo endpoint (`POST /courses/:course_id/periods`) ahora exige `capacidad` (> 0) y agrega una guarda de negocio: solo se permite crear un período si el curso existe/está aprobado (`courses.ErrCourseNotFound` si no) **y** no tiene ya un período activo sin cerrar (`ErrCoursePeriodAlreadyOpen` en caso contrario). Esto cubre ambas ramas de la regla `canOpen` de Node sin necesitar el concepto de contrato: un curso recién aprobado no tiene períodos activos (pasa), y un curso previamente cerrado tampoco (pasa); uno con un período ya abierto es rechazado con 409. **No** depende de §5.2 (contrato) — ver §11.
- **Revisión 2026-10-06:** la resolución dejó fuera la mitad de `canOpen` que exige contrato, porque entonces Go no tenía ese concepto. §5.2 ya lo agregó (`has_documentation`), así que la regla de Node se puede portar completa. Ver §2.10.

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
- **Revisión 2026-10-06:** con el criterio de §1 (Node manda) se iguala el filtro. Hay una segunda diferencia: Node muestra los cursos amparados que nunca abrieron cohorte, porque al ampararlos los pone en `cerrado` (§2.16). Go los oculta, porque exige que exista un período.
- **Estrategia:**
  1. Hacer primero §2.16.
  2. En `queries.GetPublicCourses`, reemplazar el `EXISTS` sobre períodos por `estado_gestion IN ('abierto','cerrado')`, junto con `is_active` y `has_documentation`. Es el filtro de Node.
  3. Node acepta `limit` (15 por defecto) y devuelve `id, titulo, descripcion, image, estado_gestion, documento_legal_id`. Aceptar `limit` como alias de `per_page` (§7.2) y agregar los alias `titulo`, `image` y `documento_legal_id` (§2.13).
  4. Se mantiene el sobre `{cursos, paginas}` en vez del arreglo plano de Node: es el mismo de `GET /courses` y `cursos-service.js` ya lo lee (`response.cursos`).
- **Re-evaluación contra `feature/backend-go`:** la forma de la respuesta ya está resuelta en el cliente. `getPublicCourses` llama a `courses/public` con `per_page` y traduce cada curso con `courseToLegacy`. Sigue abierta la parte del filtro, que es regla de Go: el catálogo muestra cursos en `solicitud-cierre` y oculta los amparados que nunca abrieron cohorte (§2.16).
- **Resuelto (2026-10-07):** se aplicó el punto 2, después de §2.16. `queries.GetPublicCourses` filtra por `is_active`, `has_documentation` y `estado_gestion IN ('abierto','cerrado')`, sin el `EXISTS` sobre períodos. El catálogo oculta los cursos en `solicitud-cierre` y muestra los amparados que nunca abrieron cohorte, como Node. Los puntos 3 y 4 no se tocan: el cliente ya los resuelve.

### 2.4 Anuncios/publicaciones ligados a curso+cohorte, no solo a período (⚪ Baja) — ✅ Resuelto (en el cliente)

- **Node:** `publications` (`GET/POST /api/publications`) se indexan por `course_id` **y** `cohort_id` de forma independiente, y en el detalle de curso (`GET /api/courses/[id]`) solo se muestran las del cohorte más reciente.
- **Go:** los anuncios (`Announcement`) cuelgan exclusivamente de `course_periods/:period_id/announcements` (`backend/internal/course_periods/endpoints.go:149-212`) — modelo equivalente en la práctica (un período ≈ un cohorte), solo difiere la forma de la clave. No requiere cambio funcional, se documenta como diferencia de modelado menor.
- **Estrategia:** la UI consume las publicaciones dentro del detalle del curso (`cohortes[].publicaciones`), así que se cierra con §2.14 sin cambiar el modelo de Go. En los anuncios, agregar los alias `fecha` (= `creado_el`), `cohort_id` (= período) y `course_id`. `GET /publications?course_id=&cohort_id=` solo hace falta como alias de lectura si la UI lo sigue llamando directamente.
- **Resuelto en `feature/backend-go`:** `periodToCohort` arma `publicaciones[]` con `course_id`, `cohort_id` y `fecha` (= `creado_el`) a partir de los anuncios de cada período. `getPublicationsByCourse` los lee desde `cohortes`. No hace falta cambiar Go.

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
- **Diferencia residual (⚪, 2026-10-06):** Node pasa `estado` tal cual al almacén, así que también filtra por `under_review`, `aprobada` o `rechazada`, porque en Node el curso tiene un solo campo `estado`. Go responde 400 con cualquier valor distinto de `abierto`/`cerrado`/`solicitud-cierre`. Además, la visibilidad pública depende de que los cursos amparados pasen a `cerrado` (§2.16).
- **Estrategia:** aceptar en `estado` el vocabulario unificado de §2.13 (`under_review`, `aprobada` y `rechazada` filtran por el estado de la solicitud). Aceptar también `_page`/`_limit` y `limit` (§7.2), que es lo que envía `cursos-service.js`.
- **Re-evaluación contra `feature/backend-go`:** el cliente ya manda `page`/`per_page`, y ningún componente pasa `estado` a `getAllCourses`. Los cursos en revisión del propio proveedor salen de `GET /course-requests` (`_pendingCoursesOf`). El residual queda sin impacto actual; solo la visibilidad de los cursos amparados sin cohorte depende de §2.16.

### 2.9 Cohorte sin nombre (⚪ Baja) — ✅ Resuelto

- **Node:** cada `course-cycles` tiene `nombre_cohorte` (viene de `cohortName` en `POST /api/courses/[id]/open`), y la solicitud de cierre lo repite en `payload.nombre_cohorte`.
- **Go:** `entities.CoursePeriod` (`backend/internal/entities/course_period.go`) no tiene campo de nombre; un período se identifica solo por ID y fechas.
- **Impacto:** cosmético/UX: la UI de `diplomados/` muestra el nombre de la cohorte en el detalle del curso y en la bandeja de cierres.
- **Resuelto:** nueva columna `course_cycles.name` (migración `000033`; los períodos existentes quedan con nombre vacío). `POST` y `PUT /courses/:course_id/periods` exigen `nombre_cohorte`, como el formulario de Node (`gestion-cohorte-form.tsx`). Un nombre vacío o de solo espacios se rechaza con 400 (`ErrCohortNameRequired`, validado en el servicio igual que `capacidad`). Se expone como `nombre_cohorte` en los períodos y en `ultimo_periodo` del detalle del curso. La solicitud de cierre no repite el nombre (Node lo copia en `payload.nombre_cohorte`): en Go la solicitud ya referencia el período por `course_cycle_id`, y de ahí se obtiene el nombre.

### 2.10 Apertura de cohorte sin exigir amparo legal (🔴 Alta) — ✅ Resuelto

- **Node:** `POST /api/courses/[id]/open` (`diplomados/src/app/api/courses/[id]/open/route.ts`) calcula `canOpen = estado === 'cerrado' || (estado ∈ {aprobado, aprobada} && (contrato_id || documento_legal_id))`. Si es falso, responde 400 `El curso debe estar amparado legalmente (aprobado con contrato o cerrado) para poder abrir una cohorte.`, también cuando el curso ya está `abierto` o en `solicitud-cierre`. El body es `{cohortName, startDate, endDate, capacity}` (`gestion-cohorte-form.tsx`) y responde 200 `{success: true, message: 'Cohorte abierta exitosamente'}`.
- **Go:** `course_periods.service.CreateCoursePeriod` valida capacidad, nombre, curso aprobado, que no esté en `solicitud-cierre` y que no haya un período activo, pero no revisa `has_documentation`. La ruta es `POST /courses/:course_id/periods` con `{nombre_cohorte, fecha_inicio, fecha_fin, fecha_inscripcion, capacidad}` y responde 201 `{id}`. Los rechazos son 409.
- **Impacto:** un curso aprobado sin contrato puede abrir cohorte y empezar a inscribir sin amparo legal. Es la regla de negocio que Node pone delante de la apertura.
- **Estrategia:**
  1. En `CreateCoursePeriod`, rechazar con un error nuevo (`ErrCourseNotCovered`) si `!course.HasDocumentation`. Basta con ese campo: un curso `cerrado` ya estuvo abierto y, por lo tanto, amparado.
  2. Registrar `POST /courses/:id/open` (protegida) como alias que lee `cohortName`, `startDate`, `endDate` y `capacity`, llama al mismo servicio y responde 200 `{success, message, id}`. En el alias, los tres rechazos (sin amparo, ya abierto, cierre pendiente) responden 400 con el mensaje de Node, porque Node los trata como un solo caso. La ruta actual conserva sus 409.
- **Resuelto (2026-10-06):** se aplicó solo el punto 1.
  - `CreateCoursePeriod` rechaza con `ErrCourseNotCovered` los cursos sin `has_documentation`, y `POST /courses/:course_id/periods` responde 409, como en las otras guardas de estado.
  - El alias del punto 2 se descartó: dos rutas con contratos distintos para lo mismo cuestan más que adaptar `courseService.openCohort` en la rama `feature/backend-go` de `diplomados`, que ya traduce llamadas al contrato de Go.

### 2.11 Formulación de curso: contrato de entrada incompatible (🔴 Alta) — ✅ Resuelto (en el cliente)

- **Node:** `POST /api/courses` (`diplomados/src/app/api/courses/route.ts`) recibe multipart con `payload` (JSON en texto), los archivos `cover` (o `archivo_proyecto`/`archivo`) y `cv_facilitador`, y `tipo`. Si no hay `payload`, toma los campos sueltos del form. `curso-form.tsx` manda en `payload`: `titulo`, `denominacion`, `proposito`, `fundamentacion`, `duracion`, `estructura_costos`, `perfil_docente`, `perfiles`, `exigencias`, `estructura_curricular`, `evaluacion`, `cronograma`, `contenido_competencias` y `bibliografia`. El servidor arma `nombre = titulo || denominacion || nombre` y `descripcion = fundamentacion || descripcion`. Solo exige la portada (400 `Falta imagen de portada`) y responde 201 con el curso creado.
- **Go:** `toCourseEntity` (`backend/internal/courses/decoders.go:32`) solo lee campos sueltos. Exige `nombre`, `descripcion`, `objetivos`, `fundamentacion`, etc., y el archivo `portada`. Responde 201 `{id}`.
- **Impacto:** el formulario actual de diplomados falla siempre contra Go: no llega `nombre`, `objetivos` ni `portada`.
- **Estrategia:** cambiar solo el decoder:
  1. si viene `payload`, parsearlo y usarlo como fuente principal; los campos sueltos quedan de respaldo;
  2. resolver alias como Node: `nombre ← titulo | denominacion | nombre`, `descripcion ← fundamentacion | descripcion`, `objetivos ← objetivos | proposito`;
  3. portada ← `portada | cover | archivo_proyecto | archivo`;
  4. ignorar `userId` (Go usa el token) y `tipo` (es otro concepto, ver §2.13);
  5. mantener las validaciones de obligatorios sobre los valores ya resueltos. El formulario de Node exige todos esos campos, así que en la práctica no se rechaza nada que Node acepte;
  6. responder 201 con el curso creado (`GetCourseResponse`, con los alias de §2.13) en lugar de solo `{id}`. `id` sigue presente.

  No se porta la variante JSON de Node: sin archivo solo funciona si el cliente manda una URL de imagen ya alojada, y ningún flujo de la UI lo hace.
- **Resuelto en `feature/backend-go`:** `solicitudesService.createSolicitud` pasa el formulario por `_courseFormData` (`src/servicios/solicitudes-service.js`) antes de llamar a `POST /courses`. Esa función lee el `payload` y resuelve los alias igual que Node: `nombre ← titulo | denominacion`, `descripcion` y `fundamentacion ← fundamentacion`, `objetivos ← proposito`. La portada sale de `cover`, `archivo_proyecto` o `archivo` y se envía como `portada`; el CV va como `cv_facilitador`. `curso-form.tsx` no usa la respuesta, así que la diferencia entre el curso creado y `{id}` no afecta. Se descarta la estrategia de cambiar el decoder de Go.

### 2.12 Formulación por un usuario que no es proveedor aprobado (🟡 Media) — ✅ Resuelto

- **Node:** antes de crear, busca al usuario y exige `rol === 'proveedor'`. Si no lo es, responde 403 `Acceso Denegado`. Ese rol solo se asigna al aprobar al proveedor (§6.5).
- **Go:** `courses.service.CreateCourse` usa `GetProviderByUserID`, que no filtra por `status` ni por `deleted_at` (`queries.GetProviderByUserID`). Un proveedor `under_review` o `rejected` puede formular cursos, y un usuario sin proveedor recibe 500.
- **Estrategia:** en el servicio, exigir `provider.Status == active`. Si no hay proveedor o no está activo, devolver un error nuevo (`ErrNotApprovedProvider`) que el handler mapea a 403 `Acceso Denegado`. Agregar `deleted_at IS NULL` a `GetProviderByUserID`; con §6.2 deja de haber más de un proveedor vigente por usuario. Un `RequireRoles("course_admin")` en la ruta haría casi lo mismo, pero el chequeo en el servicio también cubre a los proveedores desactivados.
- **Re-evaluación contra `feature/backend-go`:** sigue abierta y se puede reproducir desde la UI. Go genera el `codigo_proveedor` al crear el proveedor, y `GetProviderCodeByUserID` (login) no filtra por estado, así que un proveedor `under_review` recibe `v1.providerCode` en su token. `navbar.tsx` muestra "Formular" a quien tenga `codigo_proveedor`, y Go acepta el curso. La estrategia no cambia; conviene además que el login solo emita `providerCode` para proveedores `active`.
- **Resuelto (2026-10-06):**
  - `courses.service.CreateCourse` responde `ErrNotApprovedProvider` si el usuario no tiene proveedor o si su proveedor no está `active`, y `POST /courses` lo mapea a 403.
  - El código de proveedor solo se entrega a proveedores `active`, tanto en el login (`queries.GetProviderCodeByUserID`) como en el perfil (`GET /users/:id` y `GET /users`, con un `LEFT JOIN` que filtra por `status='active'` y `deleted_at IS NULL`). Así, la navbar ya no le muestra "Formular" a un proveedor pendiente.
  - `v1.providerID` se mantiene para cualquier estado, porque el cliente lo usa para que el proveedor pendiente vea su propia solicitud.

### 2.13 Forma de la respuesta del curso (🟡 Media) — ✅ Resuelto

| Node | Go | Dónde lo usa la UI |
| --- | --- | --- |
| `image_url` (la UI lee `imagen \|\| image_url \|\| cover`) | `portada` | portada en catálogo y detalle: con Go no se ve |
| `proposito` | `objetivos` | detalle del curso |
| `tipo_curso`: `formulacion-curso-directa` / `-indirecta` | `tipo`: `skill_development`, `life_skills`… (otro concepto) | `admin-actions.tsx`: si se puede remitir y si la clasificación es fija |
| `contrato_id` / `documento_legal_id`: texto, `INTENCION-…` o `ADENDA-…` | `tiene_documentacion_legal` (bool) | `solicitudes-table.tsx` (sello legal, filas cerradas), `profile-coordinator-review.tsx` (`startsWith('ADENDA')`) |
| `estado`: un solo campo con `under_review`, `aprobada`, `rechazada`, `abierto`, `cerrado`, `solicitud-cierre` | `activo` + `estado_gestion`; el estado de la solicitud va aparte | badges, filtros y `cursos-service.js` (`estado_gestion: c.estado \|\| c.estado_gestion`) |

- **Impacto:** con Go, la UI no muestra portadas, no detecta los cursos amparados (todos parecen "sin contrato") y no distingue las solicitudes directas de las indirectas.
- **Estrategia:** agregar campos sin quitar los de Go:
  - `image_url` = la misma URL que `portada`;
  - `proposito` = `objetivos`;
  - `tipo_curso` calculado: `formulacion-curso-indirecta` si `faculty != origin_faculty`, si no `formulacion-curso-directa`. Es exactamente lo que Node marca al remitir y no hace falta una columna;
  - `documento_legal_id` y `contrato_id`: el ID del contrato que ampara el curso (`provider_contract_courses`), con el prefijo de Node según el tipo: `INTENCION-<id>` para `inicial` y `ADENDA-<id>` para `adenda`. Vacíos si no hay amparo. Requiere un `LEFT JOIN` en las consultas de curso;
  - `estado`: `estado_gestion` si tiene valor; si no, el estado de la solicitud traducido según §7.1 (`under_review`, `aprobada`, `rechazada`).

  `link_certificados` no se porta: la UI lo muestra si existe, pero ninguna ruta de Node lo escribe.
- **Re-evaluación contra `feature/backend-go`:** `courseToLegacy` ya traduce `image`, `proposito`, `tipo` (directa o indirecta, con `formulationType`), `estado_gestion` y el contrato. Quedan dos huecos:
  - **`image_url`:** `course-details-view.tsx`, la vista admin de la solicitud, lee `payload.image_url || imagen || cover`. El adaptador solo arma `image`, así que la portada no se ve en la revisión. El cambio en `main` (`9a052dd`) que muestra la portada llegó con el rebase. Estrategia: agregar `image_url: c.portada` en `courseToLegacy`.
  - **ID del contrato por curso:** el adaptador pone `contrato_id = documento_legal_id = 'CONTRATO-VIGENTE'`. `profile-coordinator-review.tsx` (`getDownloadLinks`) busca los PDF según el prefijo `INTENCION`/`ADENDA` y `legal_status.adendas[].id_adenda`, así que el coordinador nunca ve los enlaces de descarga del contrato que ampara cada curso. Go ya devuelve en el detalle del proveedor `contratos_legales[].cursos_amparados`, y con eso el cliente puede resolver qué contrato ampara cada curso (`INTENCION-<id>` o `ADENDA-<id>`) sin cambiar Go.
- **Resuelto (2026-10-07):** se renombraron dos campos de Go, sin agregar campos duplicados:
  - `portada` pasa a ser `image_url` en las respuestas de curso (`GetCourseResponse`) y de solicitud de curso (`CourseInfo`). El campo del formulario de alta sigue siendo `portada`. `courseToLegacy` arma `image` desde `image_url`, y `course-details-view.tsx` lee `image_url` sin cambios.
  - `tiene_documentacion_legal` (bool) pasa a ser `documento_legal_id`, un texto con el contrato que ampara el curso: `INTENCION-<id>` para el `inicial` y `ADENDA-<id>` para una adenda (`entities.ProviderContract.DocumentID`). Queda vacío si no hay amparo, así que el chequeo `!!documento_legal_id` de la UI sigue funcionando. Sale de un `LEFT JOIN LATERAL` sobre `provider_contract_courses` (`queries.withCoveringContract`). El formato `ADENDA-<id>` coincide con el `id_adenda` que arma `providerToLegacy`, y con eso `getDownloadLinks` encuentra los PDF. `courseToLegacy` copia el valor a `contrato_id` y `documento_legal_id`, y deja de usar `'CONTRATO-VIGENTE'`. La columna `has_documentation` sigue siendo la guarda de apertura (§2.10).

### 2.14 Detalle del curso sin cohortes ni publicaciones (🟡 Media) — Parcial

- **Node:** `GET /api/courses/[id]` incluye `cohortes[]`, de la más reciente a la más vieja. Cada cohorte trae `id, course_id, nombre_cohorte, fecha_inicio, fecha_fin, capacidad, estado ('activa' | 'cerrada'), creado_en` y `publicaciones[]` (`id, course_id, cohort_id, titulo, contenido, fecha`), también de la más reciente a la más vieja. Además incluye `cohorteActiva`, que es la cohorte más reciente aunque esté cerrada. `course-owner-view.tsx` y `curso.tsx` publican sobre esa cohorte.
- **Go:** `GetCourseResponse` solo trae `ultimo_periodo` (`id, nombre_cohorte, fecha_inicio, fecha_fin, fecha_inscripcion`), sin `capacidad` ni estado. Los períodos con sus anuncios salen de `GET /courses/:course_id/periods`, con `creado_el`/`actualizado_el` (no `creado_en`) y sin `estado`.
- **Estrategia:** en `courses.service.GetCourse`, cargar los períodos del curso con sus anuncios y exponer `cohortes` y `cohorteActiva` con los nombres de Node. El `estado` sale de `is_active`. Se mantiene `ultimo_periodo`. En períodos y anuncios, agregar `creado_en` y `fecha` como alias. Esto también cierra §2.4.
- **Re-evaluación contra `feature/backend-go`:** `getCourseById` arma `cohortes` y `cohorteActiva` desde `GET /courses/:id/periods` con `periodToCohort`. Falta el historial: `queries.GetCoursePeriods` filtra `cp.is_active = true`, así que el detalle no muestra las cohortes cerradas ni sus publicaciones, y `cohorteActiva` queda en `null` después de un cierre. Node las muestra todas, con `estado: 'cerrada'`. Estrategia: que `GET /courses/:course_id/periods` incluya los períodos cerrados (o acepte un `?estado=all`) y exponga `estado` (`activa`/`cerrada`, §7.1); en el cliente, `periodToCohort` toma ese `estado` en vez de fijar `'activa'`.

### 2.15 Solicitud de cierre identificada por curso (🟡 Media) — ✅ Resuelto (en el cliente)

- **Node:** `POST /api/courses/[id]/closures` lleva el ID del **curso** en la ruta, y en el form `userId`, `titulo_curso`, `nombre_cohorte`, `observaciones` y los 3 archivos. Responde 404 `El curso especificado no existe.` o 201 `{success: true, message: 'Solicitud de cierre enviada exitosamente.'}`.
- **Go:** `POST /course-cycle-close-requests` exige `course_cycle_id` (`decoders.go` → `toCloseRequestEntity`) y responde 201 `{id}`.
- **Impacto:** el cliente (`cursos-service.js` → `requestCohortClosure(courseId, files)`) solo manda el curso. Contra Go recibe 400 `course_cycle_id is required`.
- **Estrategia:** agregar el alias `POST /courses/:id/closures` (protegido). Resuelve el período activo con `GetActiveCoursePeriodByCourseID` y delega en `SubmitCloseRequest`. `titulo_curso` y `nombre_cohorte` se ignoran: Go los toma del curso y del período. Responde 201 `{success, message, id}`. Excepción: si el curso no tiene una cohorte abierta, Go responde 409. Node igual crearía la solicitud, pero en Go el cierre necesita un período al que apuntar.
- **Resuelto en `feature/backend-go`:** `courseService.requestCohortClosure` busca la `cohorteActiva` del curso, agrega `course_cycle_id` y llama a `POST /course-cycle-close-requests`. Si no hay cohorte activa, muestra "El curso no tiene una cohorte activa para cerrar.". Se descarta el alias en Go.

### 2.16 Un curso amparado no pasa a `cerrado` (🟡 Media) — ✅ Resuelto

- **Node:** `POST /api/legal-status/[user_id]` pone `estado = estado_gestion = 'cerrado'` en cada curso que ampara. En Node, `cerrado` significa "amparado y sin cohorte abierta":
  - `canOpen` lo acepta (§2.10);
  - `GET /api/courses/public` y `GET /api/courses` lo muestran;
  - `user-profile.tsx` lo rotula "Amparado / Vigente".
- **Go:** §5.2 decidió no portar ese cambio de estado. Amparar solo marca `has_documentation`, y `estado_gestion` queda vacío hasta que se abre la primera cohorte.
- **Impacto:** en Go, un curso amparado que nunca abrió cohorte no aparece en `GET /courses/public` (exige un período) ni, para un anónimo, en `GET /courses` (exige `abierto`/`cerrado`). En Node sí aparece en ambos.
- **Estrategia:** revisar la decisión de §5.2 y portar la regla de Node acotada. Al amparar un curso (`CreateProviderContract`), poner `estado_gestion='cerrado'` en la misma transacción, **solo si está vacío**. Así no se pisa un curso `abierto` o en `solicitud-cierre`; el mock sí lo pisa, pero eso se considera un defecto y no se porta. Agregar una migración de datos que ponga `cerrado` en los cursos con `has_documentation=true` y `estado_gestion` vacío. Después de esto, §2.3 y §2.8 filtran solo por `estado_gestion`, igual que Node.
- **Re-evaluación contra `feature/backend-go`:** sigue abierta del lado de Go. En el cliente, `courseToLegacy` lleva un `estado_gestion` vacío a `aprobada`, y `gestion-cohorte-form.tsx` deja abrir cohorte con `aprobada` + contrato, así que la apertura funciona. Lo que el cliente no puede arreglar es el catálogo público, que se filtra en el servidor.
- **Resuelto (2026-10-07):**
  - `queries.SetCoursesDocumented`, el `UPDATE` que `CreateProviderContract` ya ejecuta dentro de su transacción, ahora también hace `estado_gestion = COALESCE(estado_gestion, 'cerrado')`. Un curso `abierto` o en `solicitud-cierre` no se pisa.
  - La migración `000040` pone `cerrado` en los cursos con `has_documentation=true` y `estado_gestion` vacío. El `down` lo revierte solo en los cursos que nunca tuvieron un período.
  - `CreateCoursePeriod` ya acepta `cerrado`, así que la apertura no cambia. Con esto `GET /courses` (anónimo) muestra los cursos amparados sin cohorte, y §2.3 pudo pasar a filtrar por `estado_gestion`.

---

## 3. Diplomados

**No existe un concepto de "diplomado" separado de "curso" en ninguno de los dos backends.**

- **Go:** `entities.RequestType_COURSE = "diplomado"` (`backend/internal/entities/entities.go:23`) — un `Course` **es** el diplomado; no hay módulo, tabla ni entidad `Diplomado` distinta.
- **Node:** no existe colección `diplomado` en `db.json`, ni ruta, ni servicio; todo se modela como `curso`/`course-requests`, con `tipo_curso ∈ {formulacion-curso-directa, formulacion-curso-indirecta}`.

Por lo tanto esta sección no aporta brechas nuevas de endpoints — las brechas reales de "diplomado" son las de cursos (§2) y solicitudes de curso (§5). Sí hay una diferencia de **forma de datos** a nivel de entidad/response que vale la pena señalar:

### 3.1 Campos de clasificación/evaluación ausentes en la entidad Course de Go (⚪ Baja — solapa con §5.3)

- **Node:** el registro de `course-requests`/`courses` incluye `clasificacion`, `calificacion`, `motivo_rechazo`, `archivo_evaluacion_url`, `contrato_id`/`documento_legal_id` directamente sobre el curso.
- **Go:** `Course`/`GetCourseResponse` (`backend/internal/courses/response.go`) no expone ninguno de estos campos; `motivo_rechazo`-equivalente (`Comments`) vive solo en `CourseRequest`, no en `Course`, y `clasificacion`/`calificacion`/contrato no existen en absoluto en el modelo (ver §5.3 y §5.2 para el detalle funcional).
- **Actualización (2026-10-01):** resuelto. `clasificacion` y `calificacion` viven en la solicitud (§5.3), y la propuesta completa (`contenido_competencias`, `bibliografia`, `cv_facilitador_url`) en el curso (migración `000034`).
- **Actualización (2026-09-28):** la parte de contrato y estado ya está cubierta. `GetCourseResponse` expone `tiene_documentacion_legal` (§5.2) y `estado_gestion` (`abierto`/`cerrado`/`solicitud-cierre`, §2.2/§5.4). Faltan solo `clasificacion`, `calificacion` y `archivo_evaluacion_url`, que dependen de §5.3.

---

## 4. Usuarios

Esta es la sección con **menor brecha relativa** — el módulo `auth` de Node también es mínimo (solo login), así que no hay que sobredimensionar las diferencias aquí.

### 4.1 Endpoints presentes en ambos lados, con diferencias de forma (⚪ Baja) — filtro por rol ✅ resuelto en el cliente

|          | Node                                                                    | Go                                                                                     |
| -------- | ----------------------------------------------------------------------- | -------------------------------------------------------------------------------------- |
| Registro | `POST /api/users` (multipart) — `diplomados/src/app/api/users/route.ts` | `POST /users` (multipart) — `backend/internal/users/endpoints.go:73`                   |
| Listado  | `GET /api/users?rol=`                                                   | `GET /users` (paginado, sin filtro por rol) — `backend/internal/users/endpoints.go:50` |
| Detalle  | `GET /api/users/[id]`                                                   | `GET /users/:id` — `backend/internal/users/endpoints.go:36`                            |
| Login    | `POST /api/auth/login`                                                  | `POST /auth/login` — `backend/internal/auth/endpoints.go:26`                           |

- **Nomenclatura de campos inconsistente dentro del propio Go:** `POST /users` usa `cedula, nombres` mientras `PUT /users/:id` usa `ci, primer_nombre` para conceptos equivalentes (`backend/internal/users/request.go` — comparar create vs. update). Es una inconsistencia interna, no causada por Node, pero se señala porque complica igualar el contrato con el cliente de `diplomados/` si algún día se apunta directo al backend Go.
  - **Estrategia:** Node no tiene edición de usuarios, así que no hay un contrato de Node que copiar. Hacer que `PUT /users/:id` acepte los nombres del alta (`cedula`, `nombres`, `apellidos`, `fecha_de_nacimiento`, `nivel_educativo`) y conservar los actuales (`ci`, `primer_nombre`, …) como alias. Ninguno de los dos frontends usa los nombres actuales (verificado con búsqueda en `diplomados/src` y `grupos/src`).
- **Filtro por rol en listado de usuarios:** Node soporta `GET /api/users?rol=` (usado para poblar selects de "coordinadores", por ejemplo `getCoordinadores()` en `diplomados/src/servicios/users-service.js`). Go's `GET /users` no acepta un filtro por rol — impacto bajo, se puede filtrar client-side, pero es un endpoint usado activamente en el mock para poblar UI de administración.
  - **Revisión 2026-10-06 (sube a 🟡):** `solicitud-form.tsx` (registro de proveedor) y `admin-actions.tsx` (remitir) arman la lista de facultades con `getCoordinadores()`. Esperan un **arreglo** con `id` y `facultad` por coordinador, y el `id` del coordinador viaja después como `coordinador_id` (§6.1). Go devuelve `{usuarios, paginas}`, no filtra y no expone `facultad`, así que la lista queda vacía y no se puede registrar un proveedor ni remitir.
  - **Estrategia:**
    1. Aceptar `rol` en `GET /users` con los nombres de la UI y traducirlos (`coordinador → faculty_admin`, `admin → root, deu_admin`, `proveedor → course_admin`, `visitante`), inverso a `entities.UIRole`.
    2. Con `rol`, responder un arreglo plano sin paginar, como Node, con un perfil reducido: `id`, `nombres`, `apellidos`, `facultad` y `rol`. La ruta solo exige JWT, y un visitante que se registra como proveedor la necesita: no hay que exponerle el perfil completo de otros usuarios.
    3. Sacar `facultad` de `user_roles.faculty`, la misma fuente de la claim `v1.faculty`. Agregar `facultad` también al perfil completo de `GET /users/:id`.

    El coordinador de la DEU se llama "Dirección de Extensión de Estudio" en `db.json` y `DEU` en el enum de Go. La UI muestra lo que reciba, así que alcanza con devolver el valor del enum.
  - **Resuelto en `feature/backend-go`:** `getCoordinadores()` ya no consulta usuarios. Devuelve la lista fija de facultades del enum de Go (`src/data/facultades.ts`), con `id` = valor del enum y `facultad` = etiqueta. Go dirige las solicitudes por facultad, no por coordinador, así que `solicitud-form.tsx` y `admin-actions.tsx` envían directamente la facultad. Se descarta el filtro `?rol=` en Go.
- **Rol por defecto al registrar:** Node asigna `rol: 'visitante'` por defecto si no se especifica (`diplomados/src/app/api/users/route.ts`). Confirmar que el alta en Go asigna un rol por defecto equivalente y no deja el campo vacío.
  - **Verificado (2026-10-06):** Go asigna `visitante` al crear el usuario (`backend/internal/users/decoders.go:52`). Sin brecha.

### 4.2 Sin brecha de autenticación real

Ninguno de los dos lados tiene registro completo con verificación de email, logout, refresh token o recuperación de contraseña — Node lo simula con un JWT sin firmar y Go tiene DTOs de registro (`RegisterRequest`/`RegisterResponse` en `backend/internal/auth/request.go:8-11` y `response.go:14-16`) que **no están conectados a ningún handler** (código muerto). No se lista como brecha porque Node tampoco lo implementa — es una carencia compartida, fuera del alcance de este documento comparativo.

### 4.3 Login y claims del token (⚪ Baja) — ✅ Resuelto (en el cliente)

- **Node:** `POST /api/auth/login` recibe `{usuario, password}`. El token lleva en el nivel superior `id`, `sub`, `name`, `email`, `nombre`, `nombres`, `apellidos`, `rol`, `roles`, `facultad` y `codigo_proveedor`, más `v1{userID, roles}`. La respuesta es `{success, token, access_token, jwt, user, ...claims}`. Los errores son 404 `Usuario no encontrado` y 401 `Contraseña incorrecta`.
- **Go:** el token lleva `rol` y `roles` (nombres de la UI) y `v1{roles (nombres de Go), userID, faculty, providerCode, …}`. La respuesta es `{token, usuario{id, nombre, roles, rol}, facultad, codigoProveedor, …}`.
- **Uso en la UI:** `auth-context.tsx` lee `sub || v1.userID`, que funciona con Go. Lee `name`, que en Go no existe, y por eso muestra "Usuario" hasta que carga el perfil. Lee `v1.roles`, que en Go trae los nombres internos, aunque luego los pisa con `GET /users/:id`. `auth-service.js` lee `response.token`, que funciona con Go.
- **Estrategia:** agregar al token las claims de nivel superior `sub` (= userID), `name`, `email`, `facultad` y `codigo_proveedor`. Solo se agregan, no se quita ni se cambia ninguna. En la respuesta del login, agregar `user` con esos mismos datos. **No** se portan:
  - `v1.roles` con nombres de la UI: la autorización del backend lo lee (§5.7) y cambiarlo cuesta demasiado;
  - la distinción 404/401 del login: revelaría qué correos están registrados.
- **Resuelto en `feature/backend-go`:** `src/utils/session.ts` (`sessionFromToken`) arma la sesión desde `v1`: `id`, `rol` y `roles` en el vocabulario de la UI, `facultad`, `codigo_proveedor` y `provider_id`. `auth-context.tsx`, `users-service.js` y `cursos-service.js` usan esa sesión. No hace falta agregar claims en Go.

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
- **Revisión 2026-10-07:** §2.16 revierte la decisión de no tocar `estado_gestion`. Al amparar un curso se pone en `cerrado`, pero solo si `estado_gestion` está vacío.
- **Fuera de alcance de esta resolución:** §2.3 (filtro del listado público por curso "listo") sigue pendiente — ahora existe el campo real (`has_documentation`) sobre el cual construir ese filtro, pero conectarlo queda como trabajo de seguimiento separado y acotado.

### 5.3 Aprobación sin evaluación adjunta (🟡 Media) — ✅ Resuelto

- **Node:** `POST /api/course-requests/[id]/approve` (dentro de `[action]/route.ts`) acepta multipart con `calificacion`, `clasificacion`, y `archivo_evaluacion` (PDF), que quedan guardados en el propio registro de la solicitud.
- **Go:** `POST /admin/course-requests/:id/approve` (`backend/internal/course_requests/endpoints.go:44`) solo acepta `{tipo_curso, observaciones}` — no hay campo de calificación numérica, clasificación, ni adjunto de documento de evaluación.
- **Impacto:** se pierde la trazabilidad de *por qué* se aprobó un curso y con qué nota/documento de respaldo lo evaluó el comité.
- **Resuelto:** `POST /admin/course-requests/:id/approve` acepta multipart además de JSON, con `calificacion` (número ≥ 0), `clasificacion` y el archivo opcional `archivo_evaluacion`.
  - La calificación y la clasificación se guardan en `course_auth_requests.score`/`classification` (migración `000035`).
  - El PDF se sube a B2 como archivo privado (`owner_type='course_request'`). Su fila en `files` se inserta en la misma transacción que la aprobación; si la aprobación falla, se borra de B2.
  - `GET /admin/course-requests/:id` expone `calificacion`, `clasificacion` y `archivo_evaluacion_url` (prefirmada).

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

### 5.6 Redirección de solicitud — ya cubierto en ambos lados (parcial desde 2026-10-06, ver §5.9)

`POST /admin/course-requests/:id/redirect` (Go, `backend/internal/course_requests/endpoints.go:104`) y la acción `remit` de `POST /api/course-requests/[id]/[action]` (Node) son equivalentes: reasignan la solicitud a otra facultad/coordinador. Se documenta aquí solo para confirmar que **no** es una brecha — Go ya tiene esta funcionalidad.

**Revisión 2026-10-06:** la reasignación es equivalente, pero Node marca además la solicitud como indirecta y la UI pone dos límites sobre cuándo se puede remitir. Ver §5.9.

### 5.7 Nomenclatura de roles — ✅ Resuelto

La UI de `diplomados/` decide qué mostrar comparando el rol literal. Por ejemplo, `curso-form.tsx` solo permite formular cursos si `user.rol === 'proveedor'`. Go usa otros nombres internamente. Equivalencias (`entities.UIRole` / `entities.UIRoles` en `backend/internal/entities/user.go`):

| Go (`deu.roles`, autorización) | UI de diplomados |
| --- | --- |
| `root`, `deu_admin` | `admin` |
| `faculty_admin` | `coordinador` |
| `course_admin` (proveedor aprobado) | `proveedor` |
| `visitante` | `visitante` |
| `group_admin`, `group_helper` | sin traducción |

Dónde viaja cada vocabulario:
- **JWT:** `rol` y `roles` en el nivel superior usan los nombres de la UI. `v1.roles` mantiene los de Go, y la autorización del backend (`RequireRoles`, `facultyscope`) solo lee `v1`.
- **`GET /users/:id`** (perfil completo: el propio usuario y los admins): `rol` y `roles` con los nombres de la UI. El perfil público no los incluye.
- **`POST /auth/login`:** `usuario.roles` mantiene los nombres de Go, porque `grupos/` decide sus rutas con ellos (`faculty_admin`, `deu_admin`, `root`). `usuario.rol` lleva el rol principal en el vocabulario de la UI.

No se renombraron los roles en la base: `grupos/` depende de `faculty_admin`, y la base sigue siendo la fuente de verdad.

### 5.8 Bandejas de admin sin filtro por estado ni bandeja de "aprobados sin contrato" (🟡 Media) — Parcial

- **Node:**
  - `GET /api/admin/courses` (`diplomados/src/app/api/admin/courses/route.ts`) no es un listado plano. Arma una bandeja de trabajo con dos grupos: solicitudes `under_review` y cursos `aprobado` **sin contrato** (`!contrato_id && !documento_legal_id`). Excluye rechazados y redirigidos. El coordinador ve el primer grupo según la facultad actual y el segundo según la facultad de origen (ver §2.6).
  - `GET /api/admin/closures` filtra por `estado`, y por defecto muestra solo `under_review`.
- **Go:**
  - `GET /admin/course-requests` (`backend/internal/course_requests/endpoints.go:147`) y `GET /admin/course-cycle-close-requests` (`backend/internal/course_cycle_close_requests/endpoints.go:124`) aceptan solo `faculty` y paginación, y devuelven solicitudes en **todos** los estados.
  - No hay forma de listar "cursos aprobados que aún no tienen contrato", que es justamente la lista que el admin necesita para saber a qué proveedores enviarles `POST /admin/providers/:id/legal-contracts` (§5.2). Por dentro esa consulta ya existe (`provider_contracts_queries.go` filtra `has_documentation=false` para amparar cursos), pero no está expuesta como listado. `GET /admin/providers` tampoco tiene un filtro equivalente.
- **Impacto:** el frontend tendría que traer todo y filtrar del lado del cliente, y para la bandeja de contratos pendientes directamente no hay de dónde sacar la información.
- **Parcial (2026-10-01):** `diplomados` ya arma ambas bandejas filtrando del lado del cliente.
  - Los items de `GET /admin/course-requests` traen `curso{…, tiene_documentacion_legal}` y `usuario_id`, así que "aprobados sin contrato" sale de la misma lista.
  - El admin DEU sin `?faculty` ve todas las facultades (`facultyscope.ResolveOptional`).
  - Falta el `?status=` en el servidor. Arreglo sugerido: agregar `?status=` a ambos listados admin y un filtro del tipo `has_documentation=false` (sobre cursos aprobados o sobre proveedores).
- **Revisión 2026-10-06:** Node cambió la semántica de los dos listados.
  - `GET /api/admin/courses?status=`: sin `status`, o con cualquier valor distinto de `all`, devuelve la bandeja de siempre. `status=all` devuelve el historial: el admin ve todo; el coordinador ve lo de su facultad actual **o** de origen, en cualquier estado. La página `admin/solicitudes` ahora pide siempre `status=all`, pagina en memoria y atenúa las filas terminadas (rechazadas, o aprobadas con contrato).
  - `GET /api/admin/closures?estado=`: por defecto `under_review` (también acepta `pendiente`); `all` devuelve todos; cualquier otro valor filtra por igualdad. El admin ve todo y el coordinador solo los cierres dirigidos a él.
  - Go, desde `15483aa`, filtra `GET /admin/course-requests` por facultad actual **o** de origen y devuelve todos los estados. Eso equivale a `status=all` de Node. Falta el modo bandeja (que en Node es el valor por defecto) y el filtro por estado de los cierres.
- **Estrategia:**
  1. `GET /admin/course-requests?status=`: con `all`, el comportamiento actual. Sin `status` o con otro valor, la bandeja de Node: `(r.status='under_review' AND c.faculty=F) OR (r.status='approved' AND NOT c.has_documentation AND c.origin_faculty=F)`; para el admin DEU sin facultad, lo mismo sin la condición de facultad. El default pasa a ser la bandeja, como en Node; el cliente de diplomados siempre manda `status`.
  2. `GET /admin/course-cycle-close-requests?estado=`: `under_review` por defecto, `all`, o un estado concreto, en el vocabulario de Node o de Go (§7.1).
  3. Respuesta de cierres: Node usa la clave `cierres` y Go `solicitudes`. Cambiar a `cierres`; si `feature/backend-go` ya lee `solicitudes`, devolver ambas mientras se ajusta.
  4. Cada cierre del listado: agregar `usuario_id` (= `enviado_por_id`), `curso_id`, `fecha` (= `creado_en`), `motivo_rechazo` (§5.10) y `payload{titulo_curso, nombre_cohorte, observaciones}`. `payload.archivos` se queda solo en el detalle, como hoy: el listado de la UI no lo usa, y así se evita una consulta de archivos por fila.
  5. Paginación según §7.2: hoy la bandeja con Go mostraría solo los 10 primeros de cada fuente.
- **Re-evaluación contra `feature/backend-go`:** parcial.
  - La **bandeja** se arma en el cliente: `getAllSolicitudes` trae las tres fuentes (`pageSize`/`per_page` = 100) y deja pasar solo las solicitudes `under_review` o `approved` sin `tiene_documentacion_legal`. Eso equivale a la bandeja de Node.
  - El **historial** (`status=all`) no funciona. `admin/solicitudes/page.tsx` lo pide siempre desde `9a052dd`, que llegó con el rebase. Con ese valor, los cursos se siguen filtrando como bandeja, los cierres se filtran por `c.estado === 'all'` y quedan vacíos, y `admin/providers?status=all` responde 500 en Go (§6.3), así que los proveedores desaparecen sin aviso, porque `Promise.allSettled` se traga el error.
  - **Estrategia:** en el cliente, con `status === 'all'`, no filtrar cursos ni cierres. En Go, hacer §6.3 para que `admin/providers` acepte `status=all`. El modo bandeja y `?estado=` en el servidor pasan a ser opcionales.

### 5.9 Remisión: solicitud indirecta y límites de la UI (⚪ Baja) — Parcial

- **Node:** la acción `remit` deja la solicitud en `under_review`, cambia la facultad y el coordinador actuales y pone `tipo_curso='formulacion-curso-indirecta'`. `admin-actions.tsx` agrega dos reglas:
  - una solicitud indirecta no se puede remitir de nuevo ("No pueden ser remitidas") y su clasificación queda fija;
  - remitir a la propia facultad no hace nada ("no necesitas remitir").
- **Go:** `RedirectCourseRequest` deja `under_review` y cambia `faculty` (✓), pero permite remitir varias veces y remitir a la facultad actual.
- **Estrategia:** en `course_requests.service.RedirectCourseRequest`, responder 409 si `course.Faculty != course.OriginFaculty` (ya es indirecta) y 400 si la facultad destino es la actual. `tipo_curso` sale calculado (§2.13). La clasificación fija de las indirectas es una regla de la UI: Node no la valida en el servidor, así que en Go puede quedar como validación opcional en `ApproveCourseRequest`.
- **Re-evaluación contra `feature/backend-go`:** la regla de las indirectas funciona como en Node. El tipo sale de `formulationType` (`facultad != facultad_origen`) y `admin-actions.tsx` no ofrece remitirlas. Falla la otra regla: `isRemissionSelfHandled` compara la facultad elegida (ahora un valor del enum) con `currentUserId`, así que nunca se cumple y un coordinador puede remitir a su propia facultad. Estrategia: comparar con la facultad de la sesión (`user.facultad`). Las guardas en Go pasan a ser opcionales, porque Node tampoco las tiene en el servidor.

### 5.10 Motivo de rechazo: campo y persistencia (🟡 Media) — Parcial

- **Cierres:** Node guarda por separado las observaciones del solicitante (`payload.observaciones`) y el motivo del revisor (`motivo_rechazo`). Go tiene una sola columna, `comments`. `RejectCourseCycleCloseRequest` (`backend/internal/postgres_repository/queries/course_cycle_close_request_queries.go:90`) la sobrescribe con el motivo, y se pierden las observaciones del solicitante.
- **Solicitudes de curso:** Node expone `motivo_rechazo`. Go lo guarda en `comments` y lo expone como `comentarios`. `_adaptSolicitud` (detalle) lee `comentarios || motivo_rechazo`, pero el listado admin no lee `comentarios`.
- **Estrategia:**
  - cierres: agregar la columna `rejection_reason` (migración), expuesta como `motivo_rechazo`; `observaciones` queda para el solicitante;
  - solicitudes de curso: agregar el alias `motivo_rechazo` (= `comments`) cuando el estado es rechazado. No hace falta migrar, porque ya hay una columna por solicitud.
- **Re-evaluación contra `feature/backend-go`:** las solicitudes de curso quedan resueltas en el cliente: `courseRequestToSolicitud` expone `motivo_rechazo = comentarios`. En los cierres sigue la brecha en Go: rechazar sobrescribe las observaciones del solicitante.

---

## 6. Proveedores

Los commits recientes de `diplomados` tocaron sobre todo esta parte, que el documento no cubría. En Node, las rutas son `POST`/`GET /api/providers`, `GET /api/admin/providers` y `POST /api/provider-requests/[id]/[action]`. En Go están en `/providers` (JWT) y `/admin/providers` (`backend/internal/providers`).

### 6.1 Alta de proveedor: contrato de entrada incompatible (🔴 Alta) — ✅ Resuelto

- **Node:** `POST /api/providers` recibe `userId`, `coordinador_id` (400 `Debe seleccionar una facultad destino.` si falta), `tipo_persona`, `tipo_lucro`, `nombre`, `bio` y `es_interno`, más los archivos:
  - obligatorios: `logo`, `ci`, `rif`, `resumes` y `titulo`;
  - `registro_mercantil`, obligatorio si la persona es jurídica;
  - `islr`, que pasó a ser **opcional** en los últimos commits.

  `solicitudes-service.js` traduce el formulario a estos nombres y ahora envía `tipo_persona='juridica'`, ya no `'juridical'`. Responde 201 con el proveedor creado.
- **Go:** `makeProviderFromRequest` (`backend/internal/providers/endpoints.go`):

  | Aspecto | Go hoy | Efecto contra el cliente actual |
  | --- | --- | --- |
  | `tipo_persona` | solo `natural`/`juridical` (también hay un CHECK en la base) | 400 para toda persona jurídica |
  | `islr` | obligatorio | 400 si no se adjunta |
  | `titulo`, `registro_mercantil` | no existen | se pierden y no se exige el registro mercantil |
  | facultad | lee `facultad` (opcional) e ignora `coordinador_id` | el proveedor queda con facultad `NULL`: ningún coordinador lo ve (§6.3) y sus cursos caen en `DEU` (§2.7) |
  | respuesta | 201 `{id}` | — |

  Además, `nombre_proveedor` en la respuesta es el nombre del usuario, no el `nombre` que se envió (§6.4).
- **Estrategia:**
  1. Aceptar `tipo_persona` con `juridica` o `juridical`. Guardar `juridical`, que es lo que permite el CHECK actual, y responder `juridica` (§6.4). Traducir en los bordes evita migrar el CHECK. *(Descartado: `juridical` era una errata y se migró el CHECK a `juridica`; ver "Resuelto" más abajo.)*
  2. `islr` pasa a opcional.
  3. Nuevos tipos de archivo `titulo` (obligatorio) y `registro_mercantil` (obligatorio si es jurídica), con los mismos mensajes de Node. La tabla `files` es genérica, así que no hace falta migrar.
  4. Aceptar `coordinador_id`: buscar a ese usuario, verificar que sea `faculty_admin` y usar su facultad como `provider.Faculty`. Si no llega ni `coordinador_id` ni `facultad`, responder 400 con el mensaje de Node. Depende de §4.1, porque la UI obtiene ese ID de `GET /users?rol=coordinador`.
- **Re-evaluación contra `feature/backend-go`:** `createSolicitud` (caso `codigo-proveedor`) ya resuelve tres cosas en el cliente:
  - la facultad: envía `facultad` con el valor que eligió el usuario de la lista fija (§4.1);
  - `titulo` y `registro_mercantil`: van como `others`;
  - el resto de los nombres de campo.

  Siguen abiertos dos rechazos que dejan sin poder registrarse a usuarios reales:
  - **`tipo_persona='juridica'`:** el rebase dejó el cambio de `main`, y Go responde 400 (`tipo_persona must be 'natural' or 'juridical'`) a toda persona jurídica;
  - **ISLR opcional:** el formulario ya no lo exige y el servicio solo lo envía si existe; Go responde 400 (`islr is required`).

  También sigue sin validarse en el servidor que una persona jurídica adjunte el registro mercantil; hoy solo lo valida el formulario. Estrategia: los puntos 1–3 de arriba, en Go; los puntos 4 y 5 ya no hacen falta.
- **Resuelto (2026-10-06):**
  - **`juridica`:** `juridical` era una errata. Además, la entidad decía `juridico` mientras el CHECK de la base aceptaba `juridical`, así que una persona jurídica fallaba siempre (400 o 500). Ahora el único valor es `juridica`, en la entidad (`PartyTypeJuridical`), en la base (migración `000039`, que también convierte las filas existentes) y en la API. No se acepta otro alias.
  - **ISLR:** es opcional.
  - **Archivos nuevos:** `titulo` (obligatorio) y `registro_mercantil` (obligatorio si la persona es jurídica) son tipos de archivo propios. `prepareFilesSlice` omite los archivos opcionales que no llegan, en lugar de fallar con un puntero nulo.
  - **Respuesta:** expone `tipo_persona`, `archivos.titulo` y `archivos.registro_mercantil`.
  - **Cliente (`feature/backend-go`):** envía `titulo` y `registro_mercantil` con su nombre, ya no como `others`. `providerToLegacy` los lee desde `archivos` y solo cae en `otros[0]`/`otros[1]` para proveedores registrados antes de este cambio.
  5. Responder 201 con el proveedor creado (§6.4).

### 6.2 Solicitudes de proveedor duplicadas (🟡 Media) — ✅ Resuelto

- **Node:** antes de crear, busca proveedores del mismo usuario. Si alguno está en revisión, responde 409 `Ya tienes una solicitud de colaborador en revisión.`; si alguno está aprobado, 409 `Ya eres un colaborador aprobado.` Un rechazo previo no impide volver a solicitar.
- **Go:** no lo verifica, y `deu.providers.user_id` no es único. Un usuario puede tener varios proveedores, y `GetProviderByUserID` (que usan cursos y contratos) toma uno cualquiera.
- **Estrategia:** en `providers.service.CreateProvider`, buscar proveedores no borrados del usuario y responder 409 con los mensajes de Node si alguno está `under_review` o `active`. Como respaldo, un índice único parcial `ON providers(user_id) WHERE deleted_at IS NULL AND status IN ('under_review','active')`.
- **Re-evaluación contra `feature/backend-go`:** sigue abierta. La UI tampoco lo impide. `checkPendingProviderRequest` espera un arreglo de `GET /providers?usuario_id=`, Go responde `{proveedores}` sin ese filtro, y la verificación siempre da `false` (§10).
- **Resuelto (2026-10-06):**
  - `providers.service.CreateProvider` busca el proveedor vigente del usuario. Si está `under_review`, devuelve `ErrProviderRequestPending`; si está `active`, `ErrProviderAlreadyApproved`. `POST /providers` responde 409 en los dos casos. Un proveedor `rejected` o `inactive` puede volver a solicitar, como en Node.
  - `GetProviderByUserID` filtra `deleted_at IS NULL` y toma el más reciente (`ORDER BY created_at DESC LIMIT 1`). Así el login, los cursos y "mis solicitudes" ya no eligen una fila al azar cuando un usuario tiene un proveedor rechazado y otro nuevo.
  - El cliente traduce los dos 409 a los mensajes de Node.
  - No se agregó el índice único parcial. Si en una base existente ya hay duplicados, la migración fallaría; queda como seguimiento opcional, después de revisar los datos.

### 6.3 Listado de proveedores: estado y visibilidad (🟡 Media)

- **Node:** `GET /api/providers` y `GET /api/admin/providers` aceptan `status`, que vale `all` por defecto o un valor que se compara con `estado`. El admin ve todos; un coordinador ve solo los proveedores con su `coordinador_id`; cualquier otro usuario recibe `[]`. La respuesta es `{proveedores: [...]}`.
- **Go:** `GetProviders` (`backend/internal/providers/endpoints.go`) y `queries.GetProviders`:
  - `status` solo se aplica si `isAdmin(roles)`, y esa función incluye `faculty_admin`, `deu_admin` y `group_admin`, pero **no** `root`;
  - sin `status`, devuelve solo proveedores **activos**, también al admin, así que la bandeja de pendientes queda vacía;
  - `status=all` se pasa tal cual a un enum de Postgres y termina en 500;
  - un usuario que no es admin recibe el directorio de proveedores activos, no `[]`.
- **Estrategia:**
  1. `status` acepta `all` (sin filtro) y los vocabularios de Node y de Go (§7.1). Para los admins, el default es `all`, como en Node.
  2. Agregar `root` a `isAdmin`. `group_admin` se queda, porque lo usa `grupos/`.
  3. Visibilidad: el admin ve todos y `faculty_admin` los de su facultad, que equivale al `coordinador_id` de Node siempre que se haga §6.1. Para cualquier otro usuario, devolver solo **su propio** proveedor (`usuario_id` = el del token) en vez de `[]`. Así funcionan `checkPendingProviderRequest` y `hasInitialContract` (`users-service.js`), que consultan `?usuario_id=`. Node dejó de soportar ese filtro en `7a6a24d`, y esas dos verificaciones quedaron rotas en el mock.
  4. Se mantiene `{proveedores, paginas}`, que coincide con Node. Esos dos helpers esperan un arreglo plano, así que fallan contra ambos backends. Es un defecto del cliente y se reporta a quienes mantienen el submódulo.
- **Re-evaluación contra `feature/backend-go`:** sigue abierta, y ahora tiene impacto directo. La bandeja admin pide `status=all` (§5.8) y Go responde 500. El cliente ya no depende de `?usuario_id=` para los datos del propio proveedor: `findProviderByUserId` usa el `provider_id` de la sesión. Pero un admin que abre el perfil de un proveedor pendiente no lo encuentra, porque ese método lista sin `status` y Go solo devuelve proveedores activos.

### 6.4 Forma de la respuesta del proveedor (🟡 Media) — Parcial (casi todo resuelto en el cliente)

| Node | Go |
| --- | --- |
| `nombre_proveedor` = el `nombre` enviado | `nombre_proveedor` = nombre y apellido del usuario |
| `es_interno` | `interno` |
| `tipo_persona` (`natural` / `juridica`) | no se expone |
| `archivos{logo, ci, rif, islr, curriculum, titulo, registro_mercantil}` | `archivos{ci, rif, islr, resumenes[], otros[]}`; el logo va aparte, en `provider_avatar_url` |
| `coordinador_id` | `facultad` |
| `fecha_creacion`, `fecha_actualizacion` | no se exponen |
| `motivo_rechazo` | no existe (§6.5) |
| `legal_status{tiene_carta_intencion, carta_intencion_url, carta_compromiso_url, adendas[{id_adenda, archivo_url, compromiso_url, fecha, cursos_amparados}]}` | `contratos_legales[{id, tipo, creado_en, cursos_amparados, archivos}]` |
| `estado`: `under_review` / `aprobada` / `rechazada` | `estado`: `under_review` / `active` / `rejected` / `inactive` |

- **Uso en la UI:** `codigo-proveedor-view.tsx` muestra los documentos desde `archivos.*`, `profile-coordinator-review.tsx` lee `legal_status` y `solicitudes-service.js` lee `es_interno || interno`.
- **Estrategia:** agregar los campos de Node sin quitar los de Go:
  - `nombre_proveedor` = `provider.Name` (con el nombre del usuario como respaldo);
  - `es_interno`, `tipo_persona`, `fecha_creacion` y `fecha_actualizacion`;
  - `archivos.logo`, `archivos.curriculum` (el primer CV), `archivos.titulo` y `archivos.registro_mercantil`;
  - `coordinador_id`: el `faculty_admin` de la facultad del proveedor;
  - `motivo_rechazo` (§6.5);
  - `legal_status`, derivado de los contratos: el `inicial` llena `tiene_carta_intencion` y las URLs de las cartas, y cada `adenda` es un elemento de `adendas` con `id_adenda = 'ADENDA-<id>'`, igual que en §2.13;
  - `estado` traducido según §7.1.
- **Re-evaluación contra `feature/backend-go`:** `providerToLegacy` ya traduce `estado`, `es_interno`, `archivos.{logo, curriculum, titulo, registro_mercantil}`, `coordinador_id` (= la facultad) y `legal_status`. Faltan cosas que solo Go puede dar:
  - ~~`tipo_persona`, que no se expone: `codigo-proveedor-view.tsx` muestra `null`~~ — resuelto con §6.1;
  - `motivo_rechazo`, que no se guarda (§6.5);
  - `nombre_proveedor`, que sigue siendo el nombre del usuario.

  ~~Además, `titulo` y `registro_mercantil` salen de `otros[0]` y `otros[1]`: dependen del orden de subida y se cruzan si falta uno.~~ Resuelto con §6.1. Estrategia: los tres campos en Go (`tipo_persona`, `rejection_reason` y `provider.Name`) y tipos de archivo propios para `titulo` y `registro_mercantil` (§6.1, punto 3).

### 6.5 Aprobación y rechazo de proveedores (🟡 Media) — Parcial

- **Node:** `POST /api/provider-requests/[id]/approve|reject`, permitido al admin o a cualquier coordinador, con body `{observaciones}`.
  - Aprobar pone `estado='aprobada'` y cambia al usuario a `rol='proveedor'`, con `codigo_proveedor` = ID del proveedor.
  - Rechazar guarda `motivo_rechazo`.
  - Responde 200 `{message, proveedor}`.
- **Go:** `POST /admin/providers/:id/approve|reject` responde 202 con cuerpo `null`. Aprobar cambia el rol de `visitante` a `course_admin` (que la UI ve como `proveedor`, §5.7) y envía el correo en la misma transacción. El motivo de rechazo solo viaja en el correo; no se guarda.
- **Hallazgo:** también existe `/admin/provider-requests/*` (`backend/internal/provider_requests`), que opera sobre `deu.provider_requests`. Nadie llena esa tabla: `CreateProviderRequest` está en la interfaz del repositorio, pero ningún servicio la llama. Por eso sus listados salen vacíos y sus aprobaciones no tocan al proveedor. La ruta de Node (`/provider-requests/:id/:action`) corresponde a la lógica de `providers`, no a esta.
- **Estrategia:**
  1. Guardar el motivo en una columna nueva `providers.rejection_reason` (migración), expuesta como `motivo_rechazo`.
  2. Responder 200 `{message, proveedor}` con el proveedor actualizado.
  3. Retirar `internal/provider_requests`, o hacer que delegue en `providers.ApproveProvider`/`RejectProvider`.
  4. **Excepción:** `codigo_proveedor` sigue siendo el código que genera Go (`ECP-…`), no el ID. Ya lo usan `GET /courses?codigo_proveedor=` y el filtro por tipo (`ECP-%`/`GEX-%`), y la UI lo trata como un valor opaco (`getProviderDetails` también busca por `usuario_id`).
- **Re-evaluación contra `feature/backend-go`:** `updateStatus` ya llama a `admin/providers/:id/approve|reject`, y `ApiService` tolera el 202. Pero el rebase trajo de `main` (`804c0b3`) un `fetch` directo en `admin-actions.tsx` que se salta ese servicio (§10). El motivo de rechazo sigue sin guardarse en Go.

### 6.6 Estado legal: adjunto opcional de la adenda (⚪ Baja)

- **Node:** `POST /api/legal-status/[user_id]` acepta en las adendas un archivo opcional `carta_compromiso_adenda` y lo guarda como `compromiso_url`. La ruta usa el ID del **usuario**.
- **Go:** `POST /admin/providers/:id/legal-contracts` solo lee `carta_intencion`, `carta_compromiso` y `adenda`, y la ruta usa el ID del **proveedor**. La inferencia inicial/adenda y el amparo automático de cursos son decisiones ya tomadas en §5.2 y no se revisan.
- **Estrategia:** aceptar `carta_compromiso_adenda` como archivo opcional del contrato de tipo `adenda` y exponerlo en `legal_status.adendas[].compromiso_url` (§6.4). Se mantiene el ID del proveedor en la ruta: la UI lo tiene (`provider_table_id` en `getProviderDetails`), y la ruta admin no debería resolver por usuario.
- **Re-evaluación contra `feature/backend-go`:** sigue abierta. `userService.submitLegalDocuments` solo copia `carta_intencion`, `carta_compromiso` y `adenda`, así que la carta de compromiso de la adenda se pierde también en el cliente. Hay que agregarla en los dos lados.

---

## 7. Diferencias transversales

### 7.1 Vocabulario de estados (🟡 Media) — ✅ Resuelto (en el cliente)

| Recurso | Node | Go |
| --- | --- | --- |
| Solicitud de curso | `under_review` / `aprobada` / `rechazada` (remitir la deja en `under_review`) | `under_review` / `approved` / `rejected` / `redirected` / `created` |
| Solicitud de cierre | `under_review` / `aprobada` / `rechazada` | `under_review` / `approved` / `rejected` |
| Proveedor | `under_review` / `aprobada` / `rechazada` | `under_review` / `active` / `rejected` / `inactive` |
| Cohorte | `activa` / `cerrada` | `is_active` (bool), sin campo de estado |
| Curso | un solo `estado` (§2.13) | `activo` + `estado_gestion` |

- **Impacto:** la UI compara literales. `solicitudes-table.tsx` decide qué filas están terminadas y cuándo mostrar el sello legal según `aprobada`/`rechazada`, y lo mismo hace `mis-cursos-client-component.tsx` con sus etiquetas. `solicitudes-service.js` solo traduce `under_review → pendiente`. Con Go, nada se ve aprobado ni rechazado.
- **Estrategia:** traducir en la capa de respuesta de los módulos de cursos (`course_requests`, `course_cycle_close_requests`, `courses` y `providers`) con una función del estilo de `entities.UIRole` (§5.7): `approved → aprobada`, `rejected → rechazada`, `active → aprobada`, `redirected → redirigida`; los períodos exponen `estado` como `activa`/`cerrada`. Los filtros de entrada (`status`, `estado`) aceptan ambos vocabularios. La base no cambia: `grupos/` usa `approved`/`under_review` con el mismo `request_status_enum` para las solicitudes de grupo, y traducir solo en estos módulos no lo afecta.
- **Resuelto en `feature/backend-go`:** `adapters.js` traduce los estados (`requestStatusToLegacy`, y `PROVIDER_STATUS` dentro de `providerToLegacy`) y `periodToCohort` pone `estado: 'activa'`. Solo queda el filtro `status=all` de la bandeja (§5.8) y el estado de las cohortes cerradas (§2.14).

### 7.2 Paginación (🟡 Media) — ✅ Resuelto (en el cliente)

- **Node:**
  - los listados admin (`admin/courses`, `admin/closures`, `admin/providers`) no paginan. La página `admin/solicitudes` une los tres y pagina en memoria, 10 por página;
  - `GET /api/courses` pagina solo si recibe `_page`/`page` y `_limit`/`limit`;
  - `GET /api/courses/public` usa `limit`, 15 por defecto.
- **Go:** todos los listados paginan, 10 por defecto, y el parámetro de tamaño cambia según el módulo: `pageSize` en course-requests; `per_page` en cursos, cierres, proveedores y usuarios. Contra Go, la bandeja admin vería solo los 10 primeros elementos de cada fuente.
- **Estrategia:** en los listados de cursos y proveedores, aceptar `limit` y `_limit` como alias del tamaño de página, y `_page` como alias de `page`. En los tres listados admin, si no llega tamaño, devolver todo (con un tope, p. ej. 1000), como Node.
- **Hallazgo lateral:** `CountCourseCycleCloseRequests` no filtra por facultad, así que el total de páginas de `GET /admin/course-cycle-close-requests` está mal cuando se filtra.
- **Resuelto en `feature/backend-go`:** el cliente manda los nombres de Go (`pageSize` en course-requests, `per_page` en el resto) con 100 por fuente en la bandeja, y pagina en memoria como Node. Queda un límite: si una fuente tiene más de 100 pendientes, los que pasen de 100 no se ven.

### 7.3 Respuestas de las acciones (🟡 Media) — ✅ Resuelto (en el cliente)

| Acción | Node | Go |
| --- | --- | --- |
| Aprobar / rechazar / remitir solicitud de curso | 200 `{message, curso}` | 204 sin cuerpo |
| Aprobar / rechazar cierre | 200 `{success, estado}` | 202 sin cuerpo |
| Aprobar / rechazar proveedor | 200 `{message, proveedor}` | 202 `null` |
| Abrir cohorte | 200 `{success, message}` | 201 `{id}` |
| Solicitar cierre | 201 `{success, message}` | 201 `{id}` |
| Crear curso / proveedor | 201 con el objeto creado | 201 `{id}` |

- **Impacto:** `updateStatusWithFile` (la aprobación con evaluación) y la acción de cierre llaman a `response.json()` después de comprobar `ok`. Con un 204 sin cuerpo esa llamada lanza un error, y la UI muestra "Fallo…" aunque la acción sí se hizo. `BaseApiService` sí tolera el 204. El formato de error ya coincide: Go responde `{error}`, igual que Node (`httperrors.NewHTTPErrorHandler`).
- **Estrategia:** responder 200/201 con los cuerpos de Node en estas rutas, incluyendo el recurso actualizado cuando Node lo devuelve. Sobre las rutas: las acciones de revisión se quedan bajo `/admin` en Go, porque de ese grupo cuelga el middleware de roles; ahí cambia la ruta el cliente, no el backend. Solo se agregan alias en rutas que no requieren un rol especial (§2.15). En §2.10 se descartó el alias y el cliente se adapta.
- **Resuelto en `feature/backend-go`:** todas las acciones pasan por `ApiService`, que convierte el 204 en `{success: true}`, y la UI no usa los cuerpos de Node. La única excepción es el `fetch` directo de `admin-actions.tsx` para proveedores (§10).

---

## 8. Solapamiento Curso + Grupo

**No existe ningún endpoint ni entidad que relacione cursos y grupos en ninguno de los dos backends hoy.**

- **Go:** `internal/activities` (`entities.Activity`) y `internal/group_resource_requests` (`entities.GroupResourceRequest`) son exclusivamente de grupos — ninguna de sus entidades tiene un campo `CourseID`/`CoursePeriodID`, confirmado por revisión directa.
- **Node:** las colecciones `extension-groups`, `group-requests`, `group-members`, `group-resource-requests` y `activities` existen en `db.json` pero están **vacías y sin ninguna ruta ni servicio que las use** — son remanentes de un modelo que nunca se llegó a implementar en el mock. La única referencia visible a "grupo" en la UI de administración (pestaña "Grupo de Extensión" en `diplomados/src/app/admin/usuarios/page.tsx` y `solicitudes-table.tsx`) está respaldada por arrays hardcodeados en el frontend, no por datos reales.

No hay, por lo tanto, brechas que documentar en esta sección — se deja constancia de que la revisión se hizo y no encontró superficie compartida real en ninguno de los dos sistemas.

---

## 9. Dependencias

Algunos ítems de este documento no se pueden resolver de forma aislada — dependen de que otro ítem exista primero. Esta sección deja constancia explícita de esas dependencias para evitar que se aborden en el orden equivocado.

- **§2.3 (listado público filtrado por curso "listo") depende de §5.2 (paso de "estado legal"/contrato).** La regla de Node para `GET /api/courses/public` filtra explícitamente por la presencia de un contrato legal (`contrato_id`/`documento_legal_id`). Go no tiene ningún concepto de contrato/documento legal asociado a un curso o proveedor (§5.2), así que no existe ningún campo real sobre el cual replicar ese filtro. Implementar §2.3 con un proxy inventado (p. ej. "aprobado y con al menos un período histórico") sería una interpretación de producto no confirmada, no una migración fiel de la regla de Node — por eso se decidió **no** implementar §2.3 todavía y esperar a que §5.2 exista primero.
- **§2.1 (apertura de período) NO depende de §5.2.** A diferencia de §2.3, la guarda de apertura de período implementada no necesita el concepto de contrato: basta con que el curso esté aprobado (`courses.is_active`, ya existente) y no tenga un período activo sin cerrar. Se señala explícitamente para que quede claro que §2.1 pudo resolverse de forma independiente y ya está hecho (ver §2.1).
- **§5.4 (cascada de cierre) queda parcialmente ligada a §5.2/§3.1.** Como se detalla en §5.4, Go ya cierra correctamente el período/ciclo específico al aprobar una solicitud de cierre; lo único que falta es reflejar ese cierre en un estado a nivel de **curso** (`estado_gestion`), que hoy no existe en la entidad `Course`. Modelar ese campo probablemente ocurra junto con — o como parte de — el trabajo de §5.2, ya que ambos requieren extender `entities.Course` con nuevos campos de estado/ciclo de vida.

**Estado al 2026-09-28:** las tres dependencias de arriba ya están saldadas. §5.2 se resolvió y con eso §2.3 se pudo implementar sobre `has_documentation`. `estado_gestion` se modeló (`0fc656c`/`c3788d0`) y con eso se cerraron §5.4 y el bloqueo pendiente de §2.2. Dependencias vigentes entre las brechas abiertas:

- **§5.8 (bandeja de "aprobados sin contrato") dependía de §2.6 (facultad de origen)** para la parte de coordinadores: Node asigna esa bandeja por facultad de origen. §2.6 ya está resuelta, así que §5.8 puede filtrar por `courses.origin_faculty`.
- **§2.6 y §2.7 se resolvieron juntas.** Al final sí se agregó la columna `courses.origin_faculty`, porque `providers.faculty` se puede editar y no sirve como origen inmutable.
- **§3.1 depende de §5.3:** los campos restantes de `Course` (`clasificacion`, `calificacion`, evaluación) solo existen si se capturan al aprobar.

**Estado al 2026-10-06:** dependencias entre las brechas nuevas.

- **§6.1 depende de §4.1 (filtro por rol):** la UI obtiene el `coordinador_id` que envía al registrar un proveedor de `GET /users?rol=coordinador`.
- **§6.3 depende de §6.1:** el coordinador solo ve los proveedores de su facultad, y hoy los proveedores creados desde la UI quedan sin facultad.
- **§2.3 y §2.8 dependen de §2.16:** filtrar solo por `estado_gestion` exige que los cursos amparados queden en `cerrado`. Resuelto el 2026-10-07.
- **§5.9 se apoya en §2.13:** `tipo_curso` calculado y la guarda de remisión usan la misma condición (`faculty != origin_faculty`).
- **§5.8 y §6.3 dependen de §7.1:** sus filtros tienen que aceptar el vocabulario de Node.
- **§5.10 y §6.5 llevan migración** (`rejection_reason` en cierres y en proveedores); conviene hacerlas en el mismo lote.

**Orden sugerido:**
1. Lo que hoy bloquea el uso de la UI contra Go: §4.1 (filtro por rol) → §6.1 → §6.2 → §2.11 → §2.12.
2. Reglas de negocio: §2.16 → §2.10 → §2.3/§2.8.
3. Forma de las respuestas: §7.1 → §2.13 → §2.14 → §6.4 → §7.3 → §7.2 → §5.8 → §2.15.
4. El resto: §5.9, §5.10, §6.3, §6.5, §6.6, §4.3, §4.1 (nombres de campos).

**Orden actualizado tras revisar `feature/backend-go`:**
1. Defectos de la rama (§10), que hoy rompen la compilación y los flujos de admin.
2. ~~Lo que bloquea registros reales en Go: §6.1 (`juridica` e ISLR).~~ Resuelto.
3. Reglas de negocio de Go: ~~§2.12~~ (resuelto) → ~~§2.16~~ → ~~§2.3~~ (resueltos), y §6.3 (`status=all`). §6.2 y §10.4 ya resueltos.
4. El resto: §2.14 (historial de cohortes), §2.13 (dos huecos del adaptador), §5.8 (historial en el cliente), §5.10, §6.2, §6.4, §6.5, §6.6, §5.9.

---

## 10. Defectos de la rama `feature/backend-go`

Los encontró la segunda pasada. No son diferencias entre Node y Go, sino defectos del cliente adaptado; la mayoría llegó al rebasear sobre `main@312447a`, cuyos commits todavía hablan con el mock.

### 10.1 `users-service.js` no compila (🔴 Alta) — ✅ Resuelto

- **Qué pasa:** en `getProviderDetails` quedó mezclado el bloque "anti-crasheo" de `main`, que hace `ApiService.get('providers')` y un `find` sobre el resultado, con la versión de la rama, que usa `findProviderByUserId`. El `if` queda sin cerrar y `providerData` se declara dos veces. `next lint` falla con `Parsing error: Identifier 'providerData' has already been declared`.
- **Impacto:** `next build` falla y ninguna página que importe `userService` funciona.
- **Arreglo:** quitar el bloque de `main` y dejar solo la versión con `findProviderByUserId` + `providerToLegacy`.
- **Resuelto (2026-10-06):** se restauró la versión anterior al rebase. `next lint` y `tsc --noEmit` pasan sin errores en el archivo.

### 10.2 Aprobar o rechazar un proveedor usa una ruta de Node (🔴 Alta) — ✅ Resuelto

- **Qué pasa:** `admin-actions.tsx` llama directamente a `fetch('/api/provider-requests/${solicitudId}/${action}')`. Esa llamada viene de `main` (`804c0b3`). El proxy la reenvía a Go como `/provider-requests/prov-<id>/<action>`, que no existe: la ruta de Go está bajo `/admin/providers/:id`, y el ID sin el prefijo `prov-`.
- **Impacto:** el admin no puede aprobar ni rechazar proveedores.
- **Arreglo:** volver a usar `solicitudesService.updateStatus(solicitudId, solicitudTipo, nuevoEstado, motivo_rechazo)`, que ya llama a `admin/providers/:id/approve|reject` con el ID correcto.
- **Resuelto (2026-10-06):** `admin-actions.tsx` vuelve a delegar en `solicitudesService.updateStatus`.

### 10.3 El historial de la bandeja (`status=all`) sale incompleto (🟡 Media)

Detalle en §5.8. Con `status=all`:
- los cursos se filtran como bandeja;
- los cierres quedan vacíos;
- los proveedores fallan con 500 (§6.3).

### 10.4 Verificación de "solicitud de proveedor pendiente" (🟡 Media, antes ⚪ Baja) — ✅ Resuelto

- **Qué pasa:** `checkPendingProviderRequest`, que viene de `main` y usa `navbar.tsx`, espera un arreglo de `GET /providers?usuario_id=`. Go devuelve `{proveedores, paginas}` sin ese filtro, así que siempre da `false`.
- **Impacto:** antes quedaba oculto, porque un proveedor pendiente ya tenía `codigo_proveedor` en el token (§2.12) y la navbar no le mostraba "Solicitar alianza". **Desde que se resolvió §2.12 ya no está oculto:** el proveedor pendiente no tiene código, la navbar le ofrece "Solicitar alianza", y como esta verificación siempre da `false` y Go no impide duplicados (§6.2), puede enviar otra solicitud.
- **Arreglo:** si se corrige §2.12, usar `findProviderByUserId` (con el `provider_id` de la sesión) y revisar `estado === 'under_review'`.
- **Resuelto (2026-10-06):** `checkPendingProviderRequest` usa `findProviderByUserId` y devuelve `true` si el proveedor propio está `under_review`. **Limitación aceptada:** el `providerID` se escribe en el token al iniciar sesión, así que justo después de enviar la solicitud, y hasta volver a iniciar sesión, la navbar sigue mostrando "Solicitar alianza". Si se reenvía, Go responde 409 (§6.2) y el formulario muestra "Ya tienes una solicitud de colaborador en revisión.".

---

## 11. Tabla resumen

| #   | Brecha                                                                                                                         | Sección          | Prioridad | Estado |
| --- | ------------------------------------------------------------------------------------------------------------------------------ | ---------------- | --------- | ------ |
| 1   | Sin paso de "estado legal"/contrato que habilita un curso aprobado                                                             | §5.2             | 🔴 Alta    | ✅ Resuelto |
| 2   | Sin cascada de cierre de cohorte a nivel de curso (el período sí se cierra; falta reflejarlo en un estado de curso inexistente) | §5.4             | 🟡 Media   | ✅ Resuelto (`estado_gestion` + cierre de todas las cohortes activas) |
| 3   | Sin endpoint de inscripción/gestión de participantes de un período                                                             | §5.1 (ref. §2.5) | ⚪ —       | Fuera de alcance — Node tampoco lo tiene (ver §5.1) |
| 4   | Apertura de cohorte sin reglas de estado ni campo de capacidad                                                                 | §2.1             | 🔴 Alta    | ✅ Resuelto |
| 5   | Aprobación de solicitud sin calificación/clasificación/documento de evaluación                                                 | §5.3             | 🟡 Media   | ✅ Resuelto (multipart con evaluación, migración `000035`) |
| 6   | Solicitud de cierre sin adjuntos obligatorios (participantes/vouchers/encuesta) ni bloqueo del período mientras está pendiente | §2.2             | 🟡 Media   | ✅ Resuelto (adjuntos + guarda de duplicados + bloqueo `solicitud-cierre`) |
| 7   | Listado público de cursos sin filtro por curso "listo" (contrato + estado)                                                     | §2.3             | 🟡 Media   | ✅ Resuelto (forma en el cliente; filtro por `estado_gestion` en Go) |
| 8   | Sin bandeja de "mis solicitudes" para el solicitante                                                                           | §5.5             | ⚪ Baja    | ✅ Resuelto (`GET /course-requests`) |
| 9   | Sin filtro por rol en listado de usuarios (bloquea la lista de facultades al registrar proveedor y al remitir)                 | §4.1             | 🟡 Media   | ✅ Resuelto en el cliente (lista fija de facultades) |
| 10  | Inconsistencia de nombres de campo entre alta y edición de usuario (`cedula`/`nombres` vs. `ci`/`primer_nombre`)               | §4.1             | ⚪ Baja    | Pendiente — sin consumidores en ningún frontend |
| 11  | Campos de clasificación/evaluación/contrato ausentes en la entidad `Course`                                                    | §3.1             | ⚪ Baja    | ✅ Resuelto (evaluación en la solicitud; propuesta completa en el curso) |
| 12  | Modelado de anuncios por período vs. curso+cohorte independientes                                                              | §2.4             | ⚪ Baja    | ✅ Resuelto en el cliente (`periodToCohort`) |
| 13  | Redirección sobrescribe la facultad del curso; no hay facultad/coordinador de origen para dirigir cierres                      | §2.6             | 🟡 Media   | ✅ Resuelto (`courses.origin_faculty`; cierres por facultad de origen) |
| 14  | Facultad del curso la elige el cliente en vez de heredarse del proveedor                                                       | §2.7             | 🟡 Media   | ✅ Resuelto (facultad heredada del proveedor) |
| 15  | Bandejas admin sin filtro por estado ni listado de "aprobados sin contrato"                                                    | §5.8             | 🟡 Media   | Parcial — bandeja resuelta en el cliente; el historial `status=all` falla (cliente + §6.3) |
| 16  | `GET /courses` sin filtros por dueño/estado ni visibilidad por rol                                                             | §2.8             | ⚪ Baja    | ✅ Resuelto; residual sin impacto actual |
| 17  | Cohorte/período sin nombre (`nombre_cohorte`)                                                                                  | §2.9             | ⚪ Baja    | ✅ Resuelto (`nombre_cohorte`) |
| 18  | Abrir cohorte no exige amparo legal; ruta y campos de Node (`/open`, `cohortName`…)                                            | §2.10            | 🔴 Alta    | ✅ Resuelto (regla en el servicio, 409); alias descartado, el cliente se adapta |
| 19  | Formulación de curso: Go no entiende `payload`, `cover`, `titulo`, `proposito`                                                 | §2.11            | 🔴 Alta    | ✅ Resuelto en el cliente (`_courseFormData`) |
| 20  | Un proveedor no aprobado puede formular cursos; un no proveedor recibe 500 en vez de 403                                       | §2.12            | 🟡 Media   | ✅ Resuelto (403 en el servicio; código de proveedor solo para `active` en login y perfil) |
| 21  | Respuesta del curso sin `image_url`, `proposito`, `tipo_curso`, `documento_legal_id` ni `estado` unificado                     | §2.13            | 🟡 Media   | Parcial — `courseToLegacy` lo resuelve; faltan `image_url` y el ID del contrato por curso |
| 22  | Detalle del curso sin `cohortes[]` (con publicaciones) ni `cohorteActiva`                                                     | §2.14            | 🟡 Media   | Parcial — sin historial: Go solo lista períodos activos |
| 23  | Solicitud de cierre por curso (`POST /courses/:id/closures`) en vez de por período                                             | §2.15            | 🟡 Media   | ✅ Resuelto en el cliente (`requestCohortClosure`) |
| 24  | Un curso amparado no pasa a `cerrado` (no se ve en el catálogo hasta abrir cohorte)                                            | §2.16            | 🟡 Media   | ✅ Resuelto (`cerrado` al amparar si está vacío; migración `000040`) |
| 25  | Claims y respuesta del login sin `sub`, `name`, `email`, `facultad`, `codigo_proveedor`                                        | §4.3             | ⚪ Baja    | ✅ Resuelto en el cliente (`session.ts`) |
| 26  | Remisión: se puede remitir una indirecta o a la propia facultad                                                                | §5.9             | ⚪ Baja    | Parcial — indirectas OK; la guarda de "propia facultad" compara con el ID de usuario |
| 27  | Rechazar un cierre sobrescribe las observaciones del solicitante; falta `motivo_rechazo`                                       | §5.10            | 🟡 Media   | Parcial — solicitudes de curso OK en el cliente; cierres pendientes en Go |
| 28  | Alta de proveedor: `juridica`, ISLR opcional, `titulo`, `registro_mercantil`, `coordinador_id`                                 | §6.1             | 🔴 Alta    | ✅ Resuelto (`juridica` en todas las capas, migración `000039`; ISLR opcional; `titulo`/`registro_mercantil` propios) |
| 29  | Solicitudes de proveedor duplicadas                                                                                            | §6.2             | 🟡 Media   | ✅ Resuelto (409 en el servicio; proveedor vigente = el más reciente) |
| 30  | Listado de proveedores: `status=all` da 500, `root` sin filtro, admin solo ve activos, visibilidad distinta                    | §6.3             | 🟡 Media   | Pendiente (Go) — ahora rompe el historial de la bandeja |
| 31  | Respuesta del proveedor: nombres de campos, `legal_status`, `motivo_rechazo`                                                   | §6.4             | 🟡 Media   | Parcial — `providerToLegacy` lo resuelve; faltan `motivo_rechazo` y `nombre_proveedor` |
| 32  | Aprobar/rechazar proveedor: motivo no se guarda, respuesta distinta, módulo `provider_requests` sin uso                        | §6.5             | 🟡 Media   | Parcial — el cliente ya funciona (§10.2 resuelto); falta guardar el motivo en Go |
| 33  | Estado legal: `carta_compromiso_adenda` no se acepta                                                                           | §6.6             | ⚪ Baja    | Pendiente (cliente + Go) |
| 34  | Vocabulario de estados (`approved`/`active` vs. `aprobada`, cohortes sin `estado`)                                             | §7.1             | 🟡 Media   | ✅ Resuelto en el cliente (`adapters.js`) |
| 35  | Paginación: Node no pagina las bandejas admin; nombres de parámetros distintos                                                 | §7.2             | 🟡 Media   | ✅ Resuelto en el cliente (100 por fuente) |
| 36  | Respuestas de acciones (204/202 sin cuerpo vs. 200 con JSON)                                                                   | §7.3             | 🟡 Media   | ✅ Resuelto en el cliente (`ApiService`), salvo §10.2 |
| 37  | `users-service.js` no compila (restos del rebase en `getProviderDetails`)                                                    | §10.1            | 🔴 Alta    | ✅ Resuelto |
| 38  | Aprobar/rechazar proveedor llama a `/api/provider-requests/...` de Node                                                        | §10.2            | 🔴 Alta    | ✅ Resuelto |
| 39  | Historial de la bandeja (`status=all`) incompleto                                                                             | §10.3            | 🟡 Media   | Pendiente (cliente + §6.3) |
| 40  | `checkPendingProviderRequest` siempre da `false`                                                                              | §10.4            | 🟡 Media   | ✅ Resuelto (usa el `providerID` del token; botón visible hasta re-loguearse, reenvío da 409) |

**Sin brechas:** autenticación básica (§4.2), rol por defecto al registrar (§4.1), formato de error `{error}` (§7.3), plantilla de participantes (el `.xlsx` que `diplomados` sirve en `public/formatos/` tiene los mismos encabezados que `GET /course-cycle-close-requests/participants-template`), solapamiento curso+grupo (§8). La ruta nueva `GET /uploads/[...path]` de Node solo sirve su almacenamiento local; Go sirve los archivos desde B2 y `GET /files/*`.
