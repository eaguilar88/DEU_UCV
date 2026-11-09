DELETE FROM deu.contacts c USING deu.users u
WHERE
  c.owner_type = 'user'
  AND c.contact_type = 'phone'
  AND c.owner_id = u.id
  AND u.ci BETWEEN 103 AND 114;

DELETE FROM deu.user_roles ur USING deu.users u
WHERE
  ur.user_id = u.id
  AND u.ci = 114;

DELETE FROM deu.users
WHERE
  ci = 114;

-- Restaura los datos ficticios originales de 000021
UPDATE deu.users u
SET
  email = v.email,
  first_name = 'Admin',
  last_name = v.last_name,
  address = 'UCV',
  password = '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa',
  updated_at = NOW ()
FROM
  (
    VALUES
      (103, 'coordinador_agronomia@extension.ucv.ve', 'Agronomía'),
      (104, 'coordinador_arquitectura@extension.ucv.ve', 'Arquitectura y Urbanismo'),
      (105, 'coordinador_ciencias@extension.ucv.ve', 'Ciencias'),
      (106, 'coordinador_fases@extension.ucv.ve', 'Ciencias Económicas y Sociales'),
      (107, 'coordinador_ciencias_politicas@extension.ucv.ve', 'Ciencias Jurídicas y Políticas'),
      (108, 'coordinador_veterinaria@extension.ucv.ve', 'Ciencias Veterinarias'),
      (109, 'coordinador_farmacia@extension.ucv.ve', 'Farmacia'),
      (110, 'coordinador_humanidades@extension.ucv.ve', 'Humanidades y Educación'),
      (111, 'coordinador_ingenieria@extension.ucv.ve', 'Ingeniería'),
      (112, 'coordinador_medicina@extension.ucv.ve', 'Medicina'),
      (113, 'coordinador_odontologia@extension.ucv.ve', 'Odontología')
  ) AS v (ci, email, last_name)
WHERE
  u.ci = v.ci;
