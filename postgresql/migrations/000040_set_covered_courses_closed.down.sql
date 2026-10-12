-- Only courses that never had a period can have reached 'cerrado' through the up migration.
UPDATE deu.courses
SET estado_gestion = NULL
WHERE estado_gestion = 'cerrado'
  AND NOT EXISTS (SELECT 1 FROM deu.course_cycles cc WHERE cc.course_id = deu.courses.id);
