# Anti-patterns y Áreas de Mejora en Course Periods

Este documento identifica los anti-patterns y prácticas que podrían mejorarse en el módulo `course_periods`.

## 1. Anti-patterns Identificados

### 1.1 Error Handling Inconsistente
```go
// Anti-pattern actual
if err != nil {
    h.log.Error("could not decode", zap.Error(err))
    return echo.ErrInternalServerError
}

// En otro lugar del código
if err != nil {
    return echo.ErrUnprocessableEntity
}
```

**Recomendación:**
```go
// Crear errores específicos del dominio
var (
    ErrInvalidPeriod = errors.New("invalid course period")
    ErrPeriodNotFound = errors.New("course period not found")
)

// Usar un helper para mapear errores a códigos HTTP
func mapErrorToHTTPStatus(err error) error {
    switch {
    case errors.Is(err, ErrInvalidPeriod):
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    case errors.Is(err, ErrPeriodNotFound):
        return echo.NewHTTPError(http.StatusNotFound, err.Error())
    default:
        return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
    }
}
```

### 1.2 Comentarios TODO sin Resolver
```go
// TODO: Implement service.go logic
```

**Recomendación:**
- Eliminar los TODOs o convertirlos en issues en el sistema de tracking
- Documentar la razón del TODO si debe mantenerse
- Agregar un identificador de issue si existe uno relacionado

### 1.3 Manejo de Goroutines sin Timeout
```go
// Anti-pattern actual
go func() {
    users, err := s.repo.GetUsersByCoursePeriodID(ctx, periodID)
    if err != nil {
        errorChan <- err
        return
    }
    participantsChan <- users
}()
```

**Recomendación:**
```go
func (s *CoursePeriodService) GetCoursePeriod(ctx context.Context, periodID string) (entities.CoursePeriod, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    period, err := s.repo.GetCoursePeriodByID(ctx, periodID)
    if err != nil {
        return entities.CoursePeriod{}, err
    }

    participantsChan := make(chan []entities.User)
    errorChan := make(chan error)

    go func() {
        users, err := s.repo.GetUsersByCoursePeriodID(ctx, periodID)
        if err != nil {
            errorChan <- err
            return
        }
        participantsChan <- users
    }()

    select {
    case participants := <-participantsChan:
        period.Participants = participants
    case err := <-errorChan:
        return entities.CoursePeriod{}, err
    case <-ctx.Done():
        return entities.CoursePeriod{}, ctx.Err()
    }

    return period, nil
}
```

### 1.4 Validación Insuficiente en Endpoints
```go
// Anti-pattern actual
var req CreateCoursePeriodRequest
if err := c.Bind(&req); err != nil {
    h.log.Error("could not decode", zap.Error(err))
    return echo.ErrBadRequest
}
```

**Recomendación:**
```go
func (h *CoursePeriodEndpointsHandler) CreateCoursePeriod(c echo.Context) error {
    var req CreateCoursePeriodRequest
    if err := c.Bind(&req); err != nil {
        h.log.Error("could not decode request", zap.Error(err))
        return echo.NewHTTPError(http.StatusBadRequest, "invalid request format")
    }

    if err := req.Validate(); err != nil {
        h.log.Error("invalid request", zap.Error(err))
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }

    // ... resto del código
}

func (r *CreateCoursePeriodRequest) Validate() error {
    if r.StartDate.After(r.EndDate) {
        return errors.New("start date must be before end date")
    }
    if r.Capacity < 0 {
        return errors.New("capacity must be non-negative")
    }
    // más validaciones...
    return nil
}
```

### 1.5 Logging sin Contexto Suficiente
```go
// Anti-pattern actual
h.log.Error("could not decode", zap.Error(err))
```

**Recomendación:**
```go
h.log.Error("failed to process course period request",
    zap.Error(err),
    zap.String("period_id", req.ID),
    zap.String("user_id", userID),
    zap.String("action", "create"),
    zap.Any("request_data", req),
)
```

## 2. Mejoras Sugeridas

### 2.1 Agregar Métricas y Tracing
```go
func (h *CoursePeriodEndpointsHandler) GetCoursePeriod(c echo.Context) error {
    metrics.IncRequestCounter("get_course_period")
    defer metrics.ObserveRequestDuration("get_course_period", time.Now())

    span, ctx := tracer.StartSpanFromContext(c.Request().Context(), "GetCoursePeriod")
    defer span.End()

    // ... resto del código
}
```

### 2.2 Implementar Circuit Breaker para Llamadas a Servicios Externos
```go
type CoursePeriodService struct {
    repo       Repository
    log        *zap.Logger
    breaker    *circuitbreaker.CircuitBreaker
}

func (s *CoursePeriodService) GetCoursePeriod(ctx context.Context, periodID string) (entities.CoursePeriod, error) {
    return s.breaker.Execute(func() (entities.CoursePeriod, error) {
        return s.repo.GetCoursePeriodByID(ctx, periodID)
    })
}
```

### 2.3 Mejorar el Manejo de Transacciones
```go
func (s *CoursePeriodService) CreateCoursePeriod(ctx context.Context, period entities.CoursePeriod) (int64, error) {
    tx, err := s.repo.BeginTx(ctx)
    if err != nil {
        return -1, fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()

    id, err := s.repo.CreateCoursePeriodTx(ctx, tx, period)
    if err != nil {
        return -1, fmt.Errorf("failed to create course period: %w", err)
    }

    if err := tx.Commit(); err != nil {
        return -1, fmt.Errorf("failed to commit transaction: %w", err)
    }

    return id, nil
}
```

Estas mejoras ayudarán a:
- Mejorar la robustez del código
- Facilitar el debugging
- Aumentar la mantenibilidad
- Mejorar el monitoreo y observabilidad
- Reducir errores en producción