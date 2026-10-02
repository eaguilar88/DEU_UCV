DROP INDEX IF EXISTS deu.idx_files_file_key;

UPDATE deu.files
SET public = false
WHERE (owner_type = 'course' AND purpose = 'portada')
   OR (owner_type = 'group_activity' AND purpose = 'cubierta')
   OR (owner_type IN ('group', 'provider') AND purpose = 'logo');
