-- Insert mock data for user_roles
INSERT INTO
  deu.user_roles (user_id, role_id, domain_type, faculty)
SELECT
  u.id,
  r.id,
  ur.domain_type,
  ur.faculty::deu.faculty_enum
FROM
  (
    VALUES
      ('eaguilar@email.com', 'root', 'all', 'DEU'),
      ('deu_admin1@example.com', 'deu_admin', 'all', 'DEU'),
      ('deu_admin2@example.com', 'deu_admin', 'all', 'DEU'),
      ('coordinador_agronomia@extension.ucv.ve', 'faculty_admin', 'all', 'Agronomía'),
      ('coordinador_arquitectura@extension.ucv.ve', 'faculty_admin', 'all', 'Arquitectura y Urbanismo'),
      ('coordinador_ciencias@extension.ucv.ve', 'faculty_admin', 'all', 'Ciencias'),
      ('coordinador_fases@extension.ucv.ve', 'faculty_admin', 'all', 'Ciencias Económicas y Sociales'),
      ('coordinador_ciencias_politicas@extension.ucv.ve', 'faculty_admin', 'all', 'Ciencias Jurídicas y Políticas'),
      ('coordinador_veterinaria@extension.ucv.ve', 'faculty_admin', 'all', 'Ciencias Veterinarias'),
      ('coordinador_farmacia@extension.ucv.ve', 'faculty_admin', 'all', 'Farmacia'),
      ('coordinador_humanidades@extension.ucv.ve', 'faculty_admin', 'all', 'Humanidades y Educación'),
      ('coordinador_ingenieria@extension.ucv.ve', 'faculty_admin', 'all', 'Ingeniería'),
      ('coordinador_medicina@extension.ucv.ve', 'faculty_admin', 'all', 'Medicina'),
      ('coordinador_odontologia@extension.ucv.ve', 'faculty_admin', 'all', 'Odontología')
  ) AS ur (email, role_name, domain_type, faculty)
  JOIN deu.users u ON u.email = ur.email
  JOIN deu.roles r ON r.name = ur.role_name;
