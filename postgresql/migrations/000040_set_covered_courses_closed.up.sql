-- A covered course with no cohort yet is 'cerrado', as in the diplomados mock (§2.16).
UPDATE deu.courses
SET estado_gestion = 'cerrado', updated_at = NOW()
WHERE has_documentation = true AND estado_gestion IS NULL;
