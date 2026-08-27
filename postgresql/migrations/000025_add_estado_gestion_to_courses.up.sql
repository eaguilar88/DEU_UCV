ALTER TABLE deu.courses ADD COLUMN estado_gestion VARCHAR;
ALTER TABLE deu.courses ADD CONSTRAINT courses_estado_gestion_check
    CHECK (estado_gestion IS NULL OR estado_gestion IN ('solicitud-cierre'));
