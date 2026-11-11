-- Nuevo administrador de la DEU (rol deu_admin, expuesto a los clientes como "admin").
INSERT INTO
  deu.users (
    ci,
    email,
    first_name,
    last_name,
    date_of_birth,
    gender,
    education,
    address,
    password
  )
VALUES
  (
    117,
    'deu.depgsu@gmail.com',
    'DEPGSU',
    'DEU',
    '2000-01-01',
    'masculino',
    'universitaria_completa',
    'UCV',
    '$2a$10$2HMTqzMI6/KIHtWc6/LRROKhj5MVoCdMQPyxZQjzDfCUS.Myy5YK6'
  );

INSERT INTO
  deu.user_roles (user_id, role_id, domain_type, faculty)
SELECT
  u.id,
  r.id,
  'all',
  'DEU'::deu.faculty_enum
FROM
  deu.users u
  JOIN deu.roles r ON r.name = 'deu_admin'
WHERE
  u.ci = 117;
