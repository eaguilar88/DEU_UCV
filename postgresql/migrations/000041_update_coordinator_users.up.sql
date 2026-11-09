-- Reemplaza los datos ficticios de los coordinadores de extensión (faculty_admin)
-- sembrados en 000021/000022 por los datos reales, y agrega sus teléfonos en deu.contacts.
-- Se usa ci como clave porque el email cambia.
UPDATE deu.users u
SET
  email = v.email,
  first_name = v.first_name,
  last_name = v.last_name,
  address = v.address,
  password = '$2a$10$dIcRWauLh62VLekI1cQDxesPSHo0ow7iIhIwCKpaMAn8sG5v6Q2um',
  updated_at = NOW ()
FROM
  (
    VALUES
      (103, 'coordinacionextensionfagro@gmail.com', 'Rafael', 'Mejias', 'Coordinación de Extensión, edificio Dependencias Centrales, Maracay'),
      (104, 'cextensionfauucv@gmail.com', 'Marina', 'Fernández', 'Coordinación de Extensión Edificio de la Facultad de Arquitectura y Urbanismo PB'),
      (105, 'yoslyhernandez@gmail.com', 'Yosly', 'Hernández Bieliukas', 'Facultad de Ciencias, piso 1 del Decanato'),
      (106, 'facesextension@gmail.com', 'Maria Isabel', 'Diaz', 'Edificio sede FaCES piso 8, oficina de coordinación de Extensión'),
      (107, 'josuevicenterodriguez@gmail.com', 'Josué Vicente', 'Rodríguez', 'UCV'),
      (108, 'olymarjga@gmail.com', 'Olymar', 'Giménez', 'Coordinación'),
      (109, 'extensionffucv@gmail.com', 'Marisol', 'Gomez', 'Facultad de Farmacia'),
      (110, 'extfheucv@gmail.com', 'Luisa', 'Torrealba', 'Coordinación de Extensión FHyE UCV. Ciudad Universitaria de Caracas. Galpones ubicados al frente de la Facultad de Farmacia'),
      (111, 'itfiucv@gmail.com', 'Milagros', 'Lara Williams', 'Coordinación de Extensión, piso 2, edificio Decanato de la Facultad de Ingeniería.'),
      (112, 'extensionmedicinaucv@gmail.com', 'Patricia', 'Torres', 'Coordinación de Extensión piso 1, Facultad de Medicina.'),
      (113, 'mespinal3@gmail.com', 'Martha', 'Espinoza Alizo', 'Coordinación de Extensión, piso 10. Facultad de Odontología')
  ) AS v (ci, email, first_name, last_name, address)
WHERE
  u.ci = v.ci;

-- Segunda coordinadora de Odontología
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
    114,
    'lolimarmgonzalezz@gmail.com',
    'Lolimar Maria',
    'Gonzalez',
    '2000-01-01',
    'femenino',
    'universitaria_completa',
    'Coordinación de Extensión, piso 10. Facultad de Odontología',
    '$2a$10$dIcRWauLh62VLekI1cQDxesPSHo0ow7iIhIwCKpaMAn8sG5v6Q2um'
  );

INSERT INTO
  deu.user_roles (user_id, role_id, domain_type, faculty)
SELECT
  u.id,
  r.id,
  'all',
  'Odontología'::deu.faculty_enum
FROM
  deu.users u
  JOIN deu.roles r ON r.name = 'faculty_admin'
WHERE
  u.ci = 114;

-- Teléfonos de los coordinadores
INSERT INTO
  deu.contacts (contact_type, contact_value, owner_type, owner_id)
SELECT
  'phone',
  v.phone,
  'user',
  u.id
FROM
  (
    VALUES
      (103, '0412-4363476'),
      (104, '0414-5600386'),
      (105, '0414-3337049'),
      (106, '0414-1839960'),
      (107, '0414-2910150'),
      (108, '0424-3467651'),
      (109, '0414-2477003'),
      (110, '0414-1828252'),
      (111, '0414-1833948'),
      (112, '0412-2203924'),
      (113, '0414-3384955'),
      (114, '0414-2643894')
  ) AS v (ci, phone)
  JOIN deu.users u ON u.ci = v.ci;
