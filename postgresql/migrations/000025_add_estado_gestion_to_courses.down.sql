ALTER TABLE deu.courses DROP CONSTRAINT IF EXISTS courses_estado_gestion_check;
ALTER TABLE deu.courses DROP COLUMN IF EXISTS estado_gestion;
