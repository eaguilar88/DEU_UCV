ALTER TABLE deu.courses DROP CONSTRAINT IF EXISTS courses_estado_gestion_check;
ALTER TABLE deu.courses ADD CONSTRAINT courses_estado_gestion_check
    CHECK (estado_gestion IS NULL OR estado_gestion IN ('solicitud-cierre'));
