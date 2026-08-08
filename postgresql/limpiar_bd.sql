-- Forzar la eliminación del esquema y todo lo que tiene adentro (tablas, tipos, etc.)
DROP SCHEMA IF EXISTS deu CASCADE;

-- Volver a crear el esquema limpio
CREATE SCHEMA deu;

-- Asegurar los permisos para tu usuario administrador
ALTER SCHEMA deu OWNER TO deu_admin;
GRANT ALL PRIVILEGES ON SCHEMA deu TO deu_admin;