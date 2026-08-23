DELETE FROM deu.roles
WHERE
  name IN (
    'root',
    'deu_admin',
    'faculty_admin',
    'course_admin',
    'course_manager',
    'group_helper',
    'group_admin',
    'visitante'
  );
