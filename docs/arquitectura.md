# Arquitectura de la Aplicación

## 1. Separación de Responsabilidades (SoC)

La aplicación utiliza **3 tipos de structs** bien diferenciados para mantener una clara separación entre capas:

```
┌─────────────────┐
│ Request/Response│ ← Capa HTTP (serialización/deserialización)
└────────┬────────┘
         │ decoders/encoders
┌────────▼────────┐
│    Entities     │ ← Lógica de negocio (domain layer)
└────────┬────────┘
         │ mappers
┌────────▼────────┐
│     Models      │ ← Capa de persistencia (database layer)
└─────────────────┘
```

### Descripción de Capas

#### Request/Response Structs
- **Ubicación**: `internal/*/request.go` y `internal/*/response.go`
- **Propósito**: Manejar la serialización y deserialización de datos HTTP
- **Tags**: `json`, `form`, `validate`, `path`, `query`
- **Responsabilidad**: Validación de formato y sintaxis de datos de entrada/salida

#### Entities
- **Ubicación**: `internal/entities/*.go`
- **Propósito**: Representar el modelo de dominio y contener la lógica de negocio
- **Tags**: Ninguno (structs puros de Go)
- **Responsabilidad**: Validación de reglas de negocio, comportamiento del dominio

#### Models
- **Ubicación**: `internal/postgres_repository/models/*.go`
- **Propósito**: Mapeo directo con las tablas de la base de datos
- **Tags**: `db` (si se utilizan)
- **Responsabilidad**: Representación de datos tal como se almacenan en la base de datos

### Transformaciones entre Capas

Las transformaciones se realizan mediante funciones mapper:

- **Decoders** (`decoders.go`): Convierten Request → Entity
- **Encoders** (`encoders.go`): Convierten Entity → Response
- **Repository Mappers**: Convierten Model ↔ Entity

## 2. Ventajas del Enfoque

### ✅ Desacoplamiento
Los cambios en la API no afectan la lógica de negocio. Puedes modificar la estructura de tus endpoints sin tocar las entidades del dominio.

**Ejemplo**: Si decides cambiar el nombre de un campo en la API de `first_name` a `firstName`, solo modificas el Request/Response struct, no la entidad.

### ✅ Flexibilidad
Puedes cambiar la base de datos o el ORM sin afectar las entidades de negocio.

**Ejemplo**: Migrar de PostgreSQL a MongoDB solo requiere cambiar los Models y el Repository, las Entities permanecen intactas.

### ✅ Validación por Capa
Cada capa valida lo que le corresponde:

- **Request**: Validación de formato (email válido, campos requeridos, longitud)
- **Entity**: Validación de reglas de negocio (edad mínima, estados válidos)
- **Repository**: Validación de constraints de base de datos (unique, foreign keys)

### ✅ Testabilidad
Cada capa puede ser testeada independientemente:

- Tests de endpoints con Request/Response mocks
- Tests de lógica de negocio con Entities
- Tests de persistencia con Models

### ✅ Versionado de API
Puedes mantener múltiples versiones de la API usando las mismas entidades de dominio.

**Ejemplo**:
```go
// v1/request.go
type CreateUserRequestV1 struct {
    Name string `json:"name"`
}

// v2/request.go
type CreateUserRequestV2 struct {
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
}

// Ambas se convierten a la misma Entity
type User struct {
    FirstName string
    LastName  string
}
```

## 3. Patrones de Diseño Utilizados

### DTO Pattern (Data Transfer Objects)
Los structs `Request` y `Response` actúan como DTOs, transportando datos entre la capa de presentación y la capa de aplicación.

**Ubicación**: `internal/*/request.go`, `internal/*/response.go`

**Ejemplo**:
```go
// DTO para crear un usuario
type CreateUserRequest struct {
    Document       string `json:"ci"`
    Username       string `json:"username" validate:"required,email"`
    FirstName      string `json:"first_name" validate:"required"`
    LastName       string `json:"last_name" validate:"required"`
    Password       string `json:"password"`
}

// DTO para responder con datos de usuario
type GetUserResponse struct {
    ID          string `json:"id,omitempty"`
    Username    string `json:"username,omitempty"`
    FirstName   string `json:"first_name,omitempty"`
    LastName    string `json:"last_name,omitempty"`
    CreatedAt   string `json:"created_at,omitempty"`
}
```

### Domain Model Pattern
Las `entities` representan el modelo de dominio con su comportamiento y reglas de negocio.

**Ubicación**: `internal/entities/*.go`

