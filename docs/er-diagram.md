# DEU UCV — Diagrama Entidad-Relación

Generado a partir de las migraciones en `postgresql/migrations/`. Esquema: `deu`.

> Nota: `contacts` y `files` usan relaciones polimórficas (`owner_type` + `owner_id`), no FKs reales — se muestran como comentario, no como líneas de relación.
>
> Nota: si vas a visualizar este diagrama, evita mermaid.live — su motor de layout tiende a fallar en diagramas grandes como este y su función de auto-reparación puede sobrescribir el archivo con una versión mutilada. Usa la extensión "Mermaid" de VS Code, `mmdc` (mermaid-cli) local, o el archivo `docs/schema.dbml` en dbdiagram.io.

```mermaid
erDiagram
    users ||--o{ providers : "user_id"
    users ||--o{ user_roles : "user_id"
    roles ||--o{ user_roles : "role_id"
    users ||--o{ course_auth_requests : "reviewer_id"
    users ||--o{ group_auth_requests : "reviewer_id"
    users ||--o{ provider_requests : "reviewer_id"
    users ||--o{ course_cycle_close_requests : "submitted_by"
    users ||--o{ course_cycle_close_requests : "reviewer_id"
    users ||--o{ visitantes : "user_id"
    users ||--o{ extension_groups : "user_id"
    users ||--o{ files : "uploaded_by"

    providers ||--o{ courses : "provider_id"
    providers ||--o{ provider_requests : "provider_id"

    courses ||--o{ course_auth_requests : "course_id"
    courses ||--o{ course_cycles : "course_id"

    course_cycles ||--o{ course_cycle_announcements : "course_cycle_id"
    course_cycles ||--o{ visitantes : "course_cycle_id"
    course_cycles ||--o{ course_cycle_close_requests : "course_cycle_id"

    extension_groups ||--o{ group_auth_requests : "group_id"
    extension_groups ||--o{ group_members : "group_id"
    extension_groups ||--o{ group_resource_requests : "group_id"
    extension_groups |o--o{ activities : "group_id"

    users {
        int id PK
        int ci UK
        varchar email UK
        varchar first_name
        varchar last_name
        date date_of_birth
        varchar gender
        education_level_enum education
        varchar address
        char password
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    providers {
        int id PK
        int user_id FK
        varchar name
        varchar party_type
        varchar profit_type
        bool is_internal
        text bio
        varchar code
        provider_status_enum status
        faculty_enum faculty
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    roles {
        int id PK
        varchar name UK
        text description
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    user_roles {
        int id PK
        int user_id FK
        int role_id FK
        text domain_type
        faculty_enum faculty
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    courses {
        int id PK
        varchar name
        text description
        int provider_id FK
        text objectives
        text rationale
        varchar duration
        varchar cost
        text instructor_profile
        text profiles
        text requirements
        text content
        text evaluation
        text schedule
        course_type_enum type
        faculty_enum faculty
        course_location_enum location
        bool is_active
        bool has_documentation
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    course_auth_requests {
        int id PK
        int course_id FK
        request_status_enum status
        int reviewer_id FK
        text comments
        timestamp reviewed_at
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    course_cycles {
        int id PK
        int course_id FK
        date start_date
        date end_date
        date inscription_date
        bool is_active
        timestamp closed_at
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    course_cycle_announcements {
        int id PK
        int course_cycle_id FK
        varchar title
        text content
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    visitantes {
        int id PK
        int user_id FK
        int course_cycle_id FK
        visitante_status_enum participant_status
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    extension_groups {
        int id PK
        int user_id FK
        varchar name
        text description
        bool is_multidisciplinary
        faculty_enum_array faculty
        date foundation
        text objective
        varchar code
        varchar group_director
        varchar_array type
        varchar location
        bool is_active
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    group_auth_requests {
        int id PK
        int group_id FK
        request_status_enum status
        faculty_enum faculty
        int reviewer_id FK
        timestamp reviewed_at
        text comments
        timestamp created_at
        timestamp updated_at
    }

    group_members {
        int id PK
        int group_id FK
        varchar name
        int ci
        varchar phone
        varchar email
        varchar coordination
        varchar year
        faculty_enum faculty
        varchar school
        varchar document
        bool is_leader
        bool is_active
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    group_resource_requests {
        int id PK
        int group_id FK
        varchar type
        text content
        request_status_enum status
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    activities {
        int id PK
        varchar name
        int group_id FK
        text description
        date date_start
        date date_end
        varchar location
        varchar_array knowledge_area
        varchar allies
        int group_participants
        int stimated_participants
        int actual_participants
        varchar financing
        text comments
        varchar gallery_url
        bool report_checked
        bool is_featured
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    provider_requests {
        int id PK
        int provider_id FK
        request_status_enum status
        int reviewer_id FK
        text comments
        timestamp reviewed_at
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    course_cycle_close_requests {
        int id PK
        int course_cycle_id FK
        int submitted_by FK
        request_status_enum status
        text comments
        int reviewer_id FK
        timestamp reviewed_at
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    contacts {
        int id PK
        contact_type_enum contact_type
        text contact_value
        owner_type_enum owner_type "polymorphic: user, provider, group"
        bigint owner_id "polymorphic FK, no constraint"
        timestamptz created_at
        timestamptz updated_at
        timestamp deleted_at
    }

    files {
        int id PK
        text owner_type "polymorphic: course, group, group_activity, course_cycle, provider, user"
        int owner_id "polymorphic FK, no constraint"
        text file_key
        text purpose
        int version
        bool public
        jsonb metadata
        int uploaded_by FK
        timestamp created_at
        timestamp deleted_at
    }
```

## Notas

- **Relaciones polimórficas**: `contacts.owner_id`/`owner_type` y `files.owner_id`/`owner_type` no tienen FK real en la base de datos; pueden apuntar a `users`, `providers`, `extension_groups`, `courses`, `course_cycles`, o actividades de grupo según el tipo.
- **Soft deletes**: la mayoría de las tablas tienen `deleted_at` para borrado lógico.
- **Enums definidos**: `education_level_enum`, `faculty_enum`, `provider_status_enum`, `course_location_enum`, `course_type_enum`, `request_status_enum`, `visitante_status_enum`, `contact_type_enum`, `owner_type_enum`.
- `activities.group_id` permite `NULL` (ON DELETE SET NULL), por eso la relación con `extension_groups` es opcional (`|o--o{`).
- Este mismo esquema también está disponible como `docs/schema.dbml` para dbdiagram.io, que maneja mejor diagramas grandes que mermaid.live.
