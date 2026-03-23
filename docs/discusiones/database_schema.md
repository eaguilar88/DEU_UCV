# Resumen del Esquema de Base de Datos (PostgreSQL)

Este archivo describe de manera básica la estructura y relaciones principales de la base de datos utilizada en el sistema DEU.

## Tablas principales y relaciones

- **users**: Almacena la información de los usuarios del sistema. Es la tabla base para la mayoría de las relaciones.
- **providers**: Representa a los proveedores (coordinadores de cursos o representantes de grupos de extensión). Cada proveedor está asociado a un usuario (`user_id`), relación uno a uno.
- **roles**: Define los diferentes roles que pueden tener los usuarios.
- **user_roles**: Relaciona usuarios con roles específicos. Permite que un usuario tenga varios roles y un rol pertenezca a varios usuarios (relación muchos a muchos).
- **courses**: Almacena los cursos ofrecidos. Cada curso está asociado a un proveedor (`provider_id`).
- **course_auth_requests**: Solicitudes de aval para cursos. Relacionadas con cursos y revisores (usuarios).
- **course_cycles**: Ciclos o ediciones de los cursos. Cada ciclo pertenece a un curso.
- **visitante**: Visitantes interesados en ver los cursos.
- **extension_groups**: Grupos de extensión universitaria. Cada grupo está asociado a un proveedor.
- **group_auth_requests**: Solicitudes de aval para grupos de extensión. Relacionadas con grupos y revisores (usuarios).
- **activities**: Actividades realizadas por los grupos de extensión. Relacionadas con un grupo.
- **contacts**: Información de contacto (email o teléfono) asociada a usuarios o proveedores.
- **files**: Archivos relacionados a diferentes entidades (cursos, grupos, actividades, ciclos, proveedores). Incluye metadatos y referencia al usuario que subió el archivo.

## Relaciones clave
- Un **usuario** puede ser **proveedor** (relación uno a uno).
- Un **usuario** puede tener varios **roles** (relación muchos a muchos vía `user_roles`).
- Un **proveedor** puede crear varios **cursos** y **grupos de extensión**.
- Un **curso** puede tener varios **ciclos** y cada ciclo varios **participantes**.
- Los **archivos** pueden estar asociados a cursos, grupos, actividades, ciclos o proveedores, usando los campos `owner_type` y `owner_id`.
- Las **solicitudes de aval** pueden estar asociadas a cursos o grupos y a un revisor (usuario).

Este resumen cubre las tablas y relaciones principales. Para detalles adicionales, consultar el archivo `init.sql`.


## Proceso de creación de cursos

Un curso sólo puede ser creado por un proveedor de cursos (código de proveedor ECP-xxxxx). El proveedor debe incluir toda la información solicitada para el curso (las propiedades definidas en la tabla `courses`).

El proceso es el siguiente:

1. El proveedor envía la información del curso a través de un endpoint de solicitudes.
2. El backend crea una nueva solicitud en la tabla de solicitudes de aprobación para cursos (`course_auth_requests`).
3. El backend también crea una nueva entrada en la tabla `courses` con la información recibida. Este curso estará oculto para los usuarios (campo `is_active = false`) hasta que la solicitud sea aprobada.
4. Una vez que la solicitud es revisada y aprobada por un administrador, el curso se activa (`is_active = true`) y pasa a estar disponible en el sistema.

Este flujo asegura que todos los cursos sean revisados y aprobados antes de estar disponibles para los participantes.