**Ejemplo**:
```go
type User struct {
    ID             string
    CI             string
    Username       string
    FirstName      string
    LastName       string
    DateOfBirth    string
    Age            int
    Password       string
    Roles          []string
}

// Métodos de negocio en la entidad
func (u *User) IsAdult() bool {
    return u.Age >= 18
}

func (u *User) HasRole(role string) bool {
    for _, r := range u.Roles {
        if r == role {
            return true
        }
    }
    return false
}
```

### Repository Pattern
El repositorio abstrae el acceso a datos, permitiendo cambiar la implementación sin afectar la lógica de negocio.

**Ubicación**: `internal/postgres_repository/`

**Ejemplo**:
```go
type Repository interface {
    CreateUser(ctx context.Context, user entities.User) (string, error)
    GetUser(ctx context.Context, id string) (entities.User, error)
    UpdateUser(ctx context.Context, user entities.User) error
    DeleteUser(ctx context.Context, id string) error
}
```

### Mapper/Converter Pattern
Funciones dedicadas a transformar datos entre diferentes representaciones.

**Ubicación**: `internal/*/decoders.go`, `internal/*/encoders.go`

**Ejemplo**:
```go
// Decoder: Request → Entity
func createUserRequestToEntitiesUser(req CreateUserRequest) (entities.User, error) {
    password, err := generateSecurePassword(req.Password)
    if err != nil {
        return entities.User{}, err
    }

    return entities.User{
        CI:        req.Document,
        Username:  req.Username,
        FirstName: req.FirstName,
        LastName:  req.LastName,
        Password:  password,
        Roles:     []string{req.Role},
    }, nil
}

// Encoder: Entity → Response
func UserEntityToGetUserResponse(user entities.User) GetUserResponse {
    return GetUserResponse{
        ID:        user.ID,
        Username:  user.Username,
        FirstName: user.FirstName,
        LastName:  user.LastName,
        CreatedAt: user.CreatedAt,
    }
}
```

## Flujo de Datos Completo

### Ejemplo: Crear un Usuario

```go
// 1. HTTP Request → Request Struct (Binding automático de Echo)
func (h *UserEndpointsHandler) CreateUser(c echo.Context) error {
    var req CreateUserRequest
    if err := c.Bind(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }
    
    // Validación de formato
    if err := c.Validate(req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }
    
    // 2. Request → Entity (usando decoder)
    entity, err := createUserRequestToEntitiesUser(req)
    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }
    
    // 3. Service procesa la Entity (lógica de negocio)
    userID, err := h.svc.CreateUser(c.Request().Context(), entity)
    if err != nil {
        h.log.Error("error creating user", zap.Error(err))
        return echo.ErrInternalServerError
    }
    
    // 4. Entity → Response (usando encoder)
    resp := CreateUsersResponse{ID: userID}
    
    // 5. Response → HTTP JSON
    return c.JSON(http.StatusCreated, resp)
}
```

### Flujo Visual

```
Cliente HTTP
    ↓
[Request Struct] ← Validación de formato (tags validate)
    ↓
[Decoder]
    ↓
[Entity] ← Validación de reglas de negocio
    ↓
[Service] ← Lógica de aplicación
    ↓
[Repository] ← Conversión Entity → Model
    ↓
[Database] ← Constraints de DB
    ↓
[Repository] ← Conversión Model → Entity
    ↓
[Service]
    ↓
[Encoder]
    ↓
[Response Struct]
    ↓
Cliente HTTP
```

## Tabla de Referencia Rápida

| Struct Type | Ubicación | Propósito | Tags Comunes | Validación |
|-------------|-----------|-----------|--------------|------------|
| **Request** | `internal/*/request.go` | Recibir datos HTTP | `json`, `form`, `validate`, `path`, `query` | Formato/Sintaxis |
| **Response** | `internal/*/response.go` | Enviar datos HTTP | `json` | N/A |
| **Entity** | `internal/entities/*.go` | Lógica de negocio | Ninguno (puro Go) | Reglas de negocio |
| **Model** | `internal/postgres_repository/models/*.go` | Mapeo DB | `db` | Constraints DB |

## Principios Aplicados

Esta arquitectura sigue los siguientes principios:

- **Clean Architecture**: Separación clara de capas con dependencias unidireccionales
- **SOLID Principles**:
  - **S**RP: Cada struct tiene una única responsabilidad
  - **O**CP: Abierto a extensión (nuevos endpoints), cerrado a modificación (entities estables)
  - **L**SP: Las implementaciones de Repository son intercambiables
  - **I**SP: Interfaces específicas por funcionalidad
  - **D**IP: Las capas superiores dependen de abstracciones (interfaces), no de implementaciones concretas
- **Domain-Driven Design**: Las entities representan el core del dominio
- **Go Best Practices**: Simplicidad, composición sobre herencia, interfaces pequeñas

