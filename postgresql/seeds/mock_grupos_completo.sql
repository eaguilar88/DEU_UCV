SET search_path TO deu;

-- ============================================================================
-- 1. CREACIÓN DE LOS 18 USUARIOS REPRESENTANTES DE GRUPOS
-- Todos comparten la contraseña encriptada correspondente a 'nolodire'
-- IDs autogenerados a partir del 15 (ya que tienes del 1 al 14 ocupados)
-- ============================================================================
INSERT INTO deu.users (id, ci, email, first_name, last_name, date_of_birth, gender, education, address, password) VALUES
(15, 201, 'lamun@extension.ucv.ve', 'Representante', 'LAMUN', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa'),
(16, 202, 'spectrum@extension.ucv.ve', 'Representante', 'SPECTRUM', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa'),
(17, 203, 'jamucv@extension.ucv.ve', 'Representante', 'JAM UCV', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa'),
(18, 204, 'cecobio@extension.ucv.ve', 'Representante', 'Cecobio', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa'),
(19, 205, 'cuasar@extension.ucv.ve', 'Representante', 'CUÁSAR', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa'),
(20, 206, 'coralciencias@extension.ucv.ve', 'Representante', 'CORAL CIENCIAS', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa'),
(21, 207, 'grupoera@extension.ucv.ve', 'Representante', 'GRUPO E.R.A.', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa'),
(22, 208, 'geoquimica@extension.ucv.ve', 'Representante', 'Asociación Geoquímica', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa'),
(23, 209, 'biosub@extension.ucv.ve', 'Representante', 'BIOSub', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa'),
(24, 210, 'bailaciencias@extension.ucv.ve', 'Representante', 'BAILA CIENCIAS', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa'),
(25, 211, 'ucvtango@extension.ucv.ve', 'Representante', 'UCV TANGO', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa'),
(26, 212, 'prociencias@extension.ucv.ve', 'Representante', 'PROCIENCIAS', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa'),
(27, 213, 'cineclubciencias@extension.ucv.ve', 'Representante', 'CINECLUB CIENCIAS', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa'),
(28, 214, 'teatrobuho@extension.ucv.ve', 'Representante', 'TEATRO BÚHO', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa'),
(29, 215, 'barriomat@extension.ucv.ve', 'Representante', 'BARRIO MATEMÁTICO', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa'),
(30, 216, 'physis@extension.ucv.ve', 'Representante', 'PHYSIS', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa'),
(31, 217, 'concienciagaitera@extension.ucv.ve', 'Representante', 'CONCIENCIA GAITERA', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa'),
(32, 218, 'cienciasritmo@extension.ucv.ve', 'Representante', 'CIENCIAS ES RITMO', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa'),
(33, 219, 'agronomia_verde@extension.ucv.ve', 'Representante', 'AGRO VERDE', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa'),
(34, 220, 'fau_diseno@extension.ucv.ve', 'Representante', 'FAU DISEÑO', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa'),
(35, 221, 'faces_emprende@extension.ucv.ve', 'Representante', 'FACES EMPRENDE', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa'),
(36, 222, 'derecho_humano@extension.ucv.ve', 'Representante', 'DERECHO A TIEMPO', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa'),
(37, 223, 'vet_salud@extension.ucv.ve', 'Representante', 'VET SALUD', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa'),
(38, 224, 'humanidades_letras@extension.ucv.ve', 'Representante', 'LETRAS VIVAS', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa'),
(39, 225, 'ingenieria_robotica@extension.ucv.ve', 'Representante', 'ROBÓTICA UCV', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa'),
(40, 226, 'odontologia_comunidad@extension.ucv.ve', 'Representante', 'SONRISA UCV', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa'),
(41, 227, 'deu_voluntariado@extension.ucv.ve', 'Representante', 'VOLUNTARIADO DEU', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa'),
(42, 2288888, 'bobo@amonra.com', 'Juan', 'Jo', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa'),
(43, 229, 'bioingenieria@extension.ucv.ve', 'Representante', 'BIOINGENIERÍA Y SALUD', '2000-01-01', 'NA', 'universitaria_completa', 'UCV', '$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa');

-- Sincronizar el ID secuencial de la tabla users
SELECT setval('deu.users_id_seq', 43, true);

-- ============================================================================
-- 2. ASIGNACIÓN DEL ROL 'group_admin' (ID 7) A LOS NUEVOS USUARIOS
-- ============================================================================
INSERT INTO deu.user_roles (user_id, role_id, domain_type, faculty) VALUES
(15, 7, 'group', 'DEU'),
(16, 7, 'group', 'Ciencias'),
(17, 7, 'group', 'Ciencias'),
(18, 7, 'group', 'Ciencias'),
(19, 7, 'group', 'Ciencias'),
(20, 7, 'group', 'Ciencias'),
(21, 7, 'group', 'Ciencias'),
(22, 7, 'group', 'Ciencias'),
(23, 7, 'group', 'Ciencias'),
(24, 7, 'group', 'Ciencias'),
(25, 7, 'group', 'Ciencias'),
(26, 7, 'group', 'Farmacia'),
(27, 7, 'group', 'Medicina'),
(28, 7, 'group', 'Ciencias'),
(29, 7, 'group', 'Ciencias'),
(30, 7, 'group', 'Ciencias'),
(31, 7, 'group', 'Ciencias'),
(32, 7, 'group', 'Ciencias'),
(33, 7, 'group', 'Agronomía'),
(34, 7, 'group', 'Arquitectura y Urbanismo'),
(35, 7, 'group', 'Ciencias Económicas y Sociales'),
(36, 7, 'group', 'Ciencias Jurídicas y Políticas'),
(37, 7, 'group', 'Ciencias Veterinarias'),
(38, 7, 'group', 'Humanidades y Educación'),
(39, 7, 'group', 'Ingeniería'),
(40, 7, 'group', 'Odontología'),
(41, 7, 'group', 'DEU'),
(42, 8, 'group', 'Ciencias'),
(43, 7, 'group', 'DEU');

-- ============================================================================
-- 3. INSERCIÓN DE LOS 26 GRUPOS DE EXTENSIÓN (Enlazados a sus usuarios)
-- Mapeo estricto del ENUM para faculty (Vacío o no mapeado se asigna a 'DEU')
-- ============================================================================
-- GEX columna 

INSERT INTO deu.extension_groups (id, user_id, name, description, is_multidisciplinary, faculty, foundation, objective, code, group_director, type, location, is_active) VALUES
(1, 15, 'LAMUN', 'Grupo de Modelaje de Naciones Unidas de la UCV.', TRUE, ARRAY['DEU']::faculty_enum[], '2012-03-15', 'Objetivo institucional LAMUN.', 'GEX-LAMUN-01', 'Director LAMUN', ARRAY['MULTIDISCIPLINARIO']::VARCHAR[], 'Sala de Extensión Central', TRUE),
(2, 16, 'SPECTRUM', 'Grupo enfocado en la divulgación científica y tecnológica.', FALSE, ARRAY['Ciencias']::faculty_enum[], '2018-06-20', 'Objetivo de divulgación tecnológica.', 'GEX-SPEC-02', 'Director SPECTRUM', ARRAY['ACADÉMICO']::VARCHAR[], 'Facultad de Ciencias', TRUE),
(3, 17, 'JAM UCV', 'Agrupación musical e intercambio de expresión de ritmos.', FALSE, ARRAY['Ciencias']::faculty_enum[], '2015-11-10', 'Promover cultura musical.', 'GEX-JAM-03', 'Director JAM', ARRAY['CULTURAL']::VARCHAR[], 'Pasillos de Ciencias', TRUE),
(4, 18, 'Cecobio', 'Centro de conservación de la biodiversidad biológica.', FALSE, ARRAY['Ciencias']::faculty_enum[], '2010-04-05', 'Preservación de especies locales.', 'GEX-CECO-04', 'Director Cecobio', ARRAY['CIENTÍFICO']::VARCHAR[], 'Laboratorio de Biología', TRUE),
(5, 19, 'CUÁSAR', 'Grupo de astronomía y ciencias del espacio.', FALSE, ARRAY['Ciencias']::faculty_enum[], '2016-09-12', 'Observación del espacio profundo.', 'GEX-CUAS-05', 'Director CUÁSAR', ARRAY['ACADÉMICO']::VARCHAR[], 'Observatorio de Ciencias', TRUE),
(6, 20, 'CORAL FACULTAD DE CIENCIAS', 'Coro polifónico institucional de la facultad.', FALSE, ARRAY['Ciencias']::faculty_enum[], '2005-01-20', 'Canto coral universitario.', 'GEX-CORO-06', 'Director Coral Ciencias', ARRAY['CULTURAL']::VARCHAR[], 'Auditorio de Ciencias', TRUE),
(7, 21, 'GRUPO E.R.A.', 'Equipo de respuesta ambiental y reforestación.', FALSE, ARRAY['Ciencias']::faculty_enum[], '2019-02-14', 'Reforestar áreas del campus.', 'GEX-GERA-07', 'Director ERA', ARRAY['ECOLÓGICO']::VARCHAR[], 'Áreas Verdes Ciencias', TRUE),
(8, 22, 'Asociación Geoquímica Ciencias UCV', 'Promoción e investigación de procesos geoquímicos.', FALSE, ARRAY['Ciencias']::faculty_enum[], '2014-08-30', 'Estudio químico de suelos.', 'GEX-GEOQ-08', 'Director Geoquímica', ARRAY['ACADÉMICO']::VARCHAR[], 'Instituto de Ciencias de la Tierra', TRUE),
(9, 23, 'BIOSub', 'Grupo de exploraciones submarinas y biología marina.', FALSE, ARRAY['Ciencias']::faculty_enum[], '2011-05-18', 'Exploración subacuática.', 'GEX-BSUB-09', 'Director BIOSub', ARRAY['DEPORTIVO', 'ACADÉMICO']::VARCHAR[], 'Piscina UCV / Laboratorio', TRUE),
(10, 24, 'BAILA CIENCIAS', 'Agrupación de danza tradicional y moderna.', FALSE, ARRAY['Ciencias']::faculty_enum[], '2017-10-12', 'Danza folclórica.', 'GEX-BAIL-10', 'Director Baila Ciencias', ARRAY['CULTURAL']::VARCHAR[], 'Plaza Cubierta de Ciencias', TRUE),
(11, 25, 'UCV TANGO', 'Taller y grupo de proyección de Tango.', FALSE, ARRAY['Ciencias']::faculty_enum[], '2013-07-07', 'Difusión del tango.', 'GEX-TANG-11', 'Director UCV Tango', ARRAY['CULTURAL']::VARCHAR[], 'Sala de Usos Múltiples', TRUE),
(12, 26, 'PROCIENCIAS', 'Asociación para el desarrollo de proyectos en Farmacia y Química.', FALSE, ARRAY['Farmacia']::faculty_enum[], '2015-03-22', 'Innovación farmacológica.', 'GEX-PROC-12', 'Director PROCIENCIAS', ARRAY['CIENTÍFICO']::VARCHAR[], 'Facultad de Farmacia', TRUE),
(13, 27, 'CINECLUB CIENCIAS', 'Espacio para el debate y proyección cinematográfica.', FALSE, ARRAY['Medicina']::faculty_enum[], '2008-11-05', 'Cine foro formativo.', 'GEX-CINE-13', 'Director Cineclub', ARRAY['CULTURAL']::VARCHAR[], 'Auditorio de Medicina', TRUE),
(14, 28, 'TEATRO Y CULTURA BÚHO DE CIENCIAS', 'Grupo estable de artes escénicas teatrales.', FALSE, ARRAY['Ciencias']::faculty_enum[], '2006-09-15', 'Montajes teatrales.', 'GEX-BUHO-14', 'Director Teatro Búho', ARRAY['CULTURAL']::VARCHAR[], 'Teatro de Ciencias', TRUE),
(15, 29, 'BARRIO MATEMÁTICO', 'Iniciativa de extensión para la enseñanza lúdica de la matemática.', FALSE, ARRAY['Ciencias']::faculty_enum[], '2020-01-10', 'Enseñanza interactiva.', 'GEX-BMAT-15', 'Director Barrio Mat.', ARRAY['EDUCATIVO']::VARCHAR[], 'Escuela de Matemáticas', TRUE),
(16, 30, 'PHYSIS', 'Sociedad de estudiantes de física aplicada y teórica.', FALSE, ARRAY['Ciencias']::faculty_enum[], '2018-04-12', 'Física experimental.', 'GEX-PHYS-16', 'Director PHYSIS', ARRAY['ACADÉMICO']::VARCHAR[], 'Escuela de Física', TRUE),
(17, 31, 'CONCIENCIA GAITERA', 'Agrupación musical de gaita zuliana de la facultad.', FALSE, ARRAY['Ciencias']::faculty_enum[], '2014-11-18', 'Preservación de la gaita.', 'GEX-GAIT-17', 'Director C. Gaitera', ARRAY['CULTURAL']::VARCHAR[], 'Estudio de Música Ciencias', TRUE),
(18, 32, 'CIENCIAS ES RITMO', 'Taller abierto de percusión y ritmos latinos.', FALSE, ARRAY['Ciencias']::faculty_enum[], '2019-06-01', 'Percusión afrovenezolana.', 'GEX-RITM-18', 'Director Ciencias es Ritmo', ARRAY['CULTURAL']::VARCHAR[], 'Sótano de Ciencias', TRUE),
(19, 33, 'AgroVerde UCV', 'Taller sustentable y cultivos urbanos campus Maracay.', FALSE, ARRAY['Agronomía']::faculty_enum[], '2016-02-10', 'Promover soberanía alimentaria local.', 'GEX-AGRO-19', 'Director AgroVerde', ARRAY['ECOLÓGICO']::VARCHAR[], 'Campus Maracay Agronomía', TRUE),
(20, 34, 'Taller de Diseño Urbano FAU', 'Colectivo de intervenciones arquitectónicas comunitarias.', FALSE, ARRAY['Arquitectura y Urbanismo']::faculty_enum[], '2015-08-25', 'Mejoramiento de espacios urbanos.', 'GEX-FAUD-20', 'Director FAU Diseño', ARRAY['ACADÉMICO']::VARCHAR[], 'Galpón FAU', TRUE),
(21, 35, 'FaCES Emprende', 'Incubadora de proyectos y educación financiera comunitaria.', FALSE, ARRAY['Ciencias Económicas y Sociales']::faculty_enum[], '2021-03-14', 'Fomentar la cultura emprendedora.', 'GEX-FCSE-21', 'Director FaCES Emprende', ARRAY['EDUCATIVO']::VARCHAR[], 'Edificio Trasbordo FaCES', TRUE),
(22, 36, 'Clínica Jurídica Gratuita', 'Asesoría legal para comunidades vulnerables.', FALSE, ARRAY['Ciencias Jurídicas y Políticas']::faculty_enum[], '2009-07-01', 'Atención jurídica accesible.', 'GEX-CJUR-22', 'Director Clínica Jurídica', ARRAY['SOCIAL']::VARCHAR[], 'Planta Baja CJP', TRUE),
(23, 37, 'Veterinarios en Acción', 'Atención zoonótica y jornadas de vacunación.', FALSE, ARRAY['Ciencias Veterinarias']::faculty_enum[], '2017-05-12', 'Bienestar animal público.', 'GEX-VETA-23', 'Director Vet en Acción', ARRAY['SALUD', 'SOCIAL']::VARCHAR[], 'Hospital Veterinario Maracay', TRUE),
(24, 38, 'Letras Vivas', 'Promoción de la lectura y talleres de escritura creativa.', FALSE, ARRAY['Humanidades y Educación']::faculty_enum[], '2013-10-30', 'Fomento de la literatura.', 'GEX-LETR-24', 'Director Letras Vivas', ARRAY['CULTURAL']::VARCHAR[], 'Biblioteca Central Humanidades', TRUE),
(25, 39, 'Robótica e Innovación UCV', 'Desarrollo mecatrónico y automatización social.', FALSE, ARRAY['Ingeniería']::faculty_enum[], '2018-11-11', 'Divulgación de mecatrónica.', 'GEX-ROBO-25', 'Director Robótica UCV', ARRAY['TECNOLÓGICO']::VARCHAR[], 'Escuela de Ingeniería Eléctrica', TRUE),
(26, 40, 'Sonrisas UCV', 'Jornadas comunitarias de higiene y salud bucal.', FALSE, ARRAY['Odontología']::faculty_enum[], '2011-09-09', 'Salud bucal preventiva.', 'GEX-SONR-26', 'Director Sonrisas UCV', ARRAY['SALUD', 'SOCIAL']::VARCHAR[], 'Clínica de Odontología', TRUE),
(27, 43, 'Iniciativa de Bioingeniería y Salud Comunitaria', 'Colectivo interfacultades dedicado al desarrollo de prótesis biomédicas de bajo costo y tecnología médica asistencial.', TRUE, ARRAY['DEU', 'Ingeniería', 'Ciencias', 'Medicina']::faculty_enum[], '2021-05-10', 'Desarrollar soluciones tecnológicas accesibles para el sector salud comunitario.', 'GEX-BIOS-27', 'Director Bioingeniería', ARRAY['MULTIDISCIPLINARIO']::VARCHAR[], 'Laboratorio de Prototipado Interdisciplinario', TRUE);

SELECT setval('deu.extension_groups_id_seq', 27, true);

-- ============================================================================
-- 4. CONTACTOS COMPLETOS PARA CADA GRUPO (Email y Teléfono)
-- owner_type: 'group', owner_id mapped de 1 a 18
-- ============================================================================
INSERT INTO contacts (contact_type, contact_value, owner_type, owner_id) VALUES
('email', 'contacto.lamun@email.com', 'group', 1), ('phone', '0412-1111111', 'group', 1),
('email', 'contacto.spectrum@email.com', 'group', 2), ('phone', '0412-2222222', 'group', 2),
('email', 'contacto.jam@email.com', 'group', 3), ('phone', '0412-3333333', 'group', 3),
('email', 'contacto.cecobio@email.com', 'group', 4), ('phone', '0412-4444444', 'group', 4),
('email', 'contacto.cuasar@email.com', 'group', 5), ('phone', '0412-5555555', 'group', 5),
('email', 'contacto.coral@email.com', 'group', 6), ('phone', '0412-6666666', 'group', 6),
('email', 'contacto.era@email.com', 'group', 7), ('phone', '0412-7777777', 'group', 7),
('email', 'contacto.geoquimica@email.com', 'group', 8), ('phone', '0412-8888888', 'group', 8),
('email', 'contacto.biosub@email.com', 'group', 9), ('phone', '0412-9999999', 'group', 9),
('email', 'contacto.baila@email.com', 'group', 10), ('phone', '0414-1111111', 'group', 10),
('email', 'contacto.tango@email.com', 'group', 11), ('phone', '0414-2222222', 'group', 11),
('email', 'contacto.prociencias@email.com', 'group', 12), ('phone', '0414-3333333', 'group', 12),
('email', 'contacto.cineclub@email.com', 'group', 13), ('phone', '0414-4444444', 'group', 13),
('email', 'contacto.buho@email.com', 'group', 14), ('phone', '0414-5555555', 'group', 14),
('email', 'contacto.barriomat@email.com', 'group', 15), ('phone', '0414-6666666', 'group', 15),
('email', 'contacto.physis@email.com', 'group', 16), ('phone', '0414-7777777', 'group', 16),
('email', 'contacto.gaitera@email.com', 'group', 17), ('phone', '0414-8888888', 'group', 17),
('email', 'contacto.ritmo@email.com', 'group', 18), ('phone', '0414-9999999', 'group', 18),
('email', 'contacto.agroverde@email.com', 'group', 19), ('phone', '0416-1111111', 'group', 19),
('email', 'contacto.faudiseno@email.com', 'group', 20), ('phone', '0416-2222222', 'group', 20),
('email', 'contacto.facesemprende@email.com', 'group', 21), ('phone', '0416-3333333', 'group', 21),
('email', 'contacto.cjuridica@email.com', 'group', 22), ('phone', '0416-4444444', 'group', 22),
('email', 'contacto.vetaction@email.com', 'group', 23), ('phone', '0416-5555555', 'group', 23),
('email', 'contacto.letrasvivas@email.com', 'group', 24), ('phone', '0416-6666666', 'group', 24),
('email', 'contacto.robotica@email.com', 'group', 25), ('phone', '0416-7777777', 'group', 25),
('email', 'contacto.sonrisas@email.com', 'group', 26), ('phone', '0416-8888888', 'group', 26),
('email', 'contacto.bioingenieria@email.com', 'group', 27), ('phone', '0412-0002727', 'group', 27);

-- ============================================================================
-- 5. SOLICITUD DE AVAL EN group_auth_requests PARA CADA GRUPO
-- Relacionado a deu_admin1 (reviewer_id = 2)
-- ============================================================================
INSERT INTO deu.group_auth_requests (group_id, status, faculty, reviewer_id, reviewed_at, comments) VALUES
(1, 'approved', 'DEU', 2, NOW(), 'Cumple con los estatutos de extensión soberana.'),
(2, 'approved', 'Ciencias', 2, NOW(), 'Documentación técnica de divulgación validada.'),
(3, 'approved', 'Ciencias', 2, NOW(), 'Proyecto cultural aprobado.'),
(4, 'approved', 'Ciencias', 2, NOW(), 'Aval científico otorgado.'),
(5, 'approved', 'Ciencias', 2, NOW(), 'Plan operativo anual astronómico correcto.'),
(6, 'approved', 'Ciencias', 2, NOW(), 'Visto bueno al coro polifónico.'),
(7, 'approved', 'Ciencias', 2, NOW(), 'Iniciativa ecológica avalada.'),
(8, 'approved', 'Ciencias', 2, NOW(), 'Líneas de investigación en geoquímica aprobadas.'),
(9, 'approved', 'Ciencias', 2, NOW(), 'Aval deportivo y científico unificado.'),
(10, 'approved', 'Ciencias', 2, NOW(), 'Registro dancístico tradicional completado.'),
(11, 'approved', 'Ciencias', 2, NOW(), 'Aprobado para dictar talleres libres.'),
(12, 'approved', 'Farmacia', 2, NOW(), 'Avalado por la división de extensión de Farmacia.'),
(13, 'approved', 'Medicina', 2, NOW(), 'Cineforo aprobado por la junta de Medicina.'),
(14, 'approved', 'Ciencias', 2, NOW(), 'Agrupación teatral estable registrada.'),
(15, 'approved', 'Ciencias', 2, NOW(), 'Aprobada la propuesta lúdico-educativa de matemáticas.'),
(16, 'approved', 'Ciencias', 2, NOW(), 'Sociedad física estudiantil oficialmente avalada.'),
(17, 'approved', 'Ciencias', 2, NOW(), 'Expediente cultural gaitero aprobado.'),
(18, 'approved', 'Ciencias', 2, NOW(), 'Taller rítmico aprobado satisfactoriamente.'),
(19, 'approved', 'Agronomía', 4, NOW(), 'Proyecto agronómico comunitario validado.'),
(20, 'approved', 'Arquitectura y Urbanismo', 5, NOW(), 'Propuesta de intervención urbana aprobada.'),
(21, 'approved', 'Ciencias Económicas y Sociales', 7, NOW(), 'Plan de incubadora social de empresas verificado.'),
(22, 'approved', 'Ciencias Jurídicas y Políticas', 8, NOW(), 'Asesoría jurídica universitaria autorizada.'),
(23, 'approved', 'Ciencias Veterinarias', 9, NOW(), 'Jornadas profilácticas asistenciales aprobadas.'),
(24, 'approved', 'Humanidades y Educación', 11, NOW(), 'Plan de promoción lectora validado.'),
(25, 'approved', 'Ingeniería', 12, NOW(), 'Proyecto tecnológico de automatización aprobado.'),
(26, 'approved', 'Odontología', 14, NOW(), 'Aval asistencial odontológico concedido.'),
(27, 'approved', 'DEU', 2, NOW(), 'Grupo multidisciplinario avalado centralmente por la DEU.');

-- ============================================================================
-- 6. SOLICITUDES DE RECURSOS EN group_resource_requests (Pruebas)
-- Varios registros distribuidos en algunos grupos estratégicos
-- ============================================================================
INSERT INTO deu.group_resource_requests (group_id, type, content, status) VALUES
(1, 'Espacio', 'Solicitud de Aula Magna para simulación de Asamblea General.', 'under_review'),
(1, 'Financiamiento', 'Presupuesto para viáticos de delegación interuniversitaria.', 'rejected'),
(2, 'Equipos', 'Préstamo de video beam y sonido para ponencia.', 'approved'),
(3, 'Espacio', 'Permiso para uso del anfiteatro los días viernes por ensayo.', 'approved'),
(4, 'Materiales', 'Insumos de papelería y reactivos básicos de campo.', 'under_review'),
(5, 'Equipos', 'Préstamo de telescopio portátil instituciona.', 'approved'),
(6, 'Materiales', 'Partituras impresas y atriles suplementarios.', 'approved'),
(7, 'Materiales', 'Herramientas de jardinería y bolsas plásticas para composta.', 'approved'),
(8, 'Materiales', 'Frascos de recolección de muestras de suelo.', 'under_review'),
(9, 'Equipos', 'Tanques de buceo para prácticas controladas.', 'approved'),
(10, 'Espacio', 'Asignación del escenario de la plaza cubierta.', 'approved'),
(11, 'Espacio', 'Uso del salón de espejos 2 veces por semana.', 'under_review'),
(12, 'Materiales', 'Reactivos de química analítica para desarrollo de muestras.', 'approved'),
(13, 'Equipos', 'Proyector HD para sala de minicine.', 'approved'),
(14, 'Espacio', 'Uso del auditorio los días martes para montaje de obra.', 'under_review'),
(15, 'Materiales', 'Juegos didácticos en madera para talleres matemáticos.', 'approved'),
(16, 'Equipos', 'Multímetros y osciloscopios para demostración.', 'approved'),
(17, 'Equipos', 'Mantenimiento y afinación de furrucos y tamboras.', 'approved'),
(18, 'Materiales', 'Reparación de instrumentos de percusión menor.', 'under_review'),
(19, 'Materiales', 'Semillas certificadas y compost orgánico.', 'approved'),
(20, 'Materiales', 'Planos impresos en plóter y escuadras topográficas.', 'approved'),
(21, 'Espacio', 'Reserva de aula interactiva para talleres de finanzas.', 'approved'),
(22, 'Materiales', 'Insumos de papelería para carpetas de casos.', 'approved'),
(23, 'Materiales', 'Vacunas antirrábicas e insumos de sutura.', 'approved'),
(24, 'Materiales', 'Impresión de folletos y libros para biblioteca móvil.', 'approved'),
(25, 'Equipos', 'Tarjetas Arduino, Raspberry Pi y sensores mecánicos.', 'approved'),
(26, 'Materiales', 'Kits profilácticos bucales (cepillos y pasta dental).', 'approved'),
(27, 'Equipos', 'Solicitud de cortadora láser y scanner 3D para fabricación de prótesis.', 'under_review');

-- ============================================================================
-- 7. MIEMBROS DE GRUPOS
-- ============================================================================
INSERT INTO deu.group_members (group_id, name, ci, phone, email, coordination, year, faculty, school, document, is_active) VALUES
-- Grupo 1 (LAMUN)
(1, 'Pedro Pascal', 12545623, '0412-0000001', 'pedro@mail.com', 'Elysium', '5', 'DEU', 'Estudios Generales', 'V-12545623.pdf', TRUE),
(1, 'Ana Armas', 23111222, '0412-0000002', 'ana@mail.com', 'Debate', '4', 'DEU', 'Estudios Generales', 'V-23111222.pdf', TRUE),
(1, 'Carlos Perez', 24111333, '0412-0000003', 'carlos@mail.com', 'Logística', '3', 'DEU', 'Estudios Generales', 'V-24111333.pdf', TRUE),
(1, 'Lucía Gómez', 25111444, '0412-0000004', 'lucia@mail.com', 'Finanzas', '2', 'DEU', 'Estudios Generales', 'V-25111444.pdf', TRUE),
(1, 'Mateo Díaz', 26111555, '0412-0000005', 'mateo@mail.com', 'Prensa', '1', 'DEU', 'Estudios Generales', 'V-26111555.pdf', TRUE),

-- Grupo 2 (SPECTRUM)
(2, 'Roberto Fernández', 20111111, '0414-0000001', 'roberto@mail.com', 'Divulgación', '3', 'Ciencias', 'Física', 'V-20111111.pdf', TRUE),
(2, 'Valeria Silva', 20222222, '0414-0000002', 'valeria@mail.com', 'Prensa', '2', 'Ciencias', 'Química', 'V-20222222.pdf', TRUE),
(2, 'Marcos Ruiz', 20333333, '0414-0000003', 'marcos@mail.com', 'Talleres', '4', 'Ciencias', 'Biología', 'V-20333333.pdf', TRUE),
(2, 'Sofia Torres', 20444444, '0414-0000004', 'sofia@mail.com', 'Edición', '1', 'Ciencias', 'Computación', 'V-20444444.pdf', TRUE),
(2, 'Diego Rivas', 20555555, '0414-0000005', 'diego@mail.com', 'Redes', '2', 'Ciencias', 'Matemáticas', 'V-20555555.pdf', TRUE),

-- Grupo 3 (JAM UCV)
(3, 'Gabriel Sosa', 21111001, '0416-0000001', 'gabriel@mail.com', 'Percusión', '3', 'Ciencias', 'Computación', 'V-21111001.pdf', TRUE),
(3, 'Laura Méndez', 21111002, '0416-0000002', 'laura@mail.com', 'Vientos', '2', 'Ciencias', 'Biología', 'V-21111002.pdf', TRUE),
(3, 'Daniel Ortega', 21111003, '0416-0000003', 'daniel@mail.com', 'Cuerdas', '4', 'Ciencias', 'Física', 'V-21111003.pdf', TRUE),
(3, 'Camila Núñez', 21111004, '0416-0000004', 'camila@mail.com', 'Voz', '1', 'Ciencias', 'Química', 'V-21111004.pdf', TRUE),
(3, 'Andrés Bello', 21111005, '0416-0000005', 'andres@mail.com', 'Teclados', '5', 'Ciencias', 'Matemáticas', 'V-21111005.pdf', TRUE),

-- Grupo 4 (Cecobio)
(4, 'Elena Ramos', 22111001, '0412-1000001', 'elena@mail.com', 'Botánica', '4', 'Ciencias', 'Biología', 'V-22111001.pdf', TRUE),
(4, 'Jorge Blanco', 22111002, '0412-1000002', 'jorge@mail.com', 'Zoología', '3', 'Ciencias', 'Biología', 'V-22111002.pdf', TRUE),
(4, 'Patricia Vega', 22111003, '0412-1000003', 'patricia@mail.com', 'Ecología', '2', 'Ciencias', 'Biología', 'V-22111003.pdf', TRUE),
(4, 'Santi Castro', 22111004, '0412-1000004', 'santi@mail.com', 'Micología', '1', 'Ciencias', 'Biología', 'V-22111004.pdf', TRUE),
(4, 'Beatriz Peña', 22111005, '0412-1000005', 'beatriz@mail.com', 'Conservación', '5', 'Ciencias', 'Biología', 'V-22111005.pdf', TRUE),

-- Generación masiva de miembros para los grupos del 5 al 26 (5 por grupo)
(5, 'M1 Grupo 5', 300501, '0412000', 'm51@test.com', 'Coord', '1', 'Ciencias', 'Física', 'doc.pdf', TRUE),
(5, 'M2 Grupo 5', 300502, '0412000', 'm52@test.com', 'Coord', '2', 'Ciencias', 'Física', 'doc.pdf', TRUE),
(5, 'M3 Grupo 5', 300503, '0412000', 'm53@test.com', 'Coord', '3', 'Ciencias', 'Física', 'doc.pdf', TRUE),
(5, 'M4 Grupo 5', 300504, '0412000', 'm54@test.com', 'Coord', '4', 'Ciencias', 'Física', 'doc.pdf', TRUE),
(5, 'M5 Grupo 5', 300505, '0412000', 'm55@test.com', 'Coord', '5', 'Ciencias', 'Física', 'doc.pdf', TRUE),

(6, 'M1 Grupo 6', 300601, '0412000', 'm61@test.com', 'Soprano', '1', 'Ciencias', 'Computación', 'doc.pdf', TRUE),
(6, 'M2 Grupo 6', 300602, '0412000', 'm62@test.com', 'Alto', '2', 'Ciencias', 'Computación', 'doc.pdf', TRUE),
(6, 'M3 Grupo 6', 300603, '0412000', 'm63@test.com', 'Tenor', '3', 'Ciencias', 'Computación', 'doc.pdf', TRUE),
(6, 'M4 Grupo 6', 300604, '0412000', 'm64@test.com', 'Bajo', '4', 'Ciencias', 'Computación', 'doc.pdf', TRUE),
(6, 'M5 Grupo 6', 300605, '0412000', 'm65@test.com', 'Director Adjunto', '5', 'Ciencias', 'Computación', 'doc.pdf', TRUE),

(7, 'M1 Grupo 7', 300701, '0412000', 'm71@test.com', 'Campo', '1', 'Ciencias', 'Geoquímica', 'doc.pdf', TRUE),
(7, 'M2 Grupo 7', 300702, '0412000', 'm72@test.com', 'Campo', '2', 'Ciencias', 'Geoquímica', 'doc.pdf', TRUE),
(7, 'M3 Grupo 7', 300703, '0412000', 'm73@test.com', 'Logística', '3', 'Ciencias', 'Geoquímica', 'doc.pdf', TRUE),
(7, 'M4 Grupo 7', 300704, '0412000', 'm74@test.com', 'Vivero', '4', 'Ciencias', 'Geoquímica', 'doc.pdf', TRUE),
(7, 'M5 Grupo 7', 300705, '0412000', 'm75@test.com', 'Reforestación', '5', 'Ciencias', 'Geoquímica', 'doc.pdf', TRUE),

(8, 'M1 Grupo 8', 300801, '0412000', 'm81@test.com', 'Lab', '1', 'Ciencias', 'Geoquímica', 'doc.pdf', TRUE),
(8, 'M2 Grupo 8', 300802, '0412000', 'm82@test.com', 'Lab', '2', 'Ciencias', 'Geoquímica', 'doc.pdf', TRUE),
(8, 'M3 Grupo 8', 300803, '0412000', 'm83@test.com', 'Análisis', '3', 'Ciencias', 'Geoquímica', 'doc.pdf', TRUE),
(8, 'M4 Grupo 8', 300804, '0412000', 'm84@test.com', 'Muestreo', '4', 'Ciencias', 'Geoquímica', 'doc.pdf', TRUE),
(8, 'M5 Grupo 8', 300805, '0412000', 'm85@test.com', 'Publicaciones', '5', 'Ciencias', 'Geoquímica', 'doc.pdf', TRUE),

(9, 'M1 Grupo 9', 300901, '0412000', 'm91@test.com', 'Buceo', '1', 'Ciencias', 'Biología', 'doc.pdf', TRUE),
(9, 'M2 Grupo 9', 300902, '0412000', 'm92@test.com', 'Buceo', '2', 'Ciencias', 'Biología', 'doc.pdf', TRUE),
(9, 'M3 Grupo 9', 300903, '0412000', 'm93@test.com', 'Equipos', '3', 'Ciencias', 'Biología', 'doc.pdf', TRUE),
(9, 'M4 Grupo 9', 300904, '0412000', 'm94@test.com', 'Seguridad', '4', 'Ciencias', 'Biología', 'doc.pdf', TRUE),
(9, 'M5 Grupo 9', 300905, '0412000', 'm95@test.com', 'Fotografía Sub', '5', 'Ciencias', 'Biología', 'doc.pdf', TRUE),

(10, 'M1 Grupo 10', 301001, '0412000', 'm101@test.com', 'Danza', '1', 'Ciencias', 'Química', 'doc.pdf', TRUE),
(10, 'M2 Grupo 10', 301002, '0412000', 'm102@test.com', 'Danza', '2', 'Ciencias', 'Química', 'doc.pdf', TRUE),
(10, 'M3 Grupo 10', 301003, '0412000', 'm103@test.com', 'Vestuario', '3', 'Ciencias', 'Química', 'doc.pdf', TRUE),
(10, 'M4 Grupo 10', 301004, '0412000', 'm104@test.com', 'Música', '4', 'Ciencias', 'Química', 'doc.pdf', TRUE),
(10, 'M5 Grupo 10', 301005, '0412000', 'm105@test.com', 'Coreografía', '5', 'Ciencias', 'Química', 'doc.pdf', TRUE),

(11, 'M1 Grupo 11', 301101, '0412000', 'm111@test.com', 'Tango', '1', 'Ciencias', 'Matemáticas', 'doc.pdf', TRUE),
(11, 'M2 Grupo 11', 301102, '0412000', 'm112@test.com', 'Tango', '2', 'Ciencias', 'Matemáticas', 'doc.pdf', TRUE),
(11, 'M3 Grupo 11', 301103, '0412000', 'm113@test.com', 'Eventos', '3', 'Ciencias', 'Matemáticas', 'doc.pdf', TRUE),
(11, 'M4 Grupo 11', 301104, '0412000', 'm114@test.com', 'Musicalización', '4', 'Ciencias', 'Matemáticas', 'doc.pdf', TRUE),
(11, 'M5 Grupo 11', 301105, '0412000', 'm115@test.com', 'Instrucción', '5', 'Ciencias', 'Matemáticas', 'doc.pdf', TRUE),

(12, 'M1 Grupo 12', 301201, '0412000', 'm121@test.com', 'Proyectos', '1', 'Farmacia', 'Farmacia', 'doc.pdf', TRUE),
(12, 'M2 Grupo 12', 301202, '0412000', 'm122@test.com', 'Proyectos', '2', 'Farmacia', 'Farmacia', 'doc.pdf', TRUE),
(12, 'M3 Grupo 12', 301203, '0412000', 'm123@test.com', 'Investigación', '3', 'Farmacia', 'Farmacia', 'doc.pdf', TRUE),
(12, 'M4 Grupo 12', 301204, '0412000', 'm124@test.com', 'Sintesis', '4', 'Farmacia', 'Farmacia', 'doc.pdf', TRUE),
(12, 'M5 Grupo 12', 301205, '0412000', 'm125@test.com', 'Ensayos', '5', 'Farmacia', 'Farmacia', 'doc.pdf', TRUE),

(13, 'M1 Grupo 13', 301301, '0412000', 'm131@test.com', 'Proyección', '1', 'Medicina', 'Medicina', 'doc.pdf', TRUE),
(13, 'M2 Grupo 13', 301302, '0412000', 'm132@test.com', 'Foro', '2', 'Medicina', 'Medicina', 'doc.pdf', TRUE),
(13, 'M3 Grupo 13', 301303, '0412000', 'm133@test.com', 'Prensa', '3', 'Medicina', 'Medicina', 'doc.pdf', TRUE),
(13, 'M4 Grupo 13', 301304, '0412000', 'm134@test.com', 'Archivo', '4', 'Medicina', 'Medicina', 'doc.pdf', TRUE),
(13, 'M5 Grupo 13', 301305, '0412000', 'm135@test.com', 'Logística', '5', 'Medicina', 'Medicina', 'doc.pdf', TRUE),

(14, 'M1 Grupo 14', 301401, '0412000', 'm141@test.com', 'Actuación', '1', 'Ciencias', 'Física', 'doc.pdf', TRUE),
(14, 'M2 Grupo 14', 301402, '0412000', 'm142@test.com', 'Actuación', '2', 'Ciencias', 'Física', 'doc.pdf', TRUE),
(14, 'M3 Grupo 14', 301403, '0412000', 'm143@test.com', 'Escenografía', '3', 'Ciencias', 'Física', 'doc.pdf', TRUE),
(14, 'M4 Grupo 14', 301404, '0412000', 'm144@test.com', 'Dramaturgia', '4', 'Ciencias', 'Física', 'doc.pdf', TRUE),
(14, 'M5 Grupo 14', 301405, '0412000', 'm145@test.com', 'Iluminación', '5', 'Ciencias', 'Física', 'doc.pdf', TRUE),

(15, 'M1 Grupo 15', 301501, '0412000', 'm151@test.com', 'Talleres', '1', 'Ciencias', 'Matemáticas', 'doc.pdf', TRUE),
(15, 'M2 Grupo 15', 301502, '0412000', 'm152@test.com', 'Talleres', '2', 'Ciencias', 'Matemáticas', 'doc.pdf', TRUE),
(15, 'M3 Grupo 15', 301503, '0412000', 'm153@test.com', 'Diseño Lúdico', '3', 'Ciencias', 'Matemáticas', 'doc.pdf', TRUE),
(15, 'M4 Grupo 15', 301504, '0412000', 'm154@test.com', 'Material', '4', 'Ciencias', 'Matemáticas', 'doc.pdf', TRUE),
(15, 'M5 Grupo 15', 301505, '0412000', 'm155@test.com', 'Evaluación', '5', 'Ciencias', 'Matemáticas', 'doc.pdf', TRUE),

(16, 'M1 Grupo 16', 301601, '0412000', 'm161@test.com', 'Teoría', '1', 'Ciencias', 'Física', 'doc.pdf', TRUE),
(16, 'M2 Grupo 16', 301602, '0412000', 'm162@test.com', 'Experimentos', '2', 'Ciencias', 'Física', 'doc.pdf', TRUE),
(16, 'M3 Grupo 16', 301603, '0412000', 'm163@test.com', 'Seminarios', '3', 'Ciencias', 'Física', 'doc.pdf', TRUE),
(16, 'M4 Grupo 16', 301604, '0412000', 'm164@test.com', 'Redes', '4', 'Ciencias', 'Física', 'doc.pdf', TRUE),
(16, 'M5 Grupo 16', 301605, '0412000', 'm165@test.com', 'Olimpíadas', '5', 'Ciencias', 'Física', 'doc.pdf', TRUE),

(17, 'M1 Grupo 17', 301701, '0412000', 'm171@test.com', 'Canto', '1', 'Ciencias', 'Computación', 'doc.pdf', TRUE),
(17, 'M2 Grupo 17', 301702, '0412000', 'm172@test.com', 'Furruco', '2', 'Ciencias', 'Computación', 'doc.pdf', TRUE),
(17, 'M3 Grupo 17', 301703, '0412000', 'm173@test.com', 'Cuatro', '3', 'Ciencias', 'Computación', 'doc.pdf', TRUE),
(17, 'M4 Grupo 17', 301704, '0412000', 'm174@test.com', 'Tambora', '4', 'Ciencias', 'Computación', 'doc.pdf', TRUE),
(17, 'M5 Grupo 17', 301705, '0412000', 'm175@test.com', 'Charrasca', '5', 'Ciencias', 'Computación', 'doc.pdf', TRUE),

(18, 'M1 Grupo 18', 301801, '0412000', 'm181@test.com', 'Ritmos', '1', 'Ciencias', 'Biología', 'doc.pdf', TRUE),
(18, 'M2 Grupo 18', 301802, '0412000', 'm182@test.com', 'Ritmos', '2', 'Ciencias', 'Biología', 'doc.pdf', TRUE),
(18, 'M3 Grupo 18', 301803, '0412000', 'm183@test.com', 'Ensayos', '3', 'Ciencias', 'Biología', 'doc.pdf', TRUE),
(18, 'M4 Grupo 18', 301804, '0412000', 'm184@test.com', 'Montaje', '4', 'Ciencias', 'Biología', 'doc.pdf', TRUE),
(18, 'M5 Grupo 18', 301805, '0412000', 'm185@test.com', 'Instrumentos', '5', 'Ciencias', 'Biología', 'doc.pdf', TRUE),

(19, 'M1 Grupo 19', 301901, '0416000', 'm191@test.com', 'Huertos', '1', 'Agronomía', 'Agronomía', 'doc.pdf', TRUE),
(19, 'M2 Grupo 19', 301902, '0416000', 'm192@test.com', 'Riego', '2', 'Agronomía', 'Agronomía', 'doc.pdf', TRUE),
(19, 'M3 Grupo 19', 301903, '0416000', 'm193@test.com', 'Semillas', '3', 'Agronomía', 'Agronomía', 'doc.pdf', TRUE),
(19, 'M4 Grupo 19', 301904, '0416000', 'm194@test.com', 'Abono', '4', 'Agronomía', 'Agronomía', 'doc.pdf', TRUE),
(19, 'M5 Grupo 19', 301905, '0416000', 'm195@test.com', 'Cosecha', '5', 'Agronomía', 'Agronomía', 'doc.pdf', TRUE),

(20, 'M1 Grupo 20', 302001, '0416000', 'm201@test.com', 'Diseño', '1', 'Arquitectura y Urbanismo', 'Arquitectura', 'doc.pdf', TRUE),
(20, 'M2 Grupo 20', 302002, '0416000', 'm202@test.com', 'Diseño', '2', 'Arquitectura y Urbanismo', 'Arquitectura', 'doc.pdf', TRUE),
(20, 'M3 Grupo 20', 302003, '0416000', 'm203@test.com', 'Urbano', '3', 'Arquitectura y Urbanismo', 'Arquitectura', 'doc.pdf', TRUE),
(20, 'M4 Grupo 20', 302004, '0416000', 'm204@test.com', 'Maquetas', '4', 'Arquitectura y Urbanismo', 'Arquitectura', 'doc.pdf', TRUE),
(20, 'M5 Grupo 20', 302005, '0416000', 'm205@test.com', 'Comunidad', '5', 'Arquitectura y Urbanismo', 'Arquitectura', 'doc.pdf', TRUE),

(21, 'M1 Grupo 21', 302101, '0416000', 'm211@test.com', 'Finanzas', '1', 'Ciencias Económicas y Sociales', 'Economía', 'doc.pdf', TRUE),
(21, 'M2 Grupo 21', 302102, '0416000', 'm212@test.com', 'Mercadeo', '2', 'Ciencias Económicas y Sociales', 'Administración', 'doc.pdf', TRUE),
(21, 'M3 Grupo 21', 302103, '0416000', 'm213@test.com', 'Incubadora', '3', 'Ciencias Económicas y Sociales', 'Economía', 'doc.pdf', TRUE),
(21, 'M4 Grupo 21', 302104, '0416000', 'm214@test.com', 'Proyectos', '4', 'Ciencias Económicas y Sociales', 'Administración', 'doc.pdf', TRUE),
(21, 'M5 Grupo 21', 302105, '0416000', 'm215@test.com', 'Legal', '5', 'Ciencias Económicas y Sociales', 'Economía', 'doc.pdf', TRUE),

(22, 'M1 Grupo 22', 302201, '0416000', 'm221@test.com', 'Asesoría', '1', 'Ciencias Jurídicas y Políticas', 'Derecho', 'doc.pdf', TRUE),
(22, 'M2 Grupo 22', 302202, '0416000', 'm222@test.com', 'Litigio', '2', 'Ciencias Jurídicas y Políticas', 'Derecho', 'doc.pdf', TRUE),
(22, 'M3 Grupo 22', 302203, '0416000', 'm223@test.com', 'Derechos', '3', 'Ciencias Jurídicas y Políticas', 'Derecho', 'doc.pdf', TRUE),
(22, 'M4 Grupo 22', 302204, '0416000', 'm224@test.com', 'Atención', '4', 'Ciencias Jurídicas y Políticas', 'Derecho', 'doc.pdf', TRUE),
(22, 'M5 Grupo 22', 302205, '0416000', 'm225@test.com', 'Redacción', '5', 'Ciencias Jurídicas y Políticas', 'Derecho', 'doc.pdf', TRUE),

(23, 'M1 Grupo 23', 302301, '0416000', 'm231@test.com', 'Clínica', '1', 'Ciencias Veterinarias', 'Veterinaria', 'doc.pdf', TRUE),
(23, 'M2 Grupo 23', 302302, '0416000', 'm232@test.com', 'Cirugía', '2', 'Ciencias Veterinarias', 'Veterinaria', 'doc.pdf', TRUE),
(23, 'M3 Grupo 23', 302303, '0416000', 'm233@test.com', 'Vacunación', '3', 'Ciencias Veterinarias', 'Veterinaria', 'doc.pdf', TRUE),
(23, 'M4 Grupo 23', 302304, '0416000', 'm234@test.com', 'Campo', '4', 'Ciencias Veterinarias', 'Veterinaria', 'doc.pdf', TRUE),
(23, 'M5 Grupo 23', 302305, '0416000', 'm235@test.com', 'Insumos', '5', 'Ciencias Veterinarias', 'Veterinaria', 'doc.pdf', TRUE),

(24, 'M1 Grupo 24', 302401, '0416000', 'm241@test.com', 'Lectura', '1', 'Humanidades y Educación', 'Letras', 'doc.pdf', TRUE),
(24, 'M2 Grupo 24', 302402, '0416000', 'm242@test.com', 'Edición', '2', 'Humanidades y Educación', 'Letras', 'doc.pdf', TRUE),
(24, 'M3 Grupo 24', 302403, '0416000', 'm243@test.com', 'Escritura', '3', 'Humanidades y Educación', 'Letras', 'doc.pdf', TRUE),
(24, 'M4 Grupo 24', 302404, '0416000', 'm244@test.com', 'Eventos', '4', 'Humanidades y Educación', 'Letras', 'doc.pdf', TRUE),
(24, 'M5 Grupo 24', 302405, '0416000', 'm245@test.com', 'Difusión', '5', 'Humanidades y Educación', 'Letras', 'doc.pdf', TRUE),

(25, 'M1 Grupo 25', 302501, '0416000', 'm251@test.com', 'Hardware', '1', 'Ingeniería', 'Eléctrica', 'doc.pdf', TRUE),
(25, 'M2 Grupo 25', 302502, '0416000', 'm252@test.com', 'Software', '2', 'Ingeniería', 'Computación', 'doc.pdf', TRUE),
(25, 'M3 Grupo 25', 302503, '0416000', 'm253@test.com', 'Mecánica', '3', 'Ingeniería', 'Mecánica', 'doc.pdf', TRUE),
(25, 'M4 Grupo 25', 302504, '0416000', 'm254@test.com', 'Diseño 3D', '4', 'Ingeniería', 'Mecánica', 'doc.pdf', TRUE),
(25, 'M5 Grupo 25', 302505, '0416000', 'm255@test.com', 'Circuitos', '5', 'Ingeniería', 'Eléctrica', 'doc.pdf', TRUE),

(26, 'M1 Grupo 26', 302601, '0416000', 'm261@test.com', 'Prevención', '1', 'Odontología', 'Odontología', 'doc.pdf', TRUE),
(26, 'M2 Grupo 26', 302602, '0416000', 'm262@test.com', 'Clínica', '2', 'Odontología', 'Odontología', 'doc.pdf', TRUE),
(26, 'M3 Grupo 26', 302603, '0416000', 'm263@test.com', 'Educación', '3', 'Odontología', 'Odontología', 'doc.pdf', TRUE),
(26, 'M4 Grupo 26', 302604, '0416000', 'm264@test.com', 'Logística', '4', 'Odontología', 'Odontología', 'doc.pdf', TRUE),
(26, 'M5 Grupo 26', 302605, '0416000', 'm265@test.com', 'Materiales', '5', 'Odontología', 'Odontología', 'doc.pdf', TRUE),

(27, 'Carlos Mendoza', 27001001, '0412-9000001', 'carlos.mendoza@mail.com', 'Diseño Mecánico', '4', 'Ingeniería', 'Ingeniería Mecánica', 'V-27001001.pdf', TRUE),
(27, 'Mariana Silva', 27001002, '0414-9000002', 'mariana.silva@mail.com', 'Electrónica y Sensado', '5', 'Ingeniería', 'Ingeniería Eléctrica', 'V-27001002.pdf', TRUE),
(27, 'Luis Alarcón', 27001003, '0416-9000003', 'luis.alarcon@mail.com', 'Ensayos Biológicos', '3', 'Ciencias', 'Biología', 'V-27001003.pdf', TRUE),
(27, 'Valeria Morales', 27001004, '0412-9000004', 'valeria.morales@mail.com', 'Validación Clínica', '4', 'Medicina', 'Escuela Luis Razetti', 'V-27001004.pdf', TRUE),
(27, 'Gabriel Rivas', 27001005, '0414-9000005', 'gabriel.rivas@mail.com', 'Gestión Institucional', '2', 'DEU', 'Estudios Generales', 'V-27001005.pdf', TRUE);

-- ============================================================================
-- 5. INSERCIÓN DE LOGOS EN LA TABLA FILES PARA LOS 18 GRUPOS DE EXTENSIÓN
-- Relacionando dinámicamente owner_id con extension_groups (1 al 18)
-- y asignando el creador (uploaded_by) correspondiente al usuario de cada grupo.
-- ============================================================================

INSERT INTO files (owner_type, owner_id, file_key, purpose, version, public, metadata, uploaded_by) VALUES
('group', 1, 'groups/logos/LAMUN.webp', 'logo', 1, TRUE, '{"contentType": "image/webp"}'::jsonb, 15),
('group', 2, 'groups/logos/SPECTRUM.jpg', 'logo', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 16),
('group', 3, 'groups/logos/JAM_UCV.jpg', 'logo', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 17),
('group', 4, 'groups/logos/CECOBIO.jpg', 'logo', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 18),
('group', 5, 'groups/logos/CUASAR.jpg', 'logo', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 19),
('group', 6, 'groups/logos/CORAL.jpg', 'logo', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 20),
('group', 7, 'groups/logos/E.R.A..jpg', 'logo', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 21),
('group', 8, 'groups/logos/AGQ1.jpg', 'logo', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 22),
('group', 9, 'groups/logos/biosub.jpg', 'logo', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 23),
('group', 10, 'groups/logos/BAILACIENCIAS.jpg', 'logo', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 24),
('group', 11, 'groups/logos/UCVTANGO.jpg', 'logo', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 25),
('group', 12, 'groups/logos/PROCIENCIAS.jpg', 'logo', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 26),
('group', 13, 'groups/logos/CINECLUBCIENCIAS.jpg', 'logo', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 27),
('group', 14, 'groups/logos/tcb.png', 'logo', 1, TRUE, '{"contentType": "image/png"}'::jsonb, 28),
('group', 15, 'groups/logos/barrio-mat.png', 'logo', 1, TRUE, '{"contentType": "image/png"}'::jsonb, 29),
('group', 16, 'groups/logos/PHYSYS.jpg', 'logo', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 30),
('group', 17, 'groups/logos/conciencia-gaitera.png', 'logo', 1, TRUE, '{"contentType": "image/png"}'::jsonb, 31),
('group', 18, 'groups/logos/logo-ciencias-es-ritmo.jpg', 'logo', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 32),
('group', 19, 'groups/logos/AGROVERDE.jpg', 'logo', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 33),
('group', 20, 'groups/logos/FAUDISENO.jpg', 'logo', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 34),
('group', 21, 'groups/logos/FACESEMPRENDE.jpg', 'logo', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 35),
('group', 22, 'groups/logos/CJURIDICA.jpg', 'logo', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 36),
('group', 23, 'groups/logos/VETACTION.jpg', 'logo', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 37),
('group', 24, 'groups/logos/LETRASVIVAS.jpg', 'logo', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 38),
('group', 25, 'groups/logos/ROBOTICA.jpg', 'logo', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 39),
('group', 26, 'groups/logos/SONRISAS.jpg', 'logo', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 40),
('group', 27, 'groups/logos/BIOINGENIERIA.jpg', 'logo', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 43);

-- ============================================================================
-- 6. INSERCIÓN DE ACTIVIDADES (Mapeo estricto del Mock al esquema relacional)
-- Asignando los IDs correctos de deu.extension_groups
-- ============================================================================
INSERT INTO deu.activities (id, name, group_id, description, date_start, date_end, knowledge_area, allies, stimated_participants, actual_participants, group_participants, gallery_url, report_checked, location, is_featured) VALUES
(1, 'Modelo de las Naciones Unidas - 2025 - Caracas', 1, 'Simulación diplomática centralizada.', '2025-11-28', '2025-11-30', ARRAY['Comunicación y Retórica', 'Formación']::VARCHAR[], '', 50, 25, 10, 'https://drive.google.com/drive/u/0/folders/1ytxcRHCp36Osing0CuQXnkVg_NtuRAky', true, 'Venezuela, Distrito Capital, Caracas, A', true),
(2, 'Actividad de Divulgación', 2, 'Exposición científica en plaza central.', '2025-03-10', '2025-03-10', ARRAY['Difusión del Conocimiento', 'Innovación']::VARCHAR[], '', 70, 20, 8, 'https://drive.google.com/drive/u/0/folders/1ytxcRHCp36Osing0CuQXnkVg_NtuRAky', true, 'Venezuela, Distrito Capital, Libertador, Facultad de Ciencias', true),
(3, 'Concierto de Ritmos', 3, 'Intercambio acústico libre.', '2026-07-01', '2026-10-01', ARRAY['Recreación', 'ARTE_DESCONOCIDO']::VARCHAR[], '', NULL, NULL, NULL, NULL, true, 'Venezuela, Distrito Capital, Libertador, Pasillos de Ciencias', true),
(4, 'Muestra de Especies', 4, 'Exhibición controlada de biodiversidad.', '2026-04-22', '2026-04-22', ARRAY['Ambiental', 'Difusión del Conocimiento']::VARCHAR[], '', NULL, NULL, NULL, NULL, true, 'Venezuela, Distrito Capital, Libertador, Laboratorio de Biología', true),
(5, 'Observación Lunar', 5, 'Taller abierto nocturno astronómico.', '2026-03-12', '2026-03-12', ARRAY['Difusión del Conocimiento']::VARCHAR[], '', 10, 20, 3, 'https://drive.google.com/drive/u/0/folders/1ytxcRHCp36Osing0CuQXnkVg_NtuRAky', true, 'Venezuela, Distrito Capital, Libertador, Observatorio', true),
(6, 'Gala Coral', 6, 'Presentación del repertorio litúrgico y popular.', '2025-04-22', '2025-04-22', ARRAY['Recreación']::VARCHAR[], '', 20, 5, 6, 'https://drive.google.com/drive/u/0/folders/1ytxcRHCp36Osing0CuQXnkVg_NtuRAky', true, 'Venezuela, Distrito Capital, Libertador, Auditorio de Ciencias', true),
(7, 'Jornada de Reforestación', 7, 'Siembra de plantas autóctonas en áreas verdes.', '2026-01-22', '2026-01-22', ARRAY['Ambiental', 'Acción Social']::VARCHAR[], '', 40, 35, 12, 'https://drive.google.com/drive/u/0/folders/1ytxcRHCp36Osing0CuQXnkVg_NtuRAky', true, 'Venezuela, Distrito Capital, Libertador, Áreas Verdes Ciencias', true),
(8, 'Taller de Oratoria', 1, 'Técnicas de expresión oral aplicadas a debates.', '2025-04-20', '2025-04-22', ARRAY['Comunicación y Retórica', 'Acompañamiento y Asesoría Estudiantil']::VARCHAR[], '', 11, 20, 4, 'https://drive.google.com/drive/u/0/folders/1ytxcRHCp36Osing0CuQXnkVg_NtuRAky', true, 'Venezuela, Distrito Capital, Libertador, Sala de Extensión', false),
(9, 'Simposio de Física Espacial', 2, 'Ciclo de micro-charlas para estudiantes.', '2027-04-22', '2027-04-22', ARRAY['Difusión del Conocimiento', 'Formación']::VARCHAR[], '', NULL, NULL, NULL, NULL, false, 'Venezuela, Distrito Capital, Libertador, Auditorio Ciencias', false),
(10, 'Recolección de Desechos', 7, 'Limpieza profunda de áreas comunes.', '2027-12-10', '2027-12-10', ARRAY['Ambiental']::VARCHAR[], '', NULL, NULL, NULL, NULL, false, 'Venezuela, Distrito Capital, Libertador, Campus UCV', false),
(11, 'Charla Astro-Física', 5, 'Introducción a la cosmología moderna.', '2025-12-22', '2025-12-22', ARRAY['Difusión del Conocimiento']::VARCHAR[], '', 30, 28, 5, 'https://drive.google.com/drive/u/0/folders/1ytxcRHCp36Osing0CuQXnkVg_NtuRAky', true, 'Venezuela, Distrito Capital, Libertador, Escuela de Física', false),
(12, 'Conferencia Geopolítica', 1, 'Análisis de conflictos globales.', '2026-12-20', '2026-12-22', ARRAY['Comunicación y Retórica', 'AREA_NO_VALIDA']::VARCHAR[], '', NULL, NULL, NULL, NULL, true, 'Venezuela, Distrito Capital, Libertador, Sala Central', true),
(13, 'Simulación Interna', 1, 'Sesión de entrenamiento de delegados.', '2025-12-01', '2025-12-03', ARRAY['Formación', 'Salud y Bienestar']::VARCHAR[], '', 15, 20, 6, 'https://drive.google.com/drive/u/0/folders/1ytxcRHCp36Osing0CuQXnkVg_NtuRAky', true, 'Venezuela, Distrito Capital, Libertador, Aula de Clases', false),
(14, 'Mesa Técnica Suelos', 8, 'Presentación de resultados analíticos de campo.', '2025-12-03', '2025-12-03', ARRAY['Innovación', 'Formación']::VARCHAR[], '', 25, 22, 5, 'https://drive.google.com/drive/u/0/folders/1ytxcRHCp36Osing0CuQXnkVg_NtuRAky', true, 'Venezuela, Distrito Capital, Libertador, Instituto Ciencias de la Tierra', true);

-- Sincronizar el ID secuencial de la tabla de actividades
SELECT setval('deu.activities_id_seq', 14, true);

-- ============================================================================
-- 7. INSERCIÓN DE IMÁGENES DE ACTIVIDADES EN LA TABLA FILES
-- owner_type: 'group_activity'
-- owner_id: Hace referencia directa al ID de la actividad (1 al 14)
-- uploaded_by: Relacionado al ID del usuario dueño del grupo de la actividad
-- ============================================================================

INSERT INTO files (owner_type, owner_id, file_key, purpose, version, public, metadata, uploaded_by) VALUES
('group_activity', 1, 'activities/background-1.jpg', 'cubierta', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 15),
('group_activity', 2, 'activities/background-1.jpg', 'cubierta', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 16),
('group_activity', 3, 'activities/background-1.jpg', 'cubierta', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 17),
('group_activity', 4, 'activities/background-1.jpg', 'cubierta', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 18),
('group_activity', 5, 'activities/background-1.jpg', 'cubierta', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 19),
('group_activity', 6, 'activities/background-1.jpg', 'cubierta', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 20),
('group_activity', 7, 'activities/background-1.jpg', 'cubierta', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 21),
('group_activity', 8, 'activities/background-1.jpg', 'cubierta', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 15),
('group_activity', 9, 'activities/background-1.jpg', 'cubierta', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 16),
('group_activity', 10, 'activities/background-1.jpg', 'cubierta', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 21),
('group_activity', 11, 'activities/background-1.jpg', 'cubierta', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 19),
('group_activity', 12, 'activities/background-2.jpg', 'cubierta', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 15),
('group_activity', 13, 'activities/background-1.jpg', 'cubierta', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 15),
('group_activity', 14, 'activities/background-1.jpg', 'cubierta', 1, TRUE, '{"contentType": "image/jpeg"}'::jsonb, 22);

