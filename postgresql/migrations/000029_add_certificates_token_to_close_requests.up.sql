-- Token aleatorio del enlace de descarga del ZIP con todos los certificados de la solicitud
ALTER TABLE deu.course_cycle_close_requests ADD COLUMN IF NOT EXISTS certificates_token TEXT UNIQUE;
