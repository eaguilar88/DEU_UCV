# Design Patterns and Code Structures in Course Periods

Este documento identifica los patrones de diseño y estructuras de código utilizadas en el módulo `course_periods`.

## 1. Patrones de Diseño

### 1.1 Interface Segregation (ISP)
```go
type Service interface {
    GetCoursePeriod(ctx context.Context, periodID string) (entities.CoursePeriod, error)
    GetCoursePeriods(ctx context.Context, courseID string, pageScope entities.PageScope) ([]entities.CoursePeriod, entities.PageScope, error)
    // ...
}

type Repository interface {
    GetCoursePeriodByID(ctx context.Context, periodID string) (entities.CoursePeriod, error)
    GetCoursePeriods(ctx context.Context, courseID string, pageScope entities.PageScope) ([]entities.CoursePeriod, entities.PageScope, error)
    // ...
}
```
- Las interfaces están bien definidas y segregadas por responsabilidad
- Cada método tiene un propósito claro y específico

### 1.2 Dependency Injection
```go
func NewCoursePeriodsService(repository Repository, logger *zap.Logger) *CoursePeriodService {
    return &CoursePeriodService{
        repo: repository,
        log:  logger,
    }
}
```
- Las dependencias se inyectan a través del constructor
- Facilita el testing y el cambio de implementaciones

### 1.3 Repository Pattern
```go
type Repository interface {
    // ... métodos de acceso a datos
}
```
- Abstracción de la capa de datos
- Separa la lógica de negocio del acceso a datos
- Permite cambiar la implementación del almacenamiento sin afectar la lógica de negocio

### 1.4 Handler Pattern (Echo Framework)
```go
type CoursePeriodEndpointsHandler struct {
    svc Service
    log *zap.Logger
}
```
- Maneja las peticiones HTTP
- Separa la lógica de routing de la lógica de negocio
- Implementa el patrón de middleware para la autenticación

### 1.5 Concurrent Fan-out Pattern
```go
func (s *CoursePeriodService) GetCoursePeriods(ctx context.Context, courseID string, pageScope entities.PageScope) ([]entities.CoursePeriod, entities.PageScope, error) {
    // ... código de fan-out con goroutines
}
```
- Usa goroutines para procesar múltiples periodos en paralelo
- Implementa manejo de errores concurrentes
- Utiliza sync.WaitGroup para sincronización

## 2. Buenas Prácticas

### 2.1 Error Handling
```go
if err != nil {
    h.log.Error("could not decode", zap.Error(err))
    return echo.ErrBadRequest
}
```
- Logging consistente de errores
- Uso de errores tipados del framework
- Manejo apropiado de errores en cada capa

### 2.2 Context Usage
```go
ctx := c.Request().Context()
```
- Propagación correcta del contexto
- Permite el manejo de cancelaciones y timeouts

### 2.3 Structured Logging
```go
h.log.Error("error getting course period", zap.Error(err))
```
- Uso de logging estructurado con Zap
- Campos consistentes para el logging

### 2.4 Clean Architecture
- Separación clara de responsabilidades:
  - `endpoints.go`: Capa de presentación
  - `service.go`: Lógica de negocio
  - Repository: Acceso a datos
- DTOs separados para requests y responses

### 2.5 Testing
- Mocks generados para interfaces
- Tests unitarios separados por componente
- Pruebas de integración para endpoints