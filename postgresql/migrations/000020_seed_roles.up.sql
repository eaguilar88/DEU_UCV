-- Insertar datos iniciales
INSERT INTO
  deu.roles (name, description)
VALUES
  (
    'root',
    'Superusuario con acceso total al sistema.'
  ),
  (
    'deu_admin',
    'Administrador con privilegios elevados.'
  ),
  (
    'faculty_admin',
    'Responsable de evaluar solicitudes de avales de cursos y grupos de extension. Puede ser por escuela o de la DEU'
  ),
  (
    'course_admin',
    'Representante de un ente que imparte diplomados.'
  ),
  (
    'course_manager',
    'Encargado de impartir cursos y talleres.'
  ),
  (
    'group_helper',
    'Grupo o departamento de extensión'
  ),
  (
    'group_admin',
    'Grupo o departamento de extensión.'
  ),
  (
    'visitante',
    'Participante interesados en cursos.'
  );
