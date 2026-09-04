erDiagram
    users ||--o{ activities : "group_id"
    %% s1 s2 s3 s4 s5 s6 s7 s8 s9 s10 s11 s12 s13 s14 s15 s16 s17 s18 s19 s20 s21 s22

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